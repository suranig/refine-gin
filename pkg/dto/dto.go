package dto

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// DTOProvider defines an interface for DTO providers
type DTOProvider interface {
	// GetCreateDTO returns the DTO structure for CREATE operations
	GetCreateDTO() interface{}

	// GetUpdateDTO returns the DTO structure for UPDATE operations
	GetUpdateDTO() interface{}

	// GetResponseDTO returns the DTO structure for responses
	GetResponseDTO() interface{}

	// TransformToModel converts DTO to model
	TransformToModel(dto interface{}) (interface{}, error)

	// TransformFromModel converts model to response DTO
	TransformFromModel(model interface{}) (interface{}, error)
}

// DefaultDTOProvider implements DTOProvider using the same model for all operations
type DefaultDTOProvider struct {
	Model interface{}
}

func (p *DefaultDTOProvider) GetCreateDTO() interface{} {
	return p.Model
}

func (p *DefaultDTOProvider) GetUpdateDTO() interface{} {
	return p.Model
}

func (p *DefaultDTOProvider) GetResponseDTO() interface{} {
	return p.Model
}

func (p *DefaultDTOProvider) TransformToModel(dto interface{}) (interface{}, error) {
	// For the default provider, DTO is already the model
	return dto, nil
}

func (p *DefaultDTOProvider) TransformFromModel(model interface{}) (interface{}, error) {
	// For the default provider, model is already the DTO
	return model, nil
}

// CustomDTOProvider implements DTOProvider with different structures for different operations
type CustomDTOProvider struct {
	Model       interface{}
	CreateDTO   interface{}
	UpdateDTO   interface{}
	ResponseDTO interface{}
}

func (p *CustomDTOProvider) GetCreateDTO() interface{} {
	if p.CreateDTO == nil {
		return p.Model
	}
	return reflect.New(reflect.TypeOf(p.CreateDTO).Elem()).Interface()
}

func (p *CustomDTOProvider) GetUpdateDTO() interface{} {
	if p.UpdateDTO == nil {
		return p.Model
	}
	return reflect.New(reflect.TypeOf(p.UpdateDTO).Elem()).Interface()
}

func (p *CustomDTOProvider) GetResponseDTO() interface{} {
	if p.ResponseDTO == nil {
		return p.Model
	}
	return reflect.New(reflect.TypeOf(p.ResponseDTO).Elem()).Interface()
}

func (p *CustomDTOProvider) TransformToModel(dto interface{}) (interface{}, error) {
	// Validate DTO
	if err := validate.Struct(dto); err != nil {
		return nil, err
	}

	// Create a new instance of the model
	modelType := reflect.TypeOf(p.Model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	model := reflect.New(modelType).Interface()

	// Map fields from DTO to model
	// Here you can use a library like mapstructure or manually map fields

	return model, nil
}

func (p *CustomDTOProvider) TransformFromModel(model interface{}) (interface{}, error) {
	// Create a new instance of the response DTO
	dtoType := reflect.TypeOf(p.ResponseDTO)
	if dtoType.Kind() == reflect.Ptr {
		dtoType = dtoType.Elem()
	}
	dto := reflect.New(dtoType).Interface()

	// Map fields from model to DTO
	// Here you can use a library like mapstructure or manually map fields

	return dto, nil
}
