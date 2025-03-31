package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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

		// Set it directly in the context with the expected key
		c.Set(middleware.OwnerContextKey, userID)

		// Also set in the request context to ensure it's available
		ctx := context.WithValue(c.Request.Context(), middleware.OwnerContextKey, userID)
		c.Request = c.Request.WithContext(ctx)

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
	t.Run("User can update their own note", func(t *testing.T) {
		// Create a new router with fresh database
		router, testDB := setupTestRouter()

		// Create a new note for user-1
		note := Note{
			ID:      "test-update-note-1",
			Title:   "Original Title",
			Content: "Original content",
			OwnerID: "user-1",
		}
		testDB.Create(&note)

		// Create request body
		updateData := map[string]interface{}{
			"title":   "Updated Title",
			"content": "Updated content",
		}
		jsonData, _ := json.Marshal(updateData)

		// Create request
		req, _ := http.NewRequest("PUT", "/api/notes/test-update-note-1", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-1")
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify in database
		var updatedNote Note
		testDB.First(&updatedNote, "id = ?", "test-update-note-1")
		assert.Equal(t, "Updated Title", updatedNote.Title)
		assert.Equal(t, "Updated content", updatedNote.Content)
		assert.Equal(t, "user-1", updatedNote.OwnerID) // Owner shouldn't change
	})

	t.Run("User cannot update another user's note", func(t *testing.T) {
		// Create a new router with fresh database
		router, testDB := setupTestRouter()

		// Create a new note for user-2
		note := Note{
			ID:      "test-update-note-2",
			Title:   "User 2 Note",
			Content: "This belongs to user 2",
			OwnerID: "user-2",
		}
		testDB.Create(&note)

		// Create request body
		updateData := map[string]interface{}{
			"title":   "Should Not Update",
			"content": "Should not update content",
		}
		jsonData, _ := json.Marshal(updateData)

		// Create request as user-1
		req, _ := http.NewRequest("PUT", "/api/notes/test-update-note-2", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-1")
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Check response - should be forbidden
		assert.Equal(t, http.StatusForbidden, w.Code)

		// Verify in database note is unchanged
		var note2 Note
		testDB.First(&note2, "id = ?", "test-update-note-2")
		assert.NotEqual(t, "Should Not Update", note2.Title)
		assert.Equal(t, "user-2", note2.OwnerID)
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
