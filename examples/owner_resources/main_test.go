package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suranig/refine-gin/pkg/handler"
	"github.com/suranig/refine-gin/pkg/middleware"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

// Setup test environment
func TestMain(m *testing.M) {
	// Switch to test mode
	gin.SetMode(gin.TestMode)

	// Run tests
	code := m.Run()

	// Exit
	os.Exit(code)
}

// Seed database with test data
func seedTestData(db *gorm.DB) {
	// Seed users
	users := []User{
		{ID: "user-1", Name: "Test User 1", Email: "user1@test.com", Role: "admin"},
		{ID: "user-2", Name: "Test User 2", Email: "user2@test.com", Role: "user"},
	}
	for _, user := range users {
		db.Create(&user)
	}

	// Seed notes
	notes := []Note{
		{ID: "note-1", Title: "User 1 Note 1", Content: "Test content 1", OwnerID: "user-1"},
		{ID: "note-2", Title: "User 1 Note 2", Content: "Test content 2", OwnerID: "user-1"},
		{ID: "note-3", Title: "User 2 Note", Content: "Test content 3", OwnerID: "user-2"},
	}
	for _, note := range notes {
		db.Create(&note)
	}
}

// Setup test router with fresh in-memory database
func setupTestRouter() (*gin.Engine, *gorm.DB) {
	// Create a test database
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	// Auto migrate test models
	db.AutoMigrate(&User{}, &Task{}, &Note{})

	// Seed test data
	seedTestData(db)

	r := gin.New()

	// Create resources
	userResource := resource.NewResource(resource.ResourceConfig{
		Name:  "users",
		Model: User{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationRead,
			resource.OperationCreate,
			resource.OperationUpdate,
			resource.OperationDelete,
		},
	})

	noteResource := resource.NewResource(resource.ResourceConfig{
		Name:  "notes",
		Model: Note{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationRead,
			resource.OperationCreate,
			resource.OperationUpdate,
			resource.OperationDelete,
			resource.OperationCreateMany,
			resource.OperationUpdateMany,
			resource.OperationDeleteMany,
		},
	})

	// Create owner resource
	ownerNoteResource := resource.NewOwnerResource(noteResource, resource.OwnerConfig{
		OwnerField:       "OwnerID",
		EnforceOwnership: true,
	})

	// Create repositories
	userRepo := repository.NewGenericRepositoryWithResource(db, userResource)
	noteRepo, err := repository.NewOwnerRepository(db, ownerNoteResource)
	if err != nil {
		panic("Failed to create owner repository: " + err.Error())
	}

	// API group
	api := r.Group("/api")

	// Register standard resources
	handler.RegisterResource(api, userResource, userRepo)

	// Register secured endpoints with direct owner ID middleware
	securedApi := api.Group("")
	securedApi.Use(func(c *gin.Context) {
		// Get the user ID directly from header for testing
		userID := c.Request.Header.Get("X-User-ID")
		if userID == "" {
			userID = "user-1" // Default to user-1 if not specified
		}

		// Print debug info
		fmt.Printf("Request: %s %s, User ID: %s\n", c.Request.Method, c.Request.URL.Path, userID)

		// Explicitly check the types for debugging
		fmt.Printf("User ID type: %T, value: %v\n", userID, userID)

		// Set it directly in the context with the expected key
		c.Set(middleware.OwnerContextKey, userID)

		// Also set in the request context to ensure it's available
		ctx := context.WithValue(c.Request.Context(), middleware.OwnerContextKey, userID)
		c.Request = c.Request.WithContext(ctx)

		// Debug: check if we can retrieve the value back
		storedVal, exists := c.Get(middleware.OwnerContextKey)
		fmt.Printf("Value stored in gin context: exists=%v, value=%v, type=%T\n",
			exists, storedVal, storedVal)

		fmt.Printf("Setting owner ID directly in context: %s\n", userID)

		c.Next()
	})
	handler.RegisterOwnerResource(securedApi, ownerNoteResource, noteRepo)

	return r, db
}

// Test JWT middleware for testing (no longer used - keeping for reference)
func testJWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user ID from header for testing
		userID := c.Request.Header.Get("X-User-ID")
		if userID == "" {
			userID = "user-1" // Default to user-1 if not specified
		}

		// Create a minimal JWT token
		rawToken := jwt.New(jwt.SigningMethodHS256)

		// Set claims
		claims := rawToken.Claims.(jwt.MapClaims)
		claims["sub"] = userID
		claims["exp"] = 1900000000 // Some future time

		// Store token in the context
		c.Set("token", rawToken)

		// Also set the raw claims for compatibility
		c.Set("claims", claims)

		// Debug log
		fmt.Printf("Setting user ID in JWT: %s\n", userID)

		c.Next()
	}
}

