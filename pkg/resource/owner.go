package resource

import (
	"reflect"
)

// OwnerConfig contains configuration for creating owner-based resources
type OwnerConfig struct {
	// Name of the field in the model that stores the owner ID
	OwnerField string

	// Whether to enforce ownership checks
	EnforceOwnership bool

	// Default owner ID to use if none is provided in the context
	DefaultOwnerID interface{}
}

// DefaultOwnerConfig returns a default owner configuration
func DefaultOwnerConfig() OwnerConfig {
	return OwnerConfig{
		OwnerField:       "OwnerID",
		EnforceOwnership: true,
		DefaultOwnerID:   nil,
	}
}

// OwnerResource extends the Resource interface with ownership functionality
type OwnerResource interface {
	Resource

	// Get the name of the field that stores the owner ID
	GetOwnerField() string

	// Check if ownership enforcement is enabled
	IsOwnershipEnforced() bool

	// Get the default owner ID to use if none is provided
	GetDefaultOwnerID() interface{}

	// Get the owner configuration
	GetOwnerConfig() OwnerConfig
}

// DefaultOwnerResource wraps an existing resource with ownership functionality
type DefaultOwnerResource struct {
	Resource
	Config OwnerConfig
}

// NewOwnerResource creates a new owner resource from an existing resource
func NewOwnerResource(res Resource, config OwnerConfig) OwnerResource {
	// Validate the owner field exists in the model
	if config.OwnerField != "" {
		modelType := reflect.TypeOf(res.GetModel())
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}

		// Check if owner field exists in the model
		if _, found := modelType.FieldByName(config.OwnerField); !found {
			panic("Owner field '" + config.OwnerField + "' not found in model " + modelType.Name())
		}
	}

	return &DefaultOwnerResource{
		Resource: res,
		Config:   config,
	}
}

// GetOwnerField returns the name of the field that stores the owner ID
func (r *DefaultOwnerResource) GetOwnerField() string {
	return r.Config.OwnerField
}

// IsOwnershipEnforced checks if ownership enforcement is enabled
func (r *DefaultOwnerResource) IsOwnershipEnforced() bool {
	return r.Config.EnforceOwnership
}

// GetDefaultOwnerID returns the default owner ID to use if none is provided
func (r *DefaultOwnerResource) GetDefaultOwnerID() interface{} {
	return r.Config.DefaultOwnerID
}

// GetOwnerConfig returns the owner configuration
func (r *DefaultOwnerResource) GetOwnerConfig() OwnerConfig {
	return r.Config
}

// PromoteToOwnerResource converts a regular resource to an owner resource with default configuration
func PromoteToOwnerResource(res Resource) OwnerResource {
	if ownerRes, ok := res.(OwnerResource); ok {
		return ownerRes
	}
	return NewOwnerResource(res, DefaultOwnerConfig())
}

// IsOwnerResource checks if a resource is an owner resource
func IsOwnerResource(res Resource) bool {
	_, ok := res.(OwnerResource)
	return ok
}
