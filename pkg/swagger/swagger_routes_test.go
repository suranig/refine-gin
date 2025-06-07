package swagger

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/suranig/refine-gin/pkg/resource"
)

// SimpleOwnerResource embeds MockResource and adds owner methods
type SimpleOwnerResource struct {
	MockResource
}

func (r SimpleOwnerResource) GetOwnerField() string          { return "ownerId" }
func (r SimpleOwnerResource) IsOwnershipEnforced() bool      { return true }
func (r SimpleOwnerResource) GetDefaultOwnerID() interface{} { return nil }
func (r SimpleOwnerResource) GetOwnerConfig() resource.OwnerConfig {
	return resource.OwnerConfig{OwnerField: "ownerId", EnforceOwnership: true, DefaultOwnerID: nil}
}

func TestRegisterSwaggerRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	res := MockResource{
		name:   "users",
		fields: []resource.Field{{Name: "id", Type: "int"}, {Name: "name", Type: "string"}},
		ops:    []resource.Operation{resource.OperationList},
	}

	info := DefaultSwaggerInfo()
	RegisterSwagger(router.Group(""), []resource.Resource{res}, info)

	// Test /swagger
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/swagger", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Swagger UI")

	// Test /swagger.json
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/swagger.json", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var openAPI OpenAPI
	err := json.Unmarshal(w.Body.Bytes(), &openAPI)
	assert.NoError(t, err)
	assert.Equal(t, info.Title, openAPI.Info.Title)
	assert.Contains(t, openAPI.Paths, "/users")
}

func TestRegisterSwaggerWithOwnerResourcesRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	res := MockResource{
		name:   "users",
		fields: []resource.Field{{Name: "id", Type: "int"}},
		ops:    []resource.Operation{resource.OperationList},
	}

	ownerRes := SimpleOwnerResource{
		MockResource: MockResource{
			name:   "notes",
			fields: []resource.Field{{Name: "id", Type: "string"}, {Name: "ownerId", Type: "string"}},
			ops:    []resource.Operation{resource.OperationList},
		},
	}

	info := DefaultSwaggerInfo()
	RegisterSwaggerWithOwnerResources(router.Group(""), []resource.Resource{res}, []resource.OwnerResource{ownerRes}, info)

	// /swagger
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/swagger", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Swagger UI")

	// /swagger.json
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/swagger.json", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var openAPI OpenAPI
	err := json.Unmarshal(w.Body.Bytes(), &openAPI)
	assert.NoError(t, err)
	assert.Contains(t, openAPI.Paths, "/notes")
	assert.Contains(t, openAPI.Paths, "/users")
}
