package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
)

// dummyHandler is a simple handler used for test custom actions
func dummyHandler(c *gin.Context, r resource.Resource, repo repository.Repository) (interface{}, error) {
	return nil, nil
}

func TestRegisterCustomActions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api")

	// Create a simple resource and mock repository
	res := resource.NewResource(resource.ResourceConfig{
		Name:  "tests",
		Model: struct{}{},
	})
	repo := new(MockRepository)

	actions := []CustomAction{
		{Name: "get-action", Method: http.MethodGet, Handler: dummyHandler},
		{Name: "post-action", Method: http.MethodPost, Handler: dummyHandler},
		{Name: "put-action", Method: http.MethodPut, Handler: dummyHandler},
		{Name: "patch-action", Method: http.MethodPatch, Handler: dummyHandler},
		{Name: "delete-action", Method: http.MethodDelete, Handler: dummyHandler},
	}

	RegisterCustomActions(group, res, repo, actions)

	expected := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/tests/actions/get-action"},
		{http.MethodPost, "/api/tests/actions/post-action"},
		{http.MethodPut, "/api/tests/actions/put-action"},
		{http.MethodPatch, "/api/tests/actions/patch-action"},
		{http.MethodDelete, "/api/tests/actions/delete-action"},
	}

	for _, e := range expected {
		t.Run(e.method+" "+e.path, func(t *testing.T) {
			req := httptest.NewRequest(e.method, e.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusNotFound, w.Code, "Route should exist")
		})
	}
}
