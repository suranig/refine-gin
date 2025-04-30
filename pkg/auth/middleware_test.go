package auth

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

// MockAuthorizationProvider for testing
type MockAuthorizationProvider struct {
	mock.Mock
}

func (m *MockAuthorizationProvider) CanAccess(c *gin.Context, res resource.Resource, op resource.Operation) bool {
	args := m.Called(c, res, op)
	return args.Bool(0)
}

func (m *MockAuthorizationProvider) CanAccessRecord(c *gin.Context, res resource.Resource, op resource.Operation, record interface{}) bool {
	args := m.Called(c, res, op, record)
	return args.Bool(0)
}

// Test record for testing
type TestRecord struct {
	ID     string
	UserID string
	Name   string
}

func TestDefaultAuthorizationProvider(t *testing.T) {
	// Setup
	provider := NewDefaultAuthorizationProvider()

	// Create a mock resource
	mockResource := new(MockResource)
	mockResource.On("GetName").Return("tests")
	mockResource.On("GetIDFieldName").Return("ID")

	// Create a test context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	// Test with no rules (should allow by default)
	assert.True(t, provider.CanAccess(c, mockResource, resource.OperationList))
	assert.True(t, provider.CanAccessRecord(c, mockResource, resource.OperationRead, nil))

	// Add a rule that denies access
	provider.AddRule("tests", resource.OperationList, func(c *gin.Context, record interface{}) bool {
		return false
	})

	// Test with rule that denies access
	assert.False(t, provider.CanAccess(c, mockResource, resource.OperationList))

	// Test with operation that has no rule (should allow by default)
	assert.True(t, provider.CanAccess(c, mockResource, resource.OperationCreate))

	// Add a rule that allows access based on record
	provider.AddRule("tests", resource.OperationRead, func(c *gin.Context, record interface{}) bool {
		if record == nil {
			return false
		}
		testRecord := record.(*TestRecord)
		return testRecord.UserID == "1"
	})

	// Test with record that matches rule
	record := &TestRecord{ID: "1", UserID: "1", Name: "Test"}
	assert.True(t, provider.CanAccessRecord(c, mockResource, resource.OperationRead, record))

	// Test with record that doesn't match rule
	record = &TestRecord{ID: "2", UserID: "2", Name: "Test"}
	assert.False(t, provider.CanAccessRecord(c, mockResource, resource.OperationRead, record))
}

func TestJWTAuthorizationProvider(t *testing.T) {
	// Setup
	provider := NewJWTAuthorizationProvider()

	// Create a mock resource
	mockResource := new(MockResource)
	mockResource.On("GetName").Return("tests")
	mockResource.On("GetIDFieldName").Return("ID")
	mockResource.On("GetField", mock.Anything).Return(nil)
	mockResource.On("GetSearchable").Return([]string{})

	// Create a test context with claims
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	c.Set("claims", jwt.MapClaims{
		"sub":   "1",
		"roles": []interface{}{"admin", "user"},
	})

	// Test with no rules (should allow by default)
	assert.True(t, provider.CanAccess(c, mockResource, resource.OperationList))
	assert.True(t, provider.CanAccessRecord(c, mockResource, resource.OperationRead, nil))

	// Add a rule that requires admin role
	provider.AddRule("tests", resource.OperationList, HasRole("admin"))

	// Test with rule that allows access
	assert.True(t, provider.CanAccess(c, mockResource, resource.OperationList))

	// Test with context that has no claims
	c, _ = gin.CreateTestContext(nil)
	assert.False(t, provider.CanAccess(c, mockResource, resource.OperationList))

	// Test HasAnyRole
	c, _ = gin.CreateTestContext(nil)
	c.Set("claims", jwt.MapClaims{
		"sub":   "1",
		"roles": []interface{}{"editor"},
	})

	provider.AddRule("tests", resource.OperationCreate, HasAnyRole("admin", "editor"))
	assert.True(t, provider.CanAccess(c, mockResource, resource.OperationCreate))

	// Test HasAllRoles
	c, _ = gin.CreateTestContext(nil)
	c.Set("claims", jwt.MapClaims{
		"sub":   "1",
		"roles": []interface{}{"admin", "editor"},
	})

	provider.AddRule("tests", resource.OperationUpdate, HasAllRoles("admin", "editor"))
	assert.True(t, provider.CanAccess(c, mockResource, resource.OperationUpdate))

	c, _ = gin.CreateTestContext(nil)
	c.Set("claims", jwt.MapClaims{
		"sub":   "1",
		"roles": []interface{}{"admin"},
	})

	assert.False(t, provider.CanAccess(c, mockResource, resource.OperationUpdate))

	// Test IsOwner
	c, _ = gin.CreateTestContext(nil)
	c.Set("claims", jwt.MapClaims{
		"sub":   "1",
		"roles": []interface{}{"user"},
	})

	provider.AddRule("tests", resource.OperationDelete, IsOwner("sub", "UserID"))

	record := &TestRecord{ID: "1", UserID: "1", Name: "Test"}
	assert.True(t, provider.CanAccessRecord(c, mockResource, resource.OperationDelete, record))

	record = &TestRecord{ID: "2", UserID: "2", Name: "Test"}
	assert.False(t, provider.CanAccessRecord(c, mockResource, resource.OperationDelete, record))
}

