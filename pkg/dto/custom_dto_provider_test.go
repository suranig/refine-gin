package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomDTOProviderGetters(t *testing.T) {
	provider := &CustomDTOProvider{
		Model:       &TestModel{},
		CreateDTO:   &TestCreateDTO{},
		UpdateDTO:   &TestUpdateDTO{},
		ResponseDTO: &TestResponseDTO{},
	}

	createDTO := provider.GetCreateDTO()
	assert.IsType(t, &TestCreateDTO{}, createDTO)
	assert.NotSame(t, provider.CreateDTO, createDTO)

	updateDTO := provider.GetUpdateDTO()
	assert.IsType(t, &TestUpdateDTO{}, updateDTO)
	assert.NotSame(t, provider.UpdateDTO, updateDTO)

	responseDTO := provider.GetResponseDTO()
	assert.IsType(t, &TestResponseDTO{}, responseDTO)
	assert.NotSame(t, provider.ResponseDTO, responseDTO)
}

func TestCustomDTOProviderFallbackToModel(t *testing.T) {
	provider := &CustomDTOProvider{
		Model: &TestModel{},
	}

	assert.Same(t, provider.Model, provider.GetCreateDTO())
	assert.Same(t, provider.Model, provider.GetUpdateDTO())
	assert.Same(t, provider.Model, provider.GetResponseDTO())
}
