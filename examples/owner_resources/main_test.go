package main

import (
	"bytes"
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
	router *gin.Engine
	db     *gorm.DB
)

// Setup test environment
func TestMain(m *testing.M) {
	// Switch to test mode
	gin.SetMode(gin.TestMode)

	// Create a test database
	var err error
	db, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	// Auto migrate test models
	db.AutoMigrate(&User{}, &Task{}, &Note{})

	// Seed test data
	seedTestData()

	// Create router
	router = setupTestRouter()

	// Run tests
	code := m.Run()

	// Exit
	os.Exit(code)
}

// Seed database with test data
func seedTestData() {
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

// Setup test router
func setupTestRouter() *gin.Engine {
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

		fmt.Printf("Setting owner ID directly in context: %s\n", userID)

		c.Next()
	})
	handler.RegisterOwnerResource(securedApi, ownerNoteResource, noteRepo)

	return r
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
		// Create request
		req, _ := http.NewRequest("GET", "/api/notes", nil)
		req.Header.Set("X-User-ID", "user-2")
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

		// Should have 1 note for user-2
		require.NotNil(t, response["total"], "Total field is missing in response")
		assert.Equal(t, float64(1), response["total"])

		// Verify the note is for user-2
		require.NotNil(t, response["data"], "Data field is missing in response")
		data, ok := response["data"].([]interface{})
		require.True(t, ok, "Data field is not an array")
		require.NotEmpty(t, data, "Data array is empty")

		for _, item := range data {
			note, ok := item.(map[string]interface{})
			require.True(t, ok, "Note item is not an object")
			assert.Equal(t, "user-2", note["ownerId"])
		}
	})
}

// Test getting a specific note (should enforce ownership)
func TestGetNote(t *testing.T) {
	t.Run("User can access their own note", func(t *testing.T) {
		// Create request
		req, _ := http.NewRequest("GET", "/api/notes/note-1", nil)
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
		assert.Equal(t, "note-1", data["id"])
		assert.Equal(t, "user-1", data["ownerId"])
	})

	t.Run("User cannot access another user's note", func(t *testing.T) {
		// Create request
		req, _ := http.NewRequest("GET", "/api/notes/note-3", nil)
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
		result := db.First(&note, "id = ?", "note-new")
		assert.NoError(t, result.Error)
		assert.Equal(t, "user-1", note.OwnerID)
	})
}

// Test updating a note (should enforce ownership)
func TestUpdateNote(t *testing.T) {
	t.Run("User can update their own note", func(t *testing.T) {
		// Create request body
		updateData := map[string]interface{}{
			"title":   "Updated Title",
			"content": "Updated content",
		}
		jsonData, _ := json.Marshal(updateData)

		// Create request
		req, _ := http.NewRequest("PUT", "/api/notes/note-1", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-1")
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify in database
		var note Note
		db.First(&note, "id = ?", "note-1")
		assert.Equal(t, "Updated Title", note.Title)
		assert.Equal(t, "Updated content", note.Content)
		assert.Equal(t, "user-1", note.OwnerID) // Owner shouldn't change
	})

	t.Run("User cannot update another user's note", func(t *testing.T) {
		// Create request body
		updateData := map[string]interface{}{
			"title":   "Should Not Update",
			"content": "Should not update content",
		}
		jsonData, _ := json.Marshal(updateData)

		// Create request
		req, _ := http.NewRequest("PUT", "/api/notes/note-3", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-1")
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Check response - should be forbidden
		assert.Equal(t, http.StatusForbidden, w.Code)

		// Verify in database note is unchanged
		var note Note
		db.First(&note, "id = ?", "note-3")
		assert.NotEqual(t, "Should Not Update", note.Title)
		assert.Equal(t, "user-2", note.OwnerID)
	})
}

// Test deleting a note (should enforce ownership)
func TestDeleteNote(t *testing.T) {
	t.Run("User can delete their own note", func(t *testing.T) {
		// Create request
		req, _ := http.NewRequest("DELETE", "/api/notes/note-2", nil)
		req.Header.Set("X-User-ID", "user-1")
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify in database
		var note Note
		result := db.First(&note, "id = ?", "note-2")
		assert.Error(t, result.Error) // Should not find the note
	})

	t.Run("User cannot delete another user's note", func(t *testing.T) {
		// Create request
		req, _ := http.NewRequest("DELETE", "/api/notes/note-3", nil)
		req.Header.Set("X-User-ID", "user-1")
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Check response - should be forbidden
		assert.Equal(t, http.StatusForbidden, w.Code)

		// Verify in database note still exists
		var note Note
		result := db.First(&note, "id = ?", "note-3")
		assert.NoError(t, result.Error) // Should still find the note
		assert.Equal(t, "user-2", note.OwnerID)
	})
}
