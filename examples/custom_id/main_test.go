package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/suranig/refine-gin/pkg/handler"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCustomIDIntegration(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate model
	err = db.AutoMigrate(&User{})
	assert.NoError(t, err)

	// Create resource with custom ID field
	res := resource.NewResource(resource.ResourceConfig{
		Name:        "users",
		Model:       User{},
		IDFieldName: "UID",
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
		},
	})

	// Create repository factory
	repoFactory := repository.NewGormRepositoryFactory(db)

	// Create repository for the resource
	repo := repoFactory.CreateRepository(res)

	// Register resource with custom ID parameter
	api := r.Group("/api")
	handler.RegisterResourceWithOptions(api, res, repo, handler.RegisterOptions{
		IDParamName: "uid",
	})

	// Test Create
	user := User{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	body, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Debug output
	t.Logf("Create Response Status: %d", w.Code)
	t.Logf("Create Response Body: %s", w.Body.String())

	assert.Equal(t, http.StatusCreated, w.Code)

	var createResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &createResponse)
	assert.NoError(t, err)

	data := createResponse["data"].(map[string]interface{})
	uid := data["uid"].(string)
	assert.NotEmpty(t, uid)
	assert.Equal(t, "John Doe", data["name"])

	// Test Get with custom ID parameter
	req, _ = http.NewRequest("GET", "/api/users/"+uid, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Debug output
	t.Logf("Get Response Status: %d", w.Code)
	t.Logf("Get Response Body: %s", w.Body.String())
	t.Logf("Get Request URL: %s", req.URL.String())

	assert.Equal(t, http.StatusOK, w.Code)

	var getResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &getResponse)
	assert.NoError(t, err)

	data = getResponse["data"].(map[string]interface{})
	assert.Equal(t, uid, data["uid"])
	assert.Equal(t, "John Doe", data["name"])

	// Test Update with custom ID parameter
	updatedUser := map[string]interface{}{
		"name": "Jane Doe",
	}

	body, _ = json.Marshal(updatedUser)
	req, _ = http.NewRequest("PUT", "/api/users/"+uid, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updateResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &updateResponse)
	assert.NoError(t, err)

	data = updateResponse["data"].(map[string]interface{})
	assert.Equal(t, uid, data["uid"])
	assert.Equal(t, "Jane Doe", data["name"])

	// Test List
	req, _ = http.NewRequest("GET", "/api/users", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var listResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &listResponse)
	assert.NoError(t, err)

	assert.Equal(t, float64(1), listResponse["total"])

	// Test Delete with custom ID parameter
	req, _ = http.NewRequest("DELETE", "/api/users/"+uid, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var deleteResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &deleteResponse)
	assert.NoError(t, err)

	assert.Equal(t, true, deleteResponse["success"])

	// Verify deletion
	req, _ = http.NewRequest("GET", "/api/users/"+uid, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
