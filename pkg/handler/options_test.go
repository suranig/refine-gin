package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stanxing/refine-gin/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	return args.Get(0).([]string)
}

func (m *OptionsMockResource) GetPermissions() map[string][]string {
	args := m.Called()
	if result := args.Get(0); result != nil {
		return result.(map[string][]string)
	}
	return nil
}

func (m *OptionsMockResource) HasPermission(operation string, role string) bool {
	args := m.Called(operation, role)
	return args.Bool(0)
}

// GetFormLayout returns the form layout configuration
func (m *OptionsMockResource) GetFormLayout() *resource.FormLayout {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*resource.FormLayout)
}

func TestGenerateOptionsHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create a mock resource
	mockResource := new(OptionsMockResource)

	// Setup expectations
	mockResource.On("GetName").Return("users")
	mockResource.On("GetLabel").Return("Users")
	mockResource.On("GetIcon").Return("user")
	mockResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "int"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "status", Type: "string"},
	})
	mockResource.On("GetOperations").Return([]resource.Operation{
		resource.OperationList,
	})
	mockResource.On("GetDefaultSort").Return(nil)
	mockResource.On("GetFilters").Return([]resource.Filter{})
	mockResource.On("GetRelations").Return([]resource.Relation{})
	mockResource.On("GetIDFieldName").Return("id")
	mockResource.On("GetSearchable").Return([]string{"name", "email"})
	mockResource.On("GetFilterableFields").Return([]string{"name", "email", "status"})
	mockResource.On("GetSortableFields").Return([]string{"name", "email", "createdAt"})
	mockResource.On("GetTableFields").Return([]string{"id", "name", "email", "status"})
	mockResource.On("GetFormFields").Return([]string{"name", "email", "status"})
	mockResource.On("GetRequiredFields").Return([]string{"name", "email"})
	mockResource.On("GetPermissions").Return(nil)

	// Register the options handler
	r.OPTIONS("/users", GenerateOptionsHandler(mockResource))

	// Test case 1: First request should return 200 with metadata
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/users", nil)
	r.ServeHTTP(w, req)

	// Print response body for debugging
	t.Logf("Response body: %s", w.Body.String())

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"name\":\"users\"")
	assert.Contains(t, w.Body.String(), "\"label\":\"Users\"")
	assert.Contains(t, w.Body.String(), "\"icon\":\"user\"")

	// Verify lists section
	assert.Contains(t, w.Body.String(), "\"lists\":{")
	assert.Contains(t, w.Body.String(), "\"filterable\":[\"name\",\"email\",\"status\"]")
	assert.Contains(t, w.Body.String(), "\"searchable\":[\"name\",\"email\"]")
	assert.Contains(t, w.Body.String(), "\"sortable\":[\"name\",\"email\",\"createdAt\"]")
	assert.Contains(t, w.Body.String(), "\"required\":[\"name\",\"email\"]")
	assert.Contains(t, w.Body.String(), "\"table\":[\"id\",\"name\",\"email\",\"status\"]")
	assert.Contains(t, w.Body.String(), "\"form\":[\"name\",\"email\",\"status\"]")

	// Get the ETag from the response
	etag := w.Header().Get("ETag")
	assert.NotEmpty(t, etag, "ETag should be present in the response")

	// Test case 2: Subsequent request with If-None-Match header should return 304
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("OPTIONS", "/users", nil)
	req.Header.Set("If-None-Match", etag)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotModified, w.Code)
	assert.Empty(t, w.Body.String(), "Body should be empty on 304 response")

	// Test case 3: Request with different If-None-Match should return 200
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("OPTIONS", "/users", nil)
	req.Header.Set("If-None-Match", "\"invalid-etag\"")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"name\":\"users\"")

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
	mockResource.On("GetName").Return("users")
	mockResource.On("GetLabel").Return("Users")
	mockResource.On("GetIcon").Return("user")
	mockResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "int"},
	})
	mockResource.On("GetOperations").Return([]resource.Operation{
		resource.OperationList,
	})
	mockResource.On("GetDefaultSort").Return(nil)
	mockResource.On("GetFilters").Return([]resource.Filter{})
	mockResource.On("GetRelations").Return([]resource.Relation{})
	mockResource.On("GetIDFieldName").Return("id")
	mockResource.On("GetSearchable").Return([]string{})
	mockResource.On("GetFilterableFields").Return([]string{"id"})
	mockResource.On("GetSortableFields").Return([]string{"id"})
	mockResource.On("GetRequiredFields").Return([]string{})
	mockResource.On("GetTableFields").Return([]string{"id"})
	mockResource.On("GetFormFields").Return([]string{"id"})
	mockResource.On("GetPermissions").Return(nil)

	// Register the options endpoint
	RegisterOptionsEndpoint(apiGroup, mockResource)

	// Test the endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/users", nil)
	r.ServeHTTP(w, req)

	// Print response body for debugging
	t.Logf("Response body: %s", w.Body.String())

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"name\":\"users\"")
	assert.Contains(t, w.Body.String(), "\"lists\":{")

	// Verify all expectations were met
	mockResource.AssertExpectations(t)
}

