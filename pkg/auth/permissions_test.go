package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/suranig/refine-gin/pkg/resource"
)

func TestPermissionMiddleware(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	middleware := PermissionMiddleware()

	// Create a mock resource with permissions
	mockResource := new(MockResource)
	mockResource.On("GetName").Return("tests")
	mockResource.On("GetPermissions").Return(map[string][]string{
		"read":   {"admin", "editor"},
		"update": {"admin"},
	})

	// Test with user having proper role (admin can read)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Set("resource", mockResource)
	c.Set("operation", resource.OperationRead)
	c.Set("claims", jwt.MapClaims{
		"roles": []interface{}{"admin"},
	})

	middleware(c)
	assert.False(t, c.IsAborted(), "Middleware should not abort for admin role")

	// Test with user having proper role (editor can read)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Set("resource", mockResource)
	c.Set("operation", resource.OperationRead)
	c.Set("claims", jwt.MapClaims{
		"roles": []interface{}{"editor"},
	})

	middleware(c)
	assert.False(t, c.IsAborted(), "Middleware should not abort for editor role")

	// Test with user lacking proper role (user cannot read)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Set("resource", mockResource)
	c.Set("operation", resource.OperationRead)
	c.Set("claims", jwt.MapClaims{
		"roles": []interface{}{"user"},
	})

	middleware(c)
	assert.True(t, c.IsAborted(), "Middleware should abort for user role")
	assert.Equal(t, 403, w.Code, "Response should be 403 Forbidden")

	// Test with user having proper role for specific operation (admin can update)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("PUT", "/", nil)
	c.Set("resource", mockResource)
	c.Set("operation", resource.OperationUpdate)
	c.Set("claims", jwt.MapClaims{
		"roles": []interface{}{"admin"},
	})

	middleware(c)
	assert.False(t, c.IsAborted(), "Middleware should not abort for admin role")

	// Test with user lacking proper role for specific operation (editor cannot update)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("PUT", "/", nil)
	c.Set("resource", mockResource)
	c.Set("operation", resource.OperationUpdate)
	c.Set("claims", jwt.MapClaims{
		"roles": []interface{}{"editor"},
	})

	middleware(c)
	assert.True(t, c.IsAborted(), "Middleware should abort for editor role")
	assert.Equal(t, 403, w.Code, "Response should be 403 Forbidden")

	// Test with operation that has no permissions (create is allowed for all)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/", nil)
	c.Set("resource", mockResource)
	c.Set("operation", resource.OperationCreate)
	c.Set("claims", jwt.MapClaims{
		"roles": []interface{}{"user"},
	})

	middleware(c)
	assert.False(t, c.IsAborted(), "Middleware should not abort for unspecified permissions")
}

func TestPermissionMiddleware_ContextErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	middleware := PermissionMiddleware()

	mockResource := new(MockResource)
	mockResource.On("GetName").Return("tests")

	tests := []struct {
		name    string
		setup   func(*gin.Context)
		message string
	}{
		{
			name: "missing resource",
			setup: func(c *gin.Context) {
				c.Set("operation", resource.OperationRead)
				c.Set("claims", jwt.MapClaims{"roles": []interface{}{"admin"}})
			},
			message: "resource not found in context",
		},
		{
			name: "missing operation",
			setup: func(c *gin.Context) {
				c.Set("resource", mockResource)
				c.Set("claims", jwt.MapClaims{"roles": []interface{}{"admin"}})
			},
			message: "operation not found in context",
		},
		{
			name: "missing claims",
			setup: func(c *gin.Context) {
				c.Set("resource", mockResource)
				c.Set("operation", resource.OperationRead)
			},
			message: "no authentication claims found",
		},
		{
			name: "invalid claims type",
			setup: func(c *gin.Context) {
				c.Set("resource", mockResource)
				c.Set("operation", resource.OperationRead)
				c.Set("claims", "notmap")
			},
			message: "invalid authentication claims",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/", nil)
			tt.setup(c)

			middleware(c)

			assert.True(t, c.IsAborted())
			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.Contains(t, w.Body.String(), tt.message)
		})
	}
}

