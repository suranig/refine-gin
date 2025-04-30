package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stanxing/refine-gin/pkg/dto"
	"github.com/stanxing/refine-gin/pkg/handler"
	"github.com/stanxing/refine-gin/pkg/repository"
	"github.com/stanxing/refine-gin/pkg/resource"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User model for testing
type User struct {
	ID    string `json:"id" gorm:"primaryKey"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func setupIntegrationTest(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Setup database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate schema
	err = db.AutoMigrate(&User{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Create test data
	users := []User{
		{ID: "1", Name: "John Doe", Email: "john@example.com"},
		{ID: "2", Name: "Jane Smith", Email: "jane@example.com"},
	}

	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}
	}

	// Create resource
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

	// Create repository using GenericRepository
	userRepo := repository.NewGenericRepositoryWithResource(db, userResource)

	// Register resource
	api := r.Group("/api")
	handler.RegisterResourceWithDTO(api, userResource, userRepo, &dto.DefaultDTOProvider{
		Model: &User{},
	})

	return r, db
}

func TestIntegrationListUsers(t *testing.T) {
	r, _ := setupIntegrationTest(t)

	// Create request
	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data := response["data"].([]interface{})
	assert.Equal(t, 2, len(data))
}

func TestIntegrationGetUser(t *testing.T) {
	r, _ := setupIntegrationTest(t)

	// Create request
	req := httptest.NewRequest("GET", "/api/users/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "1", data["id"])
	assert.Equal(t, "John Doe", data["name"])
}

func TestIntegrationCreateUser(t *testing.T) {
	r, _ := setupIntegrationTest(t)

	// Create request data
	newUser := User{
		ID:    fmt.Sprintf("%d", time.Now().UnixNano()), // Generate unique ID
		Name:  "Test User",
		Email: "test@example.com",
	}
	body, _ := json.Marshal(newUser)

	// Create request
	req := httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data["id"])
	assert.Equal(t, "Test User", data["name"])
	assert.Equal(t, "test@example.com", data["email"])
}

func TestIntegrationUpdateUser(t *testing.T) {
	r, _ := setupIntegrationTest(t)

	// Update request data
	updatedUser := User{
		Name:  "Updated User",
		Email: "updated@example.com",
	}
	body, _ := json.Marshal(updatedUser)

	// Create request
	req := httptest.NewRequest("PUT", "/api/users/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "1", data["id"])
	assert.Equal(t, "Updated User", data["name"])
	assert.Equal(t, "updated@example.com", data["email"])
}

func TestIntegrationDeleteUser(t *testing.T) {
	r, _ := setupIntegrationTest(t)

	// Create request
	req := httptest.NewRequest("DELETE", "/api/users/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}
