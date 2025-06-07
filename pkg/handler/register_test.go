package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// RegisterTestEntity represents a basic entity for testing
type RegisterTestEntity struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
}

func TestRegisterResourceEndpoints(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	routerGroup := router.Group("/api")

	// Create test DB
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&RegisterTestEntity{}))

	// Create test data with ID 1
	testEntity := &RegisterTestEntity{
		ID:   1,
		Name: "Test Entity",
	}
	result := db.Create(testEntity)
	require.NoError(t, result.Error)

	// Create resource
	res := resource.NewResource(resource.ResourceConfig{
		Name:  "test-entities",
		Model: &RegisterTestEntity{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationRead,
			resource.OperationCreate,
			resource.OperationUpdate,
			resource.OperationDelete,
			resource.OperationCount,
		},
	})

	// Create repository
	repo := repository.NewGenericRepository(db, &RegisterTestEntity{})

	// Register resource
	RegisterResource(routerGroup, res, repo)

	// Test that all endpoints are registered correctly
	paths := []struct {
		method string
		path   string
	}{
		{http.MethodOptions, "/api/test-entities"},
		{http.MethodGet, "/api/test-entities"},
		{http.MethodPost, "/api/test-entities"},
		{http.MethodGet, "/api/test-entities/1"},
		{http.MethodPut, "/api/test-entities/1"},
		{http.MethodDelete, "/api/test-entities/1"},
		{http.MethodGet, "/api/test-entities/count"},
	}

	for _, p := range paths {
		t.Run(p.method+" "+p.path, func(t *testing.T) {
			req := httptest.NewRequest(p.method, p.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Just check that the route exists (not 404)
			assert.NotEqual(t, http.StatusNotFound, w.Code, "Route should exist")
		})
	}
}