func TestFilterFieldsByPermission(t *testing.T) {
	// Create test fields with permissions
	fields := []resource.FieldMetadata{
		{
			Name: "id",
			Type: "number",
		},
		{
			Name: "name",
			Type: "string",
			Permissions: map[string][]string{
				"read": {"admin", "editor"},
			},
		},
		{
			Name: "email",
			Type: "string",
			Permissions: map[string][]string{
				"read": {"admin"},
			},
		},
		{
			Name: "age",
			Type: "number",
			Permissions: map[string][]string{
				"update": {"admin"},
			},
		},
	}

	// Test with admin role
	filteredFields := FilterFieldsByPermission(fields, "read", []string{"admin"})
	assert.Len(t, filteredFields, 4, "Admin should see all fields")
	assert.Equal(t, "id", filteredFields[0].Name)
	assert.Equal(t, "name", filteredFields[1].Name)
	assert.Equal(t, "email", filteredFields[2].Name)
	assert.Equal(t, "age", filteredFields[3].Name)

	// Test with editor role
	filteredFields = FilterFieldsByPermission(fields, "read", []string{"editor"})
	assert.Len(t, filteredFields, 3, "Editor should see all fields except email")
	assert.Equal(t, "id", filteredFields[0].Name)
	assert.Equal(t, "name", filteredFields[1].Name)
	assert.Equal(t, "age", filteredFields[2].Name)

	// Test with user role
	filteredFields = FilterFieldsByPermission(fields, "read", []string{"user"})
	assert.Len(t, filteredFields, 2, "User should see only fields without read permissions")
	assert.Equal(t, "id", filteredFields[0].Name)
	assert.Equal(t, "age", filteredFields[1].Name)

	// Test with update operation
	filteredFields = FilterFieldsByPermission(fields, "update", []string{"admin"})
	assert.Len(t, filteredFields, 4, "Admin should see all fields for update")

	// Test with editor on update operation
	filteredFields = FilterFieldsByPermission(fields, "update", []string{"editor"})
	assert.Len(t, filteredFields, 3, "Editor should see all fields except age for update")
	assert.Equal(t, "id", filteredFields[0].Name)
	assert.Equal(t, "name", filteredFields[1].Name)
	assert.Equal(t, "email", filteredFields[2].Name)
}

func TestExtractUserRoles(t *testing.T) {
	// Test with roles as array
	claims := jwt.MapClaims{
		"roles": []interface{}{"admin", "editor"},
	}
	roles := extractUserRoles(claims)
	assert.Len(t, roles, 2)
	assert.Contains(t, roles, "admin")
	assert.Contains(t, roles, "editor")

	// Test with role as string
	claims = jwt.MapClaims{
		"roles": "admin",
	}
	roles = extractUserRoles(claims)
	assert.Len(t, roles, 1)
	assert.Equal(t, "admin", roles[0])

	// Test with no roles
	claims = jwt.MapClaims{}
	roles = extractUserRoles(claims)
	assert.Nil(t, roles)

	// Test with invalid roles format
	claims = jwt.MapClaims{
		"roles": 123,
	}
	roles = extractUserRoles(claims)
	assert.Nil(t, roles)
}

func TestHasResourcePermission(t *testing.T) {
	// Create a mock resource with permissions
	mockResource := new(MockResource)
	mockResource.On("GetPermissions").Return(map[string][]string{
		"read":   {"admin", "editor"},
		"update": {"admin"},
	})

	// Test with user having proper role
	assert.True(t, hasResourcePermission(mockResource, "read", []string{"admin"}))
	assert.True(t, hasResourcePermission(mockResource, "read", []string{"editor"}))
	assert.True(t, hasResourcePermission(mockResource, "update", []string{"admin"}))

	// Test with user lacking proper role
	assert.False(t, hasResourcePermission(mockResource, "read", []string{"user"}))
	assert.False(t, hasResourcePermission(mockResource, "update", []string{"editor"}))

	// Test with operation that has no permissions
	assert.True(t, hasResourcePermission(mockResource, "create", []string{"user"}))

	// Test with resource that has no permissions
	emptyResource := new(MockResource)
	emptyResource.On("GetPermissions").Return(nil)
	assert.True(t, hasResourcePermission(emptyResource, "read", []string{"user"}))
}