func TestOptionsHandlerWithMultipleResources(t *testing.T) {
	// Setup
	router := gin.New()
	apiGroup := router.Group("/api")

	// Create a mock resource for users
	usersResource := new(OptionsMockResource)
	usersResource.On("GetName").Return("users")
	usersResource.On("GetLabel").Return("Users")
	usersResource.On("GetIcon").Return("user")
	usersResource.On("GetOperations").Return([]resource.Operation{resource.OperationList})
	usersResource.On("GetPermissions").Return(nil)

	// Create a mock resource for products
	productsResource := new(OptionsMockResource)
	productsResource.On("GetName").Return("products")
	productsResource.On("GetLabel").Return("Products")
	productsResource.On("GetIcon").Return("box")
	productsResource.On("GetOperations").Return([]resource.Operation{resource.OperationList})
	productsResource.On("GetPermissions").Return(nil)

	// Setup expectations for user resource
	usersResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
	})
	usersResource.On("GetDefaultSort").Return(nil)
	usersResource.On("GetFilters").Return([]resource.Filter{})
	usersResource.On("GetRelations").Return([]resource.Relation{})
	usersResource.On("GetIDFieldName").Return("ID")
	usersResource.On("GetSearchable").Return([]string{"name"})
	usersResource.On("GetFilterableFields").Return([]string{"id", "name"})
	usersResource.On("GetSortableFields").Return([]string{"id", "name"})
	usersResource.On("GetRequiredFields").Return([]string{"name"})
	usersResource.On("GetTableFields").Return([]string{"id", "name"})
	usersResource.On("GetFormFields").Return([]string{"name"})

	// Setup expectations for product resource
	productsResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "content", Type: "string"},
	})
	productsResource.On("GetDefaultSort").Return(&resource.Sort{Field: "id", Order: "desc"})
	productsResource.On("GetFilters").Return([]resource.Filter{})
	productsResource.On("GetRelations").Return([]resource.Relation{})
	productsResource.On("GetIDFieldName").Return("ID")
	productsResource.On("GetSearchable").Return([]string{"title", "content"})
	productsResource.On("GetFilterableFields").Return([]string{"id", "title"})
	productsResource.On("GetSortableFields").Return([]string{"id", "title", "content"})
	productsResource.On("GetRequiredFields").Return([]string{"title"})
	productsResource.On("GetTableFields").Return([]string{"id", "title"})
	productsResource.On("GetFormFields").Return([]string{"title", "content"})

	// Register options endpoints for both resources
	RegisterOptionsEndpoint(apiGroup, usersResource)
	RegisterOptionsEndpoint(apiGroup, productsResource)

	// Test OPTIONS for users resource
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("OPTIONS", "/api/users", nil)
	router.ServeHTTP(w1, req1)

	// Verify user resource response
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Contains(t, w1.Body.String(), "\"name\":\"users\"")
	assert.Contains(t, w1.Body.String(), "\"label\":\"Users\"")
	assert.Contains(t, w1.Body.String(), "\"icon\":\"user\"")
	assert.Contains(t, w1.Body.String(), "\"lists\":{")
	assert.Contains(t, w1.Body.String(), "\"searchable\":[\"name\"]")

	// Test OPTIONS for products resource
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("OPTIONS", "/api/products", nil)
	router.ServeHTTP(w2, req2)

	// Verify product resource response
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), "\"name\":\"products\"")
	assert.Contains(t, w2.Body.String(), "\"label\":\"Products\"")
	assert.Contains(t, w2.Body.String(), "\"lists\":{")
	assert.Contains(t, w2.Body.String(), "\"searchable\":[\"title\",\"content\"]")

	// Verify all expectations were met
	usersResource.AssertExpectations(t)
	productsResource.AssertExpectations(t)
}

