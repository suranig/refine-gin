package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestModel for DTO tests
type TestModel struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// TestCreateDTO for testing
type TestCreateDTO struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// TestUpdateDTO for testing
type TestUpdateDTO struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// TestResponseDTO for testing
type TestResponseDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func TestDefaultDTOProvider(t *testing.T) {
	// Create a DTO provider
	provider := &DefaultDTOProvider{
		Model:       &TestModel{},
		CreateDTO:   &TestCreateDTO{},
		UpdateDTO:   &TestUpdateDTO{},
		ResponseDTO: &TestResponseDTO{},
	}

	// Test GetCreateDTO
	createDTO := provider.GetCreateDTO()
	assert.IsType(t, &TestCreateDTO{}, createDTO)

	// Test GetUpdateDTO
	updateDTO := provider.GetUpdateDTO()
	assert.IsType(t, &TestUpdateDTO{}, updateDTO)

	// Test GetResponseDTO
	responseDTO := provider.GetResponseDTO()
	assert.IsType(t, &TestResponseDTO{}, responseDTO)

	// Test TransformToModel with CreateDTO
	createData := &TestCreateDTO{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	model, err := provider.TransformToModel(createData)
	assert.NoError(t, err)
	assert.IsType(t, &TestModel{}, model)

	testModel := model.(*TestModel)
	assert.Equal(t, "John Doe", testModel.Name)
	assert.Equal(t, "john@example.com", testModel.Email)
	assert.Equal(t, 30, testModel.Age)

	// Test TransformFromModel
	modelData := &TestModel{
		ID:    "1",
		Name:  "Jane Doe",
		Email: "jane@example.com",
		Age:   25,
	}

	dto, err := provider.TransformFromModel(modelData)
	assert.NoError(t, err)
	assert.IsType(t, &TestResponseDTO{}, dto)

	responseData := dto.(*TestResponseDTO)
	assert.Equal(t, "1", responseData.ID)
	assert.Equal(t, "Jane Doe", responseData.Name)
	assert.Equal(t, "jane@example.com", responseData.Email)
	assert.Equal(t, 25, responseData.Age)
}

func TestDefaultDTOProviderWithoutCustomDTOs(t *testing.T) {
	// Create a DTO provider without custom DTOs
	provider := &DefaultDTOProvider{
		Model: &TestModel{},
	}

	// Test GetCreateDTO
	createDTO := provider.GetCreateDTO()
	assert.IsType(t, &TestModel{}, createDTO)

	// Test GetUpdateDTO
	updateDTO := provider.GetUpdateDTO()
	assert.IsType(t, &TestModel{}, updateDTO)

	// Test GetResponseDTO
	responseDTO := provider.GetResponseDTO()
	assert.IsType(t, &TestModel{}, responseDTO)

	// Test TransformToModel
	modelData := &TestModel{
		ID:    "1",
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	model, err := provider.TransformToModel(modelData)
	assert.NoError(t, err)
	assert.Equal(t, modelData, model)

	// Test TransformFromModel
	dto, err := provider.TransformFromModel(modelData)
	assert.NoError(t, err)
	assert.Equal(t, modelData, dto)
}
