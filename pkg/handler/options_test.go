package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/suranig/refine-gin/pkg/resource"
)

// OptionsMockResource implements Resource interface for options testing
type OptionsMockResource struct {
	mock.Mock
}

func (m *OptionsMockResource) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *OptionsMockResource) GetLabel() string {
	args := m.Called()
	return args.String(0)
}

func (m *OptionsMockResource) GetIcon() string {
	args := m.Called()
	return args.String(0)
}

func (m *OptionsMockResource) GetModel() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *OptionsMockResource) GetFields() []resource.Field {
	args := m.Called()
	return args.Get(0).([]resource.Field)
}

func (m *OptionsMockResource) GetOperations() []resource.Operation {
	args := m.Called()
	return args.Get(0).([]resource.Operation)
}

func (m *OptionsMockResource) HasOperation(op resource.Operation) bool {
	args := m.Called(op)
	return args.Bool(0)
}

func (m *OptionsMockResource) GetDefaultSort() *resource.Sort {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*resource.Sort)
}

func (m *OptionsMockResource) GetFilters() []resource.Filter {
	args := m.Called()
	return args.Get(0).([]resource.Filter)
}

func (m *OptionsMockResource) GetMiddlewares() []interface{} {
	args := m.Called()
	return args.Get(0).([]interface{})
}

func (m *OptionsMockResource) GetRelations() []resource.Relation {
	args := m.Called()
	if args.Get(0) == nil {
		return []resource.Relation{}
	}
	return args.Get(0).([]resource.Relation)
}

func (m *OptionsMockResource) HasRelation(name string) bool {
	args := m.Called(name)
	return args.Bool(0)
}

func (m *OptionsMockResource) GetRelation(name string) *resource.Relation {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	relation := args.Get(0).(resource.Relation)
	return &relation
}

func (m *OptionsMockResource) GetIDFieldName() string {
	args := m.Called()
	return args.String(0)
}

func (m *OptionsMockResource) GetField(name string) *resource.Field {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	field := args.Get(0).(resource.Field)
	return &field
}

func (m *OptionsMockResource) GetSearchable() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

// Add new methods for field lists
func (m *OptionsMockResource) GetFilterableFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *OptionsMockResource) GetSortableFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *OptionsMockResource) GetRequiredFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *OptionsMockResource) GetTableFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *OptionsMockResource) GetFormFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *OptionsMockResource) GetEditableFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func TestGenerateOptionsHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create a mock resource
	mockResource := new(OptionsMockResource)

	// Setup expectations
	mockResource.On("GetName").Return("test_resource")
	mockResource.On("GetLabel").Return("Test Resource")
	mockResource.On("GetIcon").Return("test-icon")
	mockResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "int"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
	})
	mockResource.On("GetOperations").Return([]resource.Operation{
		resource.OperationCreate,
		resource.OperationRead,
		resource.OperationList,
	})
	mockResource.On("GetDefaultSort").Return(nil)
	mockResource.On("GetFilters").Return([]resource.Filter{})
	mockResource.On("GetRelations").Return([]resource.Relation{})
	mockResource.On("GetIDFieldName").Return("ID")
	mockResource.On("GetSearchable").Return([]string{"name"})
	mockResource.On("GetFilterableFields").Return([]string{"id", "name"})
	mockResource.On("GetSortableFields").Return([]string{"id", "name"})
	mockResource.On("GetRequiredFields").Return([]string{"name"})
	mockResource.On("GetTableFields").Return([]string{"id", "name", "description"})
	mockResource.On("GetFormFields").Return([]string{"name", "description"})

	// Register the options handler
	r.OPTIONS("/test_resource", GenerateOptionsHandler(mockResource))

	// Test case 1: First request should return 200 with metadata
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test_resource", nil)
	r.ServeHTTP(w, req)

	// Print response body for debugging
	t.Logf("Response body: %s", w.Body.String())

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"name\":\"test_resource\"")
	assert.Contains(t, w.Body.String(), "\"label\":\"Test Resource\"")
	assert.Contains(t, w.Body.String(), "\"icon\":\"test-icon\"")

	// Verify lists section
	assert.Contains(t, w.Body.String(), "\"lists\":{")
	assert.Contains(t, w.Body.String(), "\"filterable\":[\"id\",\"name\"]")
	assert.Contains(t, w.Body.String(), "\"searchable\":[\"name\"]")
	assert.Contains(t, w.Body.String(), "\"sortable\":[\"id\",\"name\"]")
	assert.Contains(t, w.Body.String(), "\"required\":[\"name\"]")
	assert.Contains(t, w.Body.String(), "\"table\":[\"id\",\"name\",\"description\"]")
	assert.Contains(t, w.Body.String(), "\"form\":[\"name\",\"description\"]")

	// Get the ETag from the response
	etag := w.Header().Get("ETag")
	assert.NotEmpty(t, etag, "ETag should be present in the response")

	// Test case 2: Subsequent request with If-None-Match header should return 304
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("OPTIONS", "/test_resource", nil)
	req.Header.Set("If-None-Match", etag)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotModified, w.Code)
	assert.Empty(t, w.Body.String(), "Body should be empty on 304 response")

	// Test case 3: Request with different If-None-Match should return 200
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("OPTIONS", "/test_resource", nil)
	req.Header.Set("If-None-Match", "\"invalid-etag\"")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"name\":\"test_resource\"")

	// Verify all expectations were met
	mockResource.AssertExpectations(t)
}

func TestRegisterOptionsEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")

	// Create a mock resource
	mockResource := new(OptionsMockResource)

	// Setup expectations
	mockResource.On("GetName").Return("test_resource")
	mockResource.On("GetLabel").Return("Test Resource")
	mockResource.On("GetIcon").Return("test-icon")
	mockResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "int"},
	})
	mockResource.On("GetOperations").Return([]resource.Operation{
		resource.OperationCreate,
		resource.OperationRead,
	})
	mockResource.On("GetDefaultSort").Return(nil)
	mockResource.On("GetFilters").Return([]resource.Filter{})
	mockResource.On("GetRelations").Return([]resource.Relation{})
	mockResource.On("GetIDFieldName").Return("ID")
	mockResource.On("GetSearchable").Return([]string{})
	mockResource.On("GetFilterableFields").Return([]string{"id"})
	mockResource.On("GetSortableFields").Return([]string{"id"})
	mockResource.On("GetRequiredFields").Return([]string{})
	mockResource.On("GetTableFields").Return([]string{"id"})
	mockResource.On("GetFormFields").Return([]string{"id"})

	// Register the options endpoint
	RegisterOptionsEndpoint(apiGroup, mockResource)

	// Test the endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/test_resource", nil)
	r.ServeHTTP(w, req)

	// Print response body for debugging
	t.Logf("Response body: %s", w.Body.String())

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"name\":\"test_resource\"")
	assert.Contains(t, w.Body.String(), "\"lists\":{")

	// Verify all expectations were met
	mockResource.AssertExpectations(t)
}

func TestOptionsHandlerWithMultipleResources(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")

	// Create mock resources
	userResource := new(OptionsMockResource)
	postResource := new(OptionsMockResource)

	// Setup expectations for user resource
	userResource.On("GetName").Return("users")
	userResource.On("GetLabel").Return("Users")
	userResource.On("GetIcon").Return("user")
	userResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
	})
	userResource.On("GetOperations").Return([]resource.Operation{
		resource.OperationCreate,
		resource.OperationRead,
		resource.OperationUpdate,
		resource.OperationDelete,
		resource.OperationList,
	})
	userResource.On("GetDefaultSort").Return(nil)
	userResource.On("GetFilters").Return([]resource.Filter{})
	userResource.On("GetRelations").Return([]resource.Relation{})
	userResource.On("GetIDFieldName").Return("ID")
	userResource.On("GetSearchable").Return([]string{"name"})
	userResource.On("GetFilterableFields").Return([]string{"id", "name"})
	userResource.On("GetSortableFields").Return([]string{"id", "name"})
	userResource.On("GetRequiredFields").Return([]string{"name"})
	userResource.On("GetTableFields").Return([]string{"id", "name"})
	userResource.On("GetFormFields").Return([]string{"name"})

	// Setup expectations for post resource
	postResource.On("GetName").Return("posts")
	postResource.On("GetLabel").Return("Blog Posts")
	postResource.On("GetIcon").Return("post")
	postResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "content", Type: "string"},
	})
	postResource.On("GetOperations").Return([]resource.Operation{
		resource.OperationCreate,
		resource.OperationRead,
		resource.OperationList,
	})
	postResource.On("GetDefaultSort").Return(&resource.Sort{Field: "id", Order: "desc"})
	postResource.On("GetFilters").Return([]resource.Filter{})
	postResource.On("GetRelations").Return([]resource.Relation{})
	postResource.On("GetIDFieldName").Return("ID")
	postResource.On("GetSearchable").Return([]string{"title", "content"})
	postResource.On("GetFilterableFields").Return([]string{"id", "title"})
	postResource.On("GetSortableFields").Return([]string{"id", "title", "content"})
	postResource.On("GetRequiredFields").Return([]string{"title"})
	postResource.On("GetTableFields").Return([]string{"id", "title"})
	postResource.On("GetFormFields").Return([]string{"title", "content"})

	// Register options endpoints for both resources
	RegisterOptionsEndpoint(apiGroup, userResource)
	RegisterOptionsEndpoint(apiGroup, postResource)

	// Test OPTIONS for users resource
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("OPTIONS", "/api/users", nil)
	r.ServeHTTP(w1, req1)

	// Verify user resource response
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Contains(t, w1.Body.String(), "\"name\":\"users\"")
	assert.Contains(t, w1.Body.String(), "\"label\":\"Users\"")
	assert.Contains(t, w1.Body.String(), "\"icon\":\"user\"")
	assert.Contains(t, w1.Body.String(), "\"lists\":{")
	assert.Contains(t, w1.Body.String(), "\"searchable\":[\"name\"]")

	// Test OPTIONS for posts resource
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("OPTIONS", "/api/posts", nil)
	r.ServeHTTP(w2, req2)

	// Verify post resource response
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), "\"name\":\"posts\"")
	assert.Contains(t, w2.Body.String(), "\"label\":\"Blog Posts\"")
	assert.Contains(t, w2.Body.String(), "\"lists\":{")
	assert.Contains(t, w2.Body.String(), "\"searchable\":[\"title\",\"content\"]")

	// Verify all expectations were met
	userResource.AssertExpectations(t)
	postResource.AssertExpectations(t)
}

