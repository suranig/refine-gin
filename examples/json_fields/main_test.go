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
	"github.com/stretchr/testify/assert"
	"github.com/suranig/refine-gin/pkg/handler"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestEnvironment() (*gin.Engine, *gorm.DB) {
	// Use test mode
	gin.SetMode(gin.TestMode)

	// Create a test database
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.AutoMigrate(&Domain{})

	// Create router
	r := gin.Default()

	// Create domain resource
	domainResource := resource.NewResource(resource.ResourceConfig{
		Name:  "domains",
		Label: "Domains",
		Model: Domain{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
		},
	})

	// Register resource
	api := r.Group("/api")
	repo := repository.NewGenericRepository(db, domainResource)
	handler.RegisterResource(api, domainResource, repo)

	return r, db
}

func cleanupTest() {
	os.Remove("test.db")
}

func TestMain(m *testing.M) {
	// Run tests
	exitCode := m.Run()

	// Cleanup
	cleanupTest()

	os.Exit(exitCode)
}

func TestCreateDomainWithValidJSON(t *testing.T) {
	r, db := setupTestEnvironment()

	// Create valid domain data
	domainData := map[string]interface{}{
		"name": "test-domain.com",
		"config": map[string]interface{}{
			"email": map[string]interface{}{
				"host":     "smtp.example.com",
				"port":     587,
				"username": "user@example.com",
				"password": "secret",
				"from":     "no-reply@example.com",
			},
			"active": true,
		},
	}

	jsonData, _ := json.Marshal(domainData)

	// Create request
	req := httptest.NewRequest("POST", "/api/domains", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert successful creation
	assert.Equal(t, http.StatusCreated, w.Code)

	// Check if domain was created
	var domain Domain
	result := db.First(&domain)
	assert.Nil(t, result.Error)
	assert.Equal(t, "test-domain.com", domain.Name)
	assert.Equal(t, "smtp.example.com", domain.Config.Email.Host)
	assert.Equal(t, 587, domain.Config.Email.Port)
	assert.True(t, domain.Config.Active)
}

func TestOptionsEndpointIncludesJsonMetadata(t *testing.T) {
	r, _ := setupTestEnvironment()

	// Create request to options endpoint
	req := httptest.NewRequest("OPTIONS", "/api/domains", nil)

	// Perform request
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert success
	assert.Equal(t, http.StatusOK, w.Code)

	// Check response contains JSON field metadata
	responseBody := w.Body.String()

	// Check field is identified as JSON
	assert.Contains(t, responseBody, `"config"`)
	assert.Contains(t, responseBody, `"type":"json"`)
}

func TestUpdateDomainWithNestedJSON(t *testing.T) {
	r, db := setupTestEnvironment()

	// Create initial domain
	domain := Domain{
		Name: "update-test.com",
		Config: Config{
			Email: EmailConfig{
				Host: "old-host.com",
				Port: 25,
			},
			Active: false,
		},
	}
	db.Create(&domain)

	// Update data with nested changes
	updateData := map[string]interface{}{
		"name": "update-test.com",
		"config": map[string]interface{}{
			"email": map[string]interface{}{
				"host": "new-host.com",
				"port": 587,
			},
			"features": map[string]interface{}{
				"enable_registration": true,
				"enable_oauth":        true,
			},
			"active": true,
		},
	}

	jsonData, _ := json.Marshal(updateData)

	// Create request
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/domains/%d", domain.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Log response for debugging
	t.Logf("Update response status: %d", w.Code)
	t.Logf("Update response body: %s", w.Body.String())

	// Assert successful update
	assert.Equal(t, http.StatusOK, w.Code)

	// Check if domain was updated
	var updatedDomain Domain
	db.First(&updatedDomain, domain.ID)

	assert.Equal(t, "new-host.com", updatedDomain.Config.Email.Host)
	assert.Equal(t, 587, updatedDomain.Config.Email.Port)
	assert.True(t, updatedDomain.Config.Features.EnableRegistration)
	assert.True(t, updatedDomain.Config.Features.EnableOAuth)
	assert.True(t, updatedDomain.Config.Active)
}
