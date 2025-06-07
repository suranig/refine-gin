package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// CustomUser is a test model with a custom ID field and nested JSON
type CustomUser struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
	Pref struct {
		Notifications bool `json:"notifications"`
	} `json:"pref"`
}

func (u *CustomUser) SetID(id interface{}) {
	if s, ok := id.(string); ok {
		u.UID = s
	}
}

// Test GenerateCustomUpdateHandler with nested JSON payload
func TestGenerateCustomUpdateHandler_NestedJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockResource := new(MockResource)
	mockRepo := new(MockRepository)

	mockResource.On("GetModel").Return(&CustomUser{}).Maybe()
	mockResource.On("GetIDFieldName").Return("UID").Maybe()

	expected := &CustomUser{UID: "abc123", Name: "Bob"}
	expected.Pref.Notifications = true

	matcher := mock.MatchedBy(func(arg interface{}) bool {
		usr, ok := arg.(*CustomUser)
		return ok && usr.UID == "abc123" && usr.Name == "Bob" && usr.Pref.Notifications
	})

	mockRepo.On("Update", mock.Anything, "abc123", matcher).Return(expected, nil).Once()

	r := gin.New()
	r.PUT("/users/:uid", GenerateCustomUpdateHandler(mockResource, mockRepo, "uid"))

	payload := `{"name":"Bob","pref":{"notifications":true}}`
	req, _ := http.NewRequest(http.MethodPut, "/users/abc123", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "abc123", data["uid"])
	assert.Equal(t, "Bob", data["name"])
	pref := data["pref"].(map[string]interface{})
	assert.Equal(t, true, pref["notifications"])

	mockRepo.AssertExpectations(t)
	mockResource.AssertExpectations(t)
}

// Test GenerateCustomUpdateHandler with invalid JSON payload
func TestGenerateCustomUpdateHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockResource := new(MockResource)
	mockRepo := new(MockRepository)

	mockResource.On("GetModel").Return(&CustomUser{}).Maybe()
	mockResource.On("GetIDFieldName").Return("UID").Maybe()

	r := gin.New()
	r.PUT("/users/:uid", GenerateCustomUpdateHandler(mockResource, mockRepo, "uid"))

	invalid := `{"name": "Bob", "pref": }`
	req, _ := http.NewRequest(http.MethodPut, "/users/abc123", bytes.NewBufferString(invalid))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"].(string), "invalid character")

	mockRepo.AssertExpectations(t)
}

// Test UpdateHandler with nested JSON payload
func TestUpdateHandler_NestedJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockResource := new(MockResource)
	mockRepo := new(MockRepository)

	expected := map[string]interface{}{
		"name": "Alice",
		"config": map[string]interface{}{
			"enabled": true,
		},
	}

	matcher := mock.MatchedBy(func(arg interface{}) bool {
		m, ok := arg.(map[string]interface{})
		return ok && reflect.DeepEqual(m, expected)
	})

	mockRepo.On("Update", mock.Anything, "42", matcher).Return(expected, nil).Once()

	r := gin.New()
	r.PUT("/cfg/:id", func(c *gin.Context) { UpdateHandler(c, mockResource, mockRepo) })

	payload := `{"data": {"name":"Alice", "config": {"enabled": true}}}`
	req, _ := http.NewRequest(http.MethodPut, "/cfg/42", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Alice", data["name"])
	cfg := data["config"].(map[string]interface{})
	assert.Equal(t, true, cfg["enabled"])

	mockRepo.AssertExpectations(t)
}

// Test UpdateHandler with invalid JSON
func TestUpdateHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockResource := new(MockResource)
	mockRepo := new(MockRepository)

	r := gin.New()
	r.PUT("/cfg/:id", func(c *gin.Context) { UpdateHandler(c, mockResource, mockRepo) })

	invalid := `{"data": {"name": "Alice", "config": }}`
	req, _ := http.NewRequest(http.MethodPut, "/cfg/42", bytes.NewBufferString(invalid))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"].(string), "invalid character")
}

// Additional subtests for GenerateCustomUpdateHandler
func TestGenerateCustomUpdateHandler_Subtests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Subtest: repository returns "record not found"
	t.Run("Record not found", func(t *testing.T) {
		mockResource := new(MockResource)
		mockRepo := new(MockRepository)

		mockResource.On("GetModel").Return(&CustomUser{}).Maybe()
		mockResource.On("GetIDFieldName").Return("UID").Maybe()

		matcher := mock.MatchedBy(func(arg interface{}) bool {
			usr, ok := arg.(*CustomUser)
			return ok && usr.UID == "abc123" && usr.Name == "Bob"
		})

		mockRepo.On("Update", mock.Anything, "abc123", matcher).
			Return(nil, errors.New("record not found")).Once()

		r := gin.New()
		r.PUT("/users/:uid", GenerateCustomUpdateHandler(mockResource, mockRepo, "uid"))

		payload := `{"name":"Bob"}`
		req, _ := http.NewRequest(http.MethodPut, "/users/abc123", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockRepo.AssertExpectations(t)
		mockResource.AssertExpectations(t)
	})

	// Subtest: repository returns generic error
	t.Run("Repository error", func(t *testing.T) {
		mockResource := new(MockResource)
		mockRepo := new(MockRepository)

		mockResource.On("GetModel").Return(&CustomUser{}).Maybe()
		mockResource.On("GetIDFieldName").Return("UID").Maybe()

		matcher := mock.MatchedBy(func(arg interface{}) bool {
			usr, ok := arg.(*CustomUser)
			return ok && usr.UID == "abc123" && usr.Name == "Bob"
		})

		mockRepo.On("Update", mock.Anything, "abc123", matcher).
			Return(nil, errors.New("db error")).Once()

		r := gin.New()
		r.PUT("/users/:uid", GenerateCustomUpdateHandler(mockResource, mockRepo, "uid"))

		payload := `{"name":"Bob"}`
		req, _ := http.NewRequest(http.MethodPut, "/users/abc123", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockRepo.AssertExpectations(t)
		mockResource.AssertExpectations(t)
	})

	// Subtest: payload missing ID field should be populated
	t.Run("Adds ID when missing", func(t *testing.T) {
		mockResource := new(MockResource)
		mockRepo := new(MockRepository)

		mockResource.On("GetModel").Return(&CustomUser{}).Maybe()
		mockResource.On("GetIDFieldName").Return("UID").Maybe()

		expected := &CustomUser{UID: "abc123", Name: "Alice"}

		matcher := mock.MatchedBy(func(arg interface{}) bool {
			usr, ok := arg.(*CustomUser)
			return ok && usr.UID == "abc123" && usr.Name == "Alice"
		})

		mockRepo.On("Update", mock.Anything, "abc123", matcher).
			Return(expected, nil).Once()

		r := gin.New()
		r.PUT("/users/:uid", GenerateCustomUpdateHandler(mockResource, mockRepo, "uid"))

		payload := `{"name":"Alice"}`
		req, _ := http.NewRequest(http.MethodPut, "/users/abc123", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "abc123", data["uid"])
		assert.Equal(t, "Alice", data["name"])

		mockRepo.AssertExpectations(t)
		mockResource.AssertExpectations(t)
	})
}
