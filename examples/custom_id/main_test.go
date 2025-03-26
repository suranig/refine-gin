package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/suranig/refine-gin/pkg/handler"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCustomIDIntegration(t *testing.T) {
	// Create a new Gin engine
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Create a new SQLite database in memory
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create the users table
	err = db.AutoMigrate(&User{})
	require.NoError(t, err)

	// Create a new resource with custom ID field
	res := resource.NewResource(resource.ResourceConfig{
		Name:        "users",
		Model:       &User{},
		IDFieldName: "UID",
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
		},
	})

	// Create a new repository
	repo := repository.NewGenericRepositoryWithResource(db, res)

	// Register the resource with custom ID parameter
	api := r.Group("/api")
	handler.RegisterResourceForRefine(api, res, repo, "uid")

	// Create a new user
	user := &User{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	// Create request
	body, err := json.Marshal(map[string]interface{}{
		"data": user,
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(w, req)

	// Check response
	t.Logf("Create Response Status: %d", w.Code)
	t.Logf("Create Response Body: %s", w.Body.String())

	require.Equal(t, 201, w.Code)

	// Parse response to get the created user's ID
	var response map[string]map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	uid := response["data"]["uid"].(string)

	// Get the created user
	req = httptest.NewRequest("GET", "/api/users/"+uid, nil)
	w = httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(w, req)

	// Check response
	t.Logf("Get Response Status: %d", w.Code)
	t.Logf("Get Response Body: %s", w.Body.String())
	t.Logf("Get Request URL: %s", req.URL.Path)

	require.Equal(t, 200, w.Code)

	// Parse response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Update the user
	user.Name = "Jane Doe"
	body, err = json.Marshal(map[string]interface{}{
		"data": user,
	})
	require.NoError(t, err)

	req = httptest.NewRequest("PUT", "/api/users/"+uid, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(w, req)

	// Check response
	require.Equal(t, 200, w.Code)

	// Delete the user
	req = httptest.NewRequest("DELETE", "/api/users/"+uid, nil)
	w = httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(w, req)

	// Check response
	require.Equal(t, 204, w.Code)
}