// Test listing notes (should only see notes owned by the user)
func TestListNotes(t *testing.T) {
	t.Run("User1 can only see their own notes", func(t *testing.T) {
		// Create a new router with fresh database
		router, _ := setupTestRouter()

		// Create request
		req, _ := http.NewRequest("GET", "/api/notes", nil)
		req.Header.Set("X-User-ID", "user-1")
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Log the response for debugging
		t.Logf("Response body: %s", w.Body.String())

		// Should have 2 notes for user-1
		require.NotNil(t, response["total"], "Total field is missing in response")
		assert.Equal(t, float64(2), response["total"])

		// Verify the notes are for user-1
		require.NotNil(t, response["data"], "Data field is missing in response")
		data, ok := response["data"].([]interface{})
		require.True(t, ok, "Data field is not an array")
		require.NotEmpty(t, data, "Data array is empty")

		for _, item := range data {
			note, ok := item.(map[string]interface{})
			require.True(t, ok, "Note item is not an object")
			assert.Equal(t, "user-1", note["ownerId"])
		}
	})

	t.Run("User2 can only see their own notes", func(t *testing.T) {
		// Create a new router with fresh database
		router, _ := setupTestRouter()

		// Create request
		req, _ := http.NewRequest("GET", "/api/notes", nil)
		req.Header.Set("X-User-ID", "user-2")
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Log the response for debugging
		t.Logf("Response body: %s", w.Body.String())

		// Check response code only
		assert.Equal(t, http.StatusOK, w.Code, "API should return 200 OK")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "Response should be valid JSON")

		// Verify we have data and total fields in the response
		require.NotNil(t, response["data"], "Data field should be present in response")
		require.NotNil(t, response["total"], "Total field should be present in response")

		// Should have 1 note for user-2
		assert.Equal(t, float64(1), response["total"], "User-2 should see 1 note")

		// Verify the notes are for user-2
		data, ok := response["data"].([]interface{})
		require.True(t, ok, "Data field is not an array")
		require.NotEmpty(t, data, "Data array is empty")

		for _, item := range data {
			note, ok := item.(map[string]interface{})
			require.True(t, ok, "Note item is not an object")
			assert.Equal(t, "user-2", note["ownerId"], "The note should belong to user-2")
		}
	})
}

