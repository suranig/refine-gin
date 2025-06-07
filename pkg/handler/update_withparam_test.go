package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/suranig/refine-gin/pkg/resource"
)

func TestGenerateUpdateHandlerWithParam_Errors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockResource := new(MockResource)
	mockResource.On("GetName").Return("users").Maybe()
	mockResource.On("GetModel").Return(UpdateUser{}).Maybe()
	mockResource.On("GetIDFieldName").Return("ID").Maybe()
	mockResource.On("GetRelations").Return([]resource.Relation{}).Maybe()
	mockResource.On("GetEditableFields").Return([]string{"name", "email"}).Maybe()
	mockResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "int"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
	}).Maybe()

	// Malformed JSON body
	t.Run("Malformed JSON", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockDTO := new(MockDTOManager)
		mockDTO.On("GetUpdateDTO").Return(&UserUpdateDTO{}).Once()

		r := gin.New()
		r.PUT("/users/:uid", GenerateUpdateHandlerWithParam(mockResource, mockRepo, mockDTO, "uid"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, "/users/1", bytes.NewBufferString(`{"name":}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockDTO.AssertExpectations(t)
	})

	// DTO transform error
	t.Run("DTO transform error", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockDTO := new(MockDTOManager)

		payload := &UserUpdateDTO{Name: "Bob", Email: "bob@example.com"}
		jsonData, _ := json.Marshal(payload)

		mockDTO.On("GetUpdateDTO").Return(&UserUpdateDTO{}).Once()
		mockDTO.On("TransformToModel", mock.AnythingOfType("*handler.UserUpdateDTO")).
			Return(nil, errors.New("dto error")).Once()

		r := gin.New()
		r.PUT("/users/:uid", GenerateUpdateHandlerWithParam(mockResource, mockRepo, mockDTO, "uid"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, "/users/2", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockDTO.AssertExpectations(t)
	})

	// Repository returns record not found
	t.Run("Record not found", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockDTO := new(MockDTOManager)

		dto := &UserUpdateDTO{Name: "Ann", Email: "ann@example.com"}
		jsonData, _ := json.Marshal(dto)
		modelData := map[string]interface{}{"name": "Ann", "email": "ann@example.com"}

		mockDTO.On("GetUpdateDTO").Return(&UserUpdateDTO{}).Once()
		mockDTO.On("TransformToModel", mock.AnythingOfType("*handler.UserUpdateDTO")).
			Return(modelData, nil).Once()

		mockRepo.On("Query", mock.Anything).Return(nil).Once()
		mockRepo.On("Update", mock.Anything, "55", modelData).Return(nil, errors.New("record not found")).Once()

		r := gin.New()
		r.PUT("/users/:uid", GenerateUpdateHandlerWithParam(mockResource, mockRepo, mockDTO, "uid"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, "/users/55", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
		mockDTO.AssertExpectations(t)
	})

	// Generic repository error
	t.Run("Repository error", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockDTO := new(MockDTOManager)

		dto := &UserUpdateDTO{Name: "Err", Email: "err@example.com"}
		jsonData, _ := json.Marshal(dto)
		modelData := map[string]interface{}{"name": "Err", "email": "err@example.com"}

		mockDTO.On("GetUpdateDTO").Return(&UserUpdateDTO{}).Once()
		mockDTO.On("TransformToModel", mock.AnythingOfType("*handler.UserUpdateDTO")).
			Return(modelData, nil).Once()

		mockRepo.On("Query", mock.Anything).Return(nil).Once()
		mockRepo.On("Update", mock.Anything, "77", modelData).Return(nil, errors.New("db error")).Once()

		r := gin.New()
		r.PUT("/users/:uid", GenerateUpdateHandlerWithParam(mockResource, mockRepo, mockDTO, "uid"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, "/users/77", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
		mockDTO.AssertExpectations(t)
	})
}
