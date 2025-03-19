package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/suranig/refine-gin/pkg/dto"
	"github.com/suranig/refine-gin/pkg/handler"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User model for testing
type User struct {
	ID    string `json:"id" gorm:"primaryKey"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserRepository implements the repository.Repository interface
type UserRepository struct {
	db *gorm.DB
}

func (r *UserRepository) List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error) {
	var users []User
	var total int64

	q := r.db.Model(&User{})
	total, err := options.ApplyWithPagination(q, &users)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) Get(ctx context.Context, id interface{}) (interface{}, error) {
	var user User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Create(ctx context.Context, data interface{}) (interface{}, error) {
	user := data.(*User)

	// Generuj ID, jeśli nie zostało podane
	if user.ID == "" {
		user.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	if err := r.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error) {
	user := data.(*User)
	user.ID = id.(string)

	if err := r.db.Save(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Delete(ctx context.Context, id interface{}) error {
	return r.db.Delete(&User{}, "id = ?", id).Error
}

// Count returns the number of resources matching the query options
func (r *UserRepository) Count(ctx context.Context, options query.QueryOptions) (int64, error) {
	var count int64
	query := r.db.Model(&User{})

	// Apply filters
	for field, value := range options.Filters {
		query = query.Where(field+" = ?", value)
	}

	// Apply search if present
	if options.Search != "" {
		query = query.Where("name LIKE ? OR email LIKE ?", "%"+options.Search+"%", "%"+options.Search+"%")
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// Setup integration test environment
func setupIntegrationTest(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Setup SQLite in-memory database
	dbName := fmt.Sprintf("file::memory:test_%d", time.Now().UnixNano())
	db, _ := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	db.AutoMigrate(&User{})

	// Create test data
	users := []User{
		{ID: "1", Name: "John Doe", Email: "john@example.com"},
		{ID: "2", Name: "Jane Smith", Email: "jane@example.com"},
	}

	for _, user := range users {
		db.Create(&user)
	}

	// Create repository
	userRepo := &UserRepository{db: db}

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
		DefaultSort: &resource.Sort{
			Field: "id",
			Order: string(query.SortOrderAsc),
		},
	})

	// Create dto provider
	dtoProvider := &dto.DefaultDTOProvider{
		Model: &User{},
	}

	// Register resource
	api := r.Group("/api")
	handler.RegisterResourceWithDTO(api, userResource, userRepo, dtoProvider)

	return r, db
}

func TestIntegrationListUsers(t *testing.T) {
	r, _ := setupIntegrationTest(t)

	// Create request
	req, _ := http.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "data")
	assert.Contains(t, response, "total")
	assert.Equal(t, float64(2), response["total"])

	data := response["data"].([]interface{})
	assert.Len(t, data, 2)
}

func TestIntegrationGetUser(t *testing.T) {
	r, _ := setupIntegrationTest(t)

	// Create request
	req, _ := http.NewRequest("GET", "/api/users/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "data")
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "1", data["id"])
	assert.Equal(t, "John Doe", data["name"])
	assert.Equal(t, "john@example.com", data["email"])
}

func TestIntegrationCreateUser(t *testing.T) {
	r, _ := setupIntegrationTest(t)

	// Create request
	newUser := struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}{
		Name:  "New User",
		Email: "new@example.com",
	}

	body, _ := json.Marshal(newUser)
	req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "data")
	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data["id"])
	assert.Equal(t, "New User", data["name"])
	assert.Equal(t, "new@example.com", data["email"])
}

func TestIntegrationUpdateUser(t *testing.T) {
	r, _ := setupIntegrationTest(t)

	// Create request
	updatedUser := User{
		Name:  "John Updated",
		Email: "john.updated@example.com",
	}

	body, _ := json.Marshal(updatedUser)
	req, _ := http.NewRequest("PUT", "/api/users/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify user was updated
	req, _ = http.NewRequest("GET", "/api/users/1", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "John Updated", data["name"])
	assert.Equal(t, "john.updated@example.com", data["email"])
}

func TestIntegrationDeleteUser(t *testing.T) {
	r, _ := setupIntegrationTest(t)

	// Create request
	req, _ := http.NewRequest("DELETE", "/api/users/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify user was deleted
	req, _ = http.NewRequest("GET", "/api/users/1", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
