package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestUpdateHandlerErrorScenarios tests error cases for GenerateUpdateHandler
func TestUpdateHandlerErrorScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("invalid JSON body", func(t *testing.T) {
		r, mockRepo, mockRes, mockDTO := setupTest()
		mockDTO.On("GetUpdateDTO").Return(&TestModel{})
		r.PUT("/tests/:id", GenerateUpdateHandler(mockRes, mockRepo, mockDTO))

		req, _ := http.NewRequest(http.MethodPut, "/tests/1", bytes.NewBufferString("{invalid"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Contains(t, resp["error"], "invalid")
		mockDTO.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("transform error", func(t *testing.T) {
		r, mockRepo, mockRes, mockDTO := setupTest()
		mockDTO.On("GetUpdateDTO").Return(&TestModel{})
		mockDTO.On("TransformToModel", mock.Anything).Return(nil, errors.New("transform error"))
		r.PUT("/tests/:id", GenerateUpdateHandler(mockRes, mockRepo, mockDTO))

		body := bytes.NewBufferString(`{"name":"bad"}`)
		req, _ := http.NewRequest(http.MethodPut, "/tests/1", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "transform error", resp["error"])
		mockDTO.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("record not found", func(t *testing.T) {
		r, mockRepo, mockRes, mockDTO := setupTest()
		mockDTO.On("GetUpdateDTO").Return(&TestModel{})
		mockDTO.On("TransformToModel", mock.Anything).Return(TestModel{Name: "x"}, nil)
		mockRepo.On("Update", mock.Anything, "1", mock.Anything).Return(nil, errors.New("record not found"))
		mockRes.On("GetEditableFields").Return([]string{"name"})
		r.PUT("/tests/:id", GenerateUpdateHandler(mockRes, mockRepo, mockDTO))

		body := bytes.NewBufferString(`{"name":"x"}`)
		req, _ := http.NewRequest(http.MethodPut, "/tests/1", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Resource not found", resp["error"])
		mockDTO.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		r, mockRepo, mockRes, mockDTO := setupTest()
		mockDTO.On("GetUpdateDTO").Return(&TestModel{})
		mockDTO.On("TransformToModel", mock.Anything).Return(TestModel{Name: "x"}, nil)
		mockRepo.On("Update", mock.Anything, "1", mock.Anything).Return(nil, errors.New("db error"))
		mockRes.On("GetEditableFields").Return([]string{"name"})
		r.PUT("/tests/:id", GenerateUpdateHandler(mockRes, mockRepo, mockDTO))

		body := bytes.NewBufferString(`{"name":"x"}`)
		req, _ := http.NewRequest(http.MethodPut, "/tests/1", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "db error", resp["error"])
		mockDTO.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}