// Test getting a specific note (should enforce ownership)
func TestGetNote(t *testing.T) {
	t.Run("User can access their own note", func(t *testing.T) {
		// Create a new router with fresh database
		router, testDB := setupTestRouter()

		// Create a new note for user-1
		note := Note{
			ID:      "test-get-note-1",
			Title:   "Test Note for Get",
			Content: "Test content for get",
			OwnerID: "user-1",
		}
		testDB.Create(&note)

		// Create request
		req, _ := http.NewRequest("GET", "/api/notes/test-get-note-1", nil)
		req.Header.Set("X-User-ID", "user-1")
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Log the response for debugging
		t.Logf("Response body: %s", w.Body.String())

		require.NotNil(t, response["data"], "Data field is missing in response")
		data, ok := response["data"].(map[string]interface{})
		require.True(t, ok, "Data field is not an object")
		assert.Equal(t, "test-get-note-1", data["id"])
		assert.Equal(t, "user-1", data["ownerId"])
	})

	t.Run("User cannot access another user's note", func(t *testing.T) {
		// Create a new router with fresh database
		router, testDB := setupTestRouter()

		// Create a new note for user-2
		note := Note{
			ID:      "test-get-note-2",
			Title:   "Test Note for Get (User 2)",
			Content: "Test content for user 2",
			OwnerID: "user-2",
		}
		testDB.Create(&note)

		// Create request as user-1
		req, _ := http.NewRequest("GET", "/api/notes/test-get-note-2", nil)
		req.Header.Set("X-User-ID", "user-1")
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Check response - should be forbidden
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// Test creating a note (should set owner ID automatically)
func TestCreateNote(t *testing.T) {
	t.Run("Creating note sets owner automatically", func(t *testing.T) {
		// Create a new router with fresh database
		router, testDB := setupTestRouter()

		// Create request body
		noteData := map[string]interface{}{
			"id":      "note-new",
			"title":   "New Test Note",
			"content": "This note should be owned by user-1",
		}
		jsonData, _ := json.Marshal(noteData)

		// Create request
		req, _ := http.NewRequest("POST", "/api/notes", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-1")
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Log the response for debugging
		t.Logf("Response body: %s", w.Body.String())

		require.NotNil(t, response["data"], "Data field is missing in response")
		data, ok := response["data"].(map[string]interface{})
		require.True(t, ok, "Data field is not an object")
		assert.Equal(t, "note-new", data["id"])
		assert.Equal(t, "user-1", data["ownerId"])

		// Verify in database
		var note Note
		result := testDB.First(&note, "id = ?", "note-new")
		assert.NoError(t, result.Error)
		assert.Equal(t, "user-1", note.OwnerID)
	})
}

// Test updating a note (should enforce ownership)
func TestUpdateNote(t *testing.T) {
	// Setup
	r, db := setupTestRouter()

	// Reset database before each test
	db.Exec("DELETE FROM notes")

	t.Run("User_can_update_their_own_note", func(t *testing.T) {
		// Create a note owned by user-1
		note := Note{
			ID:      "test-update-note-1",
			Title:   "Original Title",
			Content: "Original content",
			OwnerID: "user-1",
		}
		// Direct database insert to guarantee the owner_id is set
		result := db.Create(&note)
		require.NoError(t, result.Error)

		// Verify the note was created correctly in the database
		var checkNote Note
		db.Where("id = ?", "test-update-note-1").First(&checkNote)
		t.Logf("Created note in DB: ID=%s, Title=%s, OwnerID=%s",
			checkNote.ID, checkNote.Title, checkNote.OwnerID)
		require.Equal(t, "user-1", checkNote.OwnerID)

		// Check the raw database value to ensure OwnerID is saved correctly
		var rawOwnerID string
		db.Raw("SELECT owner_id FROM notes WHERE id = ?", "test-update-note-1").Scan(&rawOwnerID)
		t.Logf("Raw database owner_id: %s", rawOwnerID)
		require.Equal(t, "user-1", rawOwnerID)

		// Check owner relationship directly before update
		var ownerCheckResult struct {
			ID      string
			OwnerID string
		}
		ownerCheckErr := db.Raw("SELECT id, owner_id FROM notes WHERE id = ? AND owner_id = ?",
			"test-update-note-1", "user-1").Scan(&ownerCheckResult).Error
		t.Logf("Before update - owner check: %v, ID=%s, OwnerID=%s, Error=%v",
			ownerCheckResult.ID != "", ownerCheckResult.ID, ownerCheckResult.OwnerID, ownerCheckErr)

		// Make update request as the owner
		w := httptest.NewRecorder()
		updateData := `{"title": "Updated Title", "content": "Updated content"}`
		req, _ := http.NewRequest("PUT", "/api/notes/test-update-note-1", strings.NewReader(updateData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-1") // Set the X-User-ID header

		// Process the request
		r.ServeHTTP(w, req)

		// Directly check the database after the update to see what happened
		var updatedNoteCheck Note
		findResult := db.Where("id = ?", "test-update-note-1").First(&updatedNoteCheck)
		t.Logf("After update - DB check: Error=%v, Note=%+v",
			findResult.Error, updatedNoteCheck)

		// Check raw SQL as well
		var rawCheckResult struct {
			ID      string
			Title   string
			Content string
			OwnerID string `gorm:"column:owner_id"`
		}
		rawSqlErr := db.Raw("SELECT id, title, content, owner_id FROM notes WHERE id = ?",
			"test-update-note-1").Scan(&rawCheckResult).Error
		t.Logf("After update - Raw SQL check: Error=%v, Result=%+v",
			rawSqlErr, rawCheckResult)

		// Assert response
		t.Logf("Update response: %d - %s", w.Code, w.Body.String())
		require.Equal(t, http.StatusOK, w.Code)

		// Verify the note was updated
		var updatedNote Note
		db.Where("id = ?", "test-update-note-1").First(&updatedNote)
		require.Equal(t, "Updated Title", updatedNote.Title)
		require.Equal(t, "user-1", updatedNote.OwnerID)

		// Check raw database values after update
		var afterUpdateResult struct {
			ID      string
			Title   string
			OwnerID string
		}
		db.Raw("SELECT id, title, owner_id FROM notes WHERE id = ?", "test-update-note-1").Scan(&afterUpdateResult)
		t.Logf("After update - raw DB values: ID=%s, Title=%s, OwnerID=%s",
			afterUpdateResult.ID, afterUpdateResult.Title, afterUpdateResult.OwnerID)
	})

	t.Run("User_cannot_update_another_user's_note", func(t *testing.T) {
		// Create a note owned by user-2
		note := Note{
			ID:      "test-update-note-2",
			Title:   "User 2 Note",
			Content: "This belongs to user 2",
			OwnerID: "user-2",
		}
		db.Create(&note)

		// Verify the note was created correctly with owner_id
		var checkNote Note
		db.Where("id = ?", "test-update-note-2").First(&checkNote)
		require.Equal(t, "user-2", checkNote.OwnerID)

		// Make update request as user-1 (not the owner)
		w := httptest.NewRecorder()
		updateData := `{"title": "Should Not Update", "content": "Should not update content"}`
		req, _ := http.NewRequest("PUT", "/api/notes/test-update-note-2", strings.NewReader(updateData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-1") // Set the X-User-ID header

		// Process the request
		r.ServeHTTP(w, req)

		// Check response - should be forbidden
		require.Equal(t, http.StatusForbidden, w.Code)

		// Verify in database that note is unchanged
		var updatedNote Note
		db.Where("id = ?", "test-update-note-2").First(&updatedNote)
		require.Equal(t, "User 2 Note", updatedNote.Title)
		require.Equal(t, "user-2", updatedNote.OwnerID)
	})
}

// Test deleting a note (should enforce ownership)
func TestDeleteNote(t *testing.T) {
	t.Run("User can delete their own note", func(t *testing.T) {
		// Create a new router with fresh database
		router, testDB := setupTestRouter()

		// Create a new note for user-1
		note := Note{
			ID:      "test-delete-note-1",
			Title:   "Note to Delete",
			Content: "This will be deleted",
			OwnerID: "user-1",
		}
		testDB.Create(&note)

		// Create request
		req, _ := http.NewRequest("DELETE", "/api/notes/test-delete-note-1", nil)
		req.Header.Set("X-User-ID", "user-1")
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify in database
		var note1 Note
		result := testDB.First(&note1, "id = ?", "test-delete-note-1")
		assert.Error(t, result.Error) // Should not find the note
	})

	t.Run("User cannot delete another user's note", func(t *testing.T) {
		// Create a new router with fresh database
		router, testDB := setupTestRouter()

		// Create a new note for user-2
		note := Note{
			ID:      "test-delete-note-2",
			Title:   "User 2 Note for Delete Test",
			Content: "This should not be deleted by user 1",
			OwnerID: "user-2",
		}
		testDB.Create(&note)

		// Create request as user-1
		req, _ := http.NewRequest("DELETE", "/api/notes/test-delete-note-2", nil)
		req.Header.Set("X-User-ID", "user-1")
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Check response - should be forbidden
		assert.Equal(t, http.StatusForbidden, w.Code)

		// Verify in database note still exists
		var note2 Note
		result := testDB.First(&note2, "id = ?", "test-delete-note-2")
		assert.NoError(t, result.Error) // Should still find the note
		assert.Equal(t, "user-2", note2.OwnerID)
	})
}

// Test the owner repository directly
func TestOwnerRepository(t *testing.T) {
	// Create a test database
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate test models
	db.AutoMigrate(&Note{})

	// Create a note
	note := Note{
		ID:      "direct-test-note",
		Title:   "Test Direct Repository",
		Content: "Testing repository directly",
		OwnerID: "user-1",
	}
	result := db.Create(&note)
	require.NoError(t, result.Error)

	// Verify the note was created
	var createdNote Note
	db.First(&createdNote, "id = ?", "direct-test-note")
	require.Equal(t, "user-1", createdNote.OwnerID)

	// Create the resource
	noteResource := resource.NewResource(resource.ResourceConfig{
		Name:  "notes",
		Model: Note{},
	})

	// Create owner resource
	ownerNoteResource := resource.NewOwnerResource(noteResource, resource.OwnerConfig{
		OwnerField:       "OwnerID",
		EnforceOwnership: true,
	})

	// Create owner repository
	noteRepo, err := repository.NewOwnerRepository(db, ownerNoteResource)
	require.NoError(t, err)

	// Create a context with owner ID
	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.OwnerContextKey, "user-1")

	// Try to get the note
	retrievedNote, err := noteRepo.Get(ctx, "direct-test-note")
	t.Logf("Get result: %v, Error: %v", retrievedNote, err)
	require.NoError(t, err)
	require.NotNil(t, retrievedNote)

	// Try to update the note
	updateData := map[string]interface{}{
		"title": "Updated Direct Test",
	}
	updatedNote, err := noteRepo.Update(ctx, "direct-test-note", updateData)
	t.Logf("Update result: %v, Error: %v", updatedNote, err)
	require.NoError(t, err)
	require.NotNil(t, updatedNote)
}

// Test the database schema
func TestDatabaseSchema(t *testing.T) {
	// Create a test database
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate test models
	db.AutoMigrate(&Note{})

	// Create a note
	note := Note{
		ID:      "schema-test-note",
		Title:   "Schema Test",
		Content: "Testing database schema",
		OwnerID: "user-1",
	}
	result := db.Create(&note)
	require.NoError(t, result.Error)

	// Get table info
	var columns []struct {
		CID       int
		Name      string
		Type      string
		NotNull   int
		DfltValue interface{}
		PK        int
	}

	db.Raw("PRAGMA table_info(notes)").Scan(&columns)
	t.Log("Table schema for 'notes':")
	for _, col := range columns {
		t.Logf("Column: %s, Type: %s, PK: %d", col.Name, col.Type, col.PK)
	}

	// Test direct SQL query
	var directQueryResult []struct {
		ID      string
		Title   string
		OwnerID string `gorm:"column:owner_id"`
	}

	db.Raw("SELECT id, title, owner_id FROM notes WHERE id = ? AND owner_id = ?",
		"schema-test-note", "user-1").Scan(&directQueryResult)

	t.Logf("Direct SQL query result: %v", directQueryResult)

	// Test with quoted column name
	var quotedQueryResult []struct {
		ID      string
		Title   string
		OwnerID string `gorm:"column:owner_id"`
	}

	db.Raw("SELECT id, title, \"owner_id\" FROM notes WHERE id = ? AND \"owner_id\" = ?",
		"schema-test-note", "user-1").Scan(&quotedQueryResult)

	t.Logf("Quoted SQL query result: %v", quotedQueryResult)
}

// Test the Note schema and OwnerResource configuration
func TestNoteSchemaAndOwnerResource(t *testing.T) {
	// Create a test database
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate the Note model
	db.AutoMigrate(&Note{})

	// Create a note
	note := Note{
		ID:      "schema-test-note",
		Title:   "Schema Test",
		Content: "Testing owner resource",
		OwnerID: "user-1",
	}
	result := db.Create(&note)
	require.NoError(t, result.Error)

	// Create note resource
	noteResource := resource.NewResource(resource.ResourceConfig{
		Name:  "notes",
		Model: Note{},
	})

	// Print the Note struct fields
	t.Log("Note struct fields:")
	noteType := reflect.TypeOf(Note{})
	for i := 0; i < noteType.NumField(); i++ {
		field := noteType.Field(i)
		t.Logf("Field: %s, JSON: %s, GORM: %s",
			field.Name,
			field.Tag.Get("json"),
			field.Tag.Get("gorm"))
	}

	t.Run("Using OwnerID", func(t *testing.T) {
		// Create owner resource with the specified field
		ownerResource := resource.NewOwnerResource(noteResource, resource.OwnerConfig{
			OwnerField:       "OwnerID",
			EnforceOwnership: true,
		})

		// Check the owner field
		t.Logf("Owner field configured as: %s", ownerResource.GetOwnerField())

		// Create owner repository
		repo, err := repository.NewOwnerRepository(db, ownerResource)
		require.NoError(t, err)

		// Create context with owner ID
		ctx := context.Background()
		ctx = context.WithValue(ctx, middleware.OwnerContextKey, "user-1")

		// Try to get the note
		found, err := repo.Get(ctx, "schema-test-note")
		t.Logf("Get result for %s: %v, err: %v", "OwnerID", found, err)

		// Try direct SQL query with owner field
		var directQueryResult []struct {
			ID      string
			OwnerID string `gorm:"column:owner_id"`
		}

		// Use the configured owner field to build the query
		columnName := db.Config.NamingStrategy.ColumnName("", "OwnerID")
		t.Logf("Column name for %s: %s", "OwnerID", columnName)

		db.Raw("SELECT id, owner_id FROM notes WHERE id = ? AND "+columnName+" = ?",
			"schema-test-note", "user-1").Scan(&directQueryResult)

		t.Logf("Direct SQL query result for %s: %v", "OwnerID", directQueryResult)
	})
}