func TestOptionsHandlerWithJsonFields(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create a mock resource
	mockResource := new(OptionsMockResource)

	// Setup fields with JSON configuration
	jsonField := resource.Field{
		Name: "config",
		Type: "json",
		Json: &resource.JsonConfig{
			DefaultExpanded: true,
			EditorType:      "form",
			Properties: []resource.JsonProperty{
				{
					Path:  "email",
					Label: "Email Configuration",
					Type:  "object",
					Properties: []resource.JsonProperty{
						{
							Path:  "email.host",
							Label: "SMTP Host",
							Type:  "string",
							Validation: &resource.JsonValidation{
								Required: true,
							},
						},
						{
							Path:  "email.port",
							Label: "SMTP Port",
							Type:  "number",
						},
					},
				},
				{
					Path:  "oauth",
					Label: "OAuth Settings",
					Type:  "object",
					Properties: []resource.JsonProperty{
						{
							Path:  "oauth.google_client_id",
							Label: "Google Client ID",
							Type:  "string",
						},
					},
				},
			},
		},
	}

	// Setup expectations
	mockResource.On("GetName").Return("domains")
	mockResource.On("GetLabel").Return("Domains")
	mockResource.On("GetIcon").Return("domain")
	mockResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "int"},
		{Name: "name", Type: "string"},
		jsonField,
	})
	mockResource.On("GetOperations").Return([]resource.Operation{
		resource.OperationCreate,
		resource.OperationRead,
		resource.OperationUpdate,
		resource.OperationList,
	})
	mockResource.On("GetDefaultSort").Return(nil)
	mockResource.On("GetFilters").Return([]resource.Filter{})
	mockResource.On("GetRelations").Return([]resource.Relation{})
	mockResource.On("GetIDFieldName").Return("ID")
	mockResource.On("GetSearchable").Return([]string{"name"})
	mockResource.On("GetFilterableFields").Return([]string{"id", "name"})
	mockResource.On("GetSortableFields").Return([]string{"id", "name"})
	mockResource.On("GetRequiredFields").Return([]string{"name"})
	mockResource.On("GetTableFields").Return([]string{"id", "name"})
	mockResource.On("GetFormFields").Return([]string{"name", "config"})

	// Register the options handler
	r.OPTIONS("/domains", GenerateOptionsHandler(mockResource))

	// Test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/domains", nil)
	r.ServeHTTP(w, req)

	// Verify response status
	assert.Equal(t, http.StatusOK, w.Code)

	// Print response body for debugging
	t.Logf("Response body: %s", w.Body.String())

	// Check basic resource properties
	assert.Contains(t, w.Body.String(), `"name":"domains"`)
	assert.Contains(t, w.Body.String(), `"label":"Domains"`)

	// Check JSON field exists
	assert.Contains(t, w.Body.String(), `"config"`)
	assert.Contains(t, w.Body.String(), `"type":"json"`)

	// Check JSON properties
	assert.Contains(t, w.Body.String(), `"defaultExpanded":true`)
	assert.Contains(t, w.Body.String(), `"editorType":"form"`)

	// Check nested properties
	assert.Contains(t, w.Body.String(), `"email"`)
	assert.Contains(t, w.Body.String(), `"Email Configuration"`)
	assert.Contains(t, w.Body.String(), `"email.host"`)
	assert.Contains(t, w.Body.String(), `"SMTP Host"`)
	assert.Contains(t, w.Body.String(), `"email.port"`)
	assert.Contains(t, w.Body.String(), `"SMTP Port"`)
	assert.Contains(t, w.Body.String(), `"oauth"`)
	assert.Contains(t, w.Body.String(), `"OAuth Settings"`)
	assert.Contains(t, w.Body.String(), `"oauth.google_client_id"`)
	assert.Contains(t, w.Body.String(), `"Google Client ID"`)

	// Verify metadata format includes lists
	assert.Contains(t, w.Body.String(), `"lists":`)
	assert.Contains(t, w.Body.String(), `"form":["name","config"]`)

	// Verify all expectations were met
	mockResource.AssertExpectations(t)
}