func TestAuthorizationMiddleware(t *testing.T) {
	// Setup
	mockProvider := new(MockAuthorizationProvider)
	middleware := AuthorizationMiddleware(mockProvider)

	// Create a mock resource
	mockResource := new(MockResource)

	// Test with missing resource in context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)

	middleware(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Forbidden")

	// Test with missing operation in context
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Set("resource", mockResource)

	middleware(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Forbidden")

	// Test with provider that denies access
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Set("resource", mockResource)
	c.Set("operation", resource.OperationList)

	mockProvider.On("CanAccess", c, mockResource, resource.OperationList).Return(false)

	middleware(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Forbidden")

	// Test with provider that allows access
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Set("resource", mockResource)
	c.Set("operation", resource.OperationList)
	c.Set("nextCalled", false)

	mockProvider.On("CanAccess", c, mockResource, resource.OperationList).Return(true)

	originalHandler := middleware
	middleware = func(c *gin.Context) {
		originalHandler(c)
		if !c.IsAborted() {
			c.Set("nextCalled", true)
		}
	}

	middleware(c)

	nextCalled, _ := c.Get("nextCalled")
	assert.True(t, nextCalled.(bool))
}

// MockResource for testing
type MockResource struct {
	mock.Mock
}

func (m *MockResource) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockResource) GetLabel() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockResource) GetIcon() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockResource) GetModel() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockResource) GetFields() []resource.Field {
	args := m.Called()
	if args.Get(0) == nil {
		return []resource.Field{}
	}
	return args.Get(0).([]resource.Field)
}

func (m *MockResource) GetOperations() []resource.Operation {
	args := m.Called()
	if args.Get(0) == nil {
		return []resource.Operation{}
	}
	return args.Get(0).([]resource.Operation)
}

func (m *MockResource) HasOperation(op resource.Operation) bool {
	args := m.Called(op)
	return args.Bool(0)
}

func (m *MockResource) GetDefaultSort() *resource.Sort {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*resource.Sort)
}

func (m *MockResource) GetFilters() []resource.Filter {
	args := m.Called()
	if args.Get(0) == nil {
		return []resource.Filter{}
	}
	return args.Get(0).([]resource.Filter)
}

func (m *MockResource) GetMiddlewares() []interface{} {
	args := m.Called()
	if args.Get(0) == nil {
		return []interface{}{}
	}
	return args.Get(0).([]interface{})
}

func (m *MockResource) GetRelations() []resource.Relation {
	args := m.Called()
	if args.Get(0) == nil {
		return []resource.Relation{}
	}
	return args.Get(0).([]resource.Relation)
}

func (m *MockResource) HasRelation(name string) bool {
	args := m.Called(name)
	return args.Bool(0)
}

func (m *MockResource) GetRelation(name string) *resource.Relation {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	relation := args.Get(0).(resource.Relation)
	return &relation
}

func (m *MockResource) GetIDFieldName() string {
	args := m.Called()
	return args.String(0)
}

// GetField returns a field by name
func (m *MockResource) GetField(name string) *resource.Field {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	field := args.Get(0).(resource.Field)
	return &field
}

// GetSearchable returns searchable field names
func (m *MockResource) GetSearchable() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

// GetFilterableFields returns filterable field names
func (m *MockResource) GetFilterableFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

// GetSortableFields returns sortable field names
func (m *MockResource) GetSortableFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

// GetRequiredFields returns required field names
func (m *MockResource) GetRequiredFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

// GetTableFields returns table field names
func (m *MockResource) GetTableFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

// GetFormFields returns form field names
func (m *MockResource) GetFormFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

// GetEditableFields returns editable field names
func (m *MockResource) GetEditableFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

// Add the GetPermissions method to MockResource to implement the Resource interface
func (m *MockResource) GetPermissions() map[string][]string {
	args := m.Called()
	if result := args.Get(0); result != nil {
		return result.(map[string][]string)
	}
	return nil
}

// Add the HasPermission method to MockResource to implement the Resource interface
func (m *MockResource) HasPermission(operation string, role string) bool {
	args := m.Called(operation, role)
	return args.Bool(0)
}

func (m *MockResource) GetFormLayout() *resource.FormLayout {
	return nil
}
