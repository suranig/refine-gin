package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test model with owner field
type OwnerTestModel struct {
	ID      uint
	Name    string
	OwnerID string
}

// Test model without owner field
type NonOwnerTestModel struct {
	ID   uint
	Name string
}

func TestDefaultOwnerConfig(t *testing.T) {
	config := DefaultOwnerConfig()

	assert.Equal(t, "OwnerID", config.OwnerField)
	assert.True(t, config.EnforceOwnership)
	assert.Nil(t, config.DefaultOwnerID)
}

func TestNewOwnerResource(t *testing.T) {
	// Test with owner field
	t.Run("With valid owner field", func(t *testing.T) {
		res := NewResource(ResourceConfig{
			Name:  "owner-test",
			Model: &OwnerTestModel{},
		})

		config := DefaultOwnerConfig()
		ownerRes := NewOwnerResource(res, config)

		assert.NotNil(t, ownerRes)
		assert.Equal(t, res.GetName(), ownerRes.GetName())
		assert.Equal(t, config.OwnerField, ownerRes.GetOwnerField())
		assert.True(t, ownerRes.IsOwnershipEnforced())
		assert.Nil(t, ownerRes.GetDefaultOwnerID())
	})

	// Test with custom owner field
	t.Run("With custom owner field", func(t *testing.T) {
		res := NewResource(ResourceConfig{
			Name:  "owner-test",
			Model: &OwnerTestModel{},
		})

		config := OwnerConfig{
			OwnerField:       "OwnerID", // This exists in OwnerTestModel
			EnforceOwnership: false,
			DefaultOwnerID:   "default-owner",
		}
		ownerRes := NewOwnerResource(res, config)

		assert.NotNil(t, ownerRes)
		assert.Equal(t, "OwnerID", ownerRes.GetOwnerField())
		assert.False(t, ownerRes.IsOwnershipEnforced())
		assert.Equal(t, "default-owner", ownerRes.GetDefaultOwnerID())
	})

	// Test with invalid owner field should panic
	t.Run("With invalid owner field", func(t *testing.T) {
		res := NewResource(ResourceConfig{
			Name:  "owner-test",
			Model: &OwnerTestModel{},
		})

		config := OwnerConfig{
			OwnerField:       "MissingField", // This doesn't exist in OwnerTestModel
			EnforceOwnership: true,
		}

		assert.Panics(t, func() {
			NewOwnerResource(res, config)
		})
	})

	// Test with non-owner model
	t.Run("With non-owner model", func(t *testing.T) {
		res := NewResource(ResourceConfig{
			Name:  "non-owner-test",
			Model: &NonOwnerTestModel{},
		})

		config := OwnerConfig{
			OwnerField:       "OwnerID", // This doesn't exist in NonOwnerTestModel
			EnforceOwnership: true,
		}

		assert.Panics(t, func() {
			NewOwnerResource(res, config)
		})
	})

	// Test with empty owner field (should not panic)
	t.Run("With empty owner field", func(t *testing.T) {
		res := NewResource(ResourceConfig{
			Name:  "owner-test",
			Model: &OwnerTestModel{},
		})

		config := OwnerConfig{
			OwnerField:       "", // Empty, should not validate or panic
			EnforceOwnership: true,
		}

		assert.NotPanics(t, func() {
			NewOwnerResource(res, config)
		})
	})
}

func TestPromoteToOwnerResource(t *testing.T) {
	// Test promoting a regular resource
	t.Run("Promote regular resource", func(t *testing.T) {
		res := NewResource(ResourceConfig{
			Name:  "owner-test",
			Model: &OwnerTestModel{},
		})

		ownerRes := PromoteToOwnerResource(res)

		assert.NotNil(t, ownerRes)
		assert.Equal(t, res.GetName(), ownerRes.GetName())
		assert.Equal(t, "OwnerID", ownerRes.GetOwnerField())
		assert.True(t, ownerRes.IsOwnershipEnforced())
	})

	// Test promoting an already-owner resource
	t.Run("Promote owner resource", func(t *testing.T) {
		res := NewResource(ResourceConfig{
			Name:  "owner-test",
			Model: &OwnerTestModel{},
		})

		config := OwnerConfig{
			OwnerField:       "OwnerID",
			EnforceOwnership: false,
			DefaultOwnerID:   "custom-default",
		}

		originalOwnerRes := NewOwnerResource(res, config)
		promotedOwnerRes := PromoteToOwnerResource(originalOwnerRes)

		assert.Same(t, originalOwnerRes, promotedOwnerRes)
		assert.Equal(t, "OwnerID", promotedOwnerRes.GetOwnerField())
		assert.False(t, promotedOwnerRes.IsOwnershipEnforced())
		assert.Equal(t, "custom-default", promotedOwnerRes.GetDefaultOwnerID())
	})
}

func TestIsOwnerResource(t *testing.T) {
	// Test with owner resource
	t.Run("With owner resource", func(t *testing.T) {
		res := NewResource(ResourceConfig{
			Name:  "owner-test",
			Model: &OwnerTestModel{},
		})

		ownerRes := NewOwnerResource(res, DefaultOwnerConfig())
		assert.True(t, IsOwnerResource(ownerRes))
	})

	// Test with regular resource
	t.Run("With regular resource", func(t *testing.T) {
		res := NewResource(ResourceConfig{
			Name:  "regular-test",
			Model: &OwnerTestModel{},
		})

		assert.False(t, IsOwnerResource(res))
	})
}

func TestDefaultOwnerResource_Methods(t *testing.T) {
	res := NewResource(ResourceConfig{
		Name:  "owner-test",
		Model: &OwnerTestModel{},
	})

	config := OwnerConfig{
		OwnerField:       "OwnerID",
		EnforceOwnership: true,
		DefaultOwnerID:   "default-owner-123",
	}

	ownerRes := NewOwnerResource(res, config)

	// Test GetOwnerField
	assert.Equal(t, "OwnerID", ownerRes.GetOwnerField())

	// Test IsOwnershipEnforced
	assert.True(t, ownerRes.IsOwnershipEnforced())

	// Test GetDefaultOwnerID
	assert.Equal(t, "default-owner-123", ownerRes.GetDefaultOwnerID())

	// Test GetOwnerConfig
	assert.Equal(t, config, ownerRes.GetOwnerConfig())
}