func TestOptionsHandlerWithJsonFields(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create a mock resource with JSON fields
	mockResource := new(OptionsMockResource)
	mockResource.On("GetName").Return("users")
	mockResource.On("GetLabel").Return("Users")
	mockResource.On("GetIcon").Return("user")
	mockResource.On("GetPermissions").Return(nil)

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
	r.OPTIONS("/users", GenerateOptionsHandler(mockResource))

	// Test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/users", nil)
	r.ServeHTTP(w, req)

	// Verify response status
	assert.Equal(t, http.StatusOK, w.Code)

	// Print response body for debugging
	t.Logf("Response body: %s", w.Body.String())

	// Check basic resource properties
	assert.Contains(t, w.Body.String(), `"name":"users"`)
	assert.Contains(t, w.Body.String(), `"label":"Users"`)

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

func TestOptionsWithAuth(t *testing.T) {
	// Setup
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("claims", jwt.MapClaims{
			"sub":   "1",
			"roles": []interface{}{"admin"},
		})
		c.Next()
	})

	userResource := new(OptionsMockResource)
	userResource.On("GetName").Return("users")
	userResource.On("GetLabel").Return("Users")
	userResource.On("GetIcon").Return("user")
	userResource.On("GetOperations").Return([]resource.Operation{resource.OperationList})
	userResource.On("GetIDFieldName").Return("id")
	userResource.On("GetDefaultSort").Return(nil)
	userResource.On("GetFilters").Return([]resource.Filter{})
	userResource.On("GetSearchable").Return([]string{"name", "email"})
	userResource.On("GetFilterableFields").Return([]string{"name", "email", "status"})
	userResource.On("GetSortableFields").Return([]string{"name", "email", "createdAt"})
	userResource.On("GetTableFields").Return([]string{"id", "name", "email", "status"})
	userResource.On("GetFormFields").Return([]string{"name", "email", "status"})
	userResource.On("GetRequiredFields").Return([]string{"name", "email"})
	userResource.On("GetPermissions").Return(nil)
	userResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "int"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "status", Type: "string"},
	})
	userResource.On("GetRelations").Return([]resource.Relation{})

	postResource := new(OptionsMockResource)
	postResource.On("GetName").Return("posts")
	postResource.On("GetLabel").Return("Posts")
	postResource.On("GetIcon").Return("file-text")
	postResource.On("GetOperations").Return([]resource.Operation{resource.OperationList})
	postResource.On("GetIDFieldName").Return("id")
	postResource.On("GetDefaultSort").Return(nil)
	postResource.On("GetFilters").Return([]resource.Filter{})
	postResource.On("GetSearchable").Return([]string{"title", "content"})
	postResource.On("GetFilterableFields").Return([]string{"title", "status"})
	postResource.On("GetSortableFields").Return([]string{"title", "createdAt"})
	postResource.On("GetTableFields").Return([]string{"id", "title", "status"})
	postResource.On("GetFormFields").Return([]string{"title", "content", "status"})
	postResource.On("GetRequiredFields").Return([]string{"title"})
	postResource.On("GetPermissions").Return(nil)
	postResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "int"},
		{Name: "title", Type: "string"},
		{Name: "content", Type: "string"},
		{Name: "status", Type: "string"},
	})
	postResource.On("GetRelations").Return([]resource.Relation{})

	// Register options endpoints for both resources
	RegisterOptionsEndpoint(router.Group("/api"), userResource)
	RegisterOptionsEndpoint(router.Group("/api"), postResource)

	// Test OPTIONS for users resource
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("OPTIONS", "/api/users", nil)
	router.ServeHTTP(w1, req1)

	// Verify user resource response
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Contains(t, w1.Body.String(), "\"name\":\"users\"")
	assert.Contains(t, w1.Body.String(), "\"label\":\"Users\"")
	assert.Contains(t, w1.Body.String(), "\"icon\":\"user\"")
	assert.Contains(t, w1.Body.String(), "\"lists\":{")
	assert.Contains(t, w1.Body.String(), "\"searchable\":[\"name\",\"email\"]")

	// Test OPTIONS for posts resource
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("OPTIONS", "/api/posts", nil)
	router.ServeHTTP(w2, req2)

	// Verify post resource response
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), "\"name\":\"posts\"")
	assert.Contains(t, w2.Body.String(), "\"label\":\"Posts\"")
	assert.Contains(t, w2.Body.String(), "\"lists\":{")
	assert.Contains(t, w2.Body.String(), "\"searchable\":[\"title\",\"content\"]")

	// Verify all expectations were met
	userResource.AssertExpectations(t)
	postResource.AssertExpectations(t)
}

func TestGenerateOptionsHandlerWithETag(t *testing.T) {
	// Create a mock resource
	mockResource := new(OptionsMockResource)
	mockResource.On("GetName").Return("users")
	mockResource.On("GetLabel").Return("Users")
	mockResource.On("GetIcon").Return("user")
	mockResource.On("GetOperations").Return([]resource.Operation{resource.OperationList})
	mockResource.On("GetIDFieldName").Return("id")
	mockResource.On("GetDefaultSort").Return(nil)
	mockResource.On("GetFilters").Return([]resource.Filter{})
	mockResource.On("GetSearchable").Return([]string{"name", "email"})
	mockResource.On("GetFilterableFields").Return([]string{"name", "email", "status"})
	mockResource.On("GetSortableFields").Return([]string{"name", "email", "createdAt"})
	mockResource.On("GetTableFields").Return([]string{"id", "name", "email", "status"})
	mockResource.On("GetFormFields").Return([]string{"name", "email", "status"})
	mockResource.On("GetRequiredFields").Return([]string{"name", "email"})
	mockResource.On("GetPermissions").Return(nil)
	mockResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "int"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "status", Type: "string"},
	})
	mockResource.On("GetRelations").Return([]resource.Relation{})

	// Register the options handler
	r := gin.New()
	r.OPTIONS("/users", GenerateOptionsHandler(mockResource))

	// Test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/users", nil)
	r.ServeHTTP(w, req)

	// Verify response status
	assert.Equal(t, http.StatusOK, w.Code)

	// Print response body for debugging
	t.Logf("Response body: %s", w.Body.String())

	// Check basic resource properties
	assert.Contains(t, w.Body.String(), `"name":"users"`)
	assert.Contains(t, w.Body.String(), `"label":"Users"`)
	assert.Contains(t, w.Body.String(), `"icon":"user"`)

	// Check lists section
	assert.Contains(t, w.Body.String(), `"lists":`)
	assert.Contains(t, w.Body.String(), `"filterable":["name","email","status"]`)
	assert.Contains(t, w.Body.String(), `"searchable":["name","email"]`)
	assert.Contains(t, w.Body.String(), `"sortable":["name","email","createdAt"]`)
	assert.Contains(t, w.Body.String(), `"required":["name","email"]`)
	assert.Contains(t, w.Body.String(), `"table":["id","name","email","status"]`)
	assert.Contains(t, w.Body.String(), `"form":["name","email","status"]`)

	// Get the ETag from the response
	etag := w.Header().Get("ETag")
	assert.NotEmpty(t, etag, "ETag should be present in the response")

	// Verify all expectations were met
	mockResource.AssertExpectations(t)
}
