package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/suranig/refine-gin/pkg/naming"
)

func TestDefaultCacheOptions(t *testing.T) {
	options := DefaultCacheOptions()

	// Verify default values
	assert.True(t, options.Enabled, "Default cache should be enabled")
	assert.Equal(t, 60, options.MaxAge, "Default max age should be 60 seconds")
	assert.Equal(t, []string{"Accept", "Accept-Encoding", "Authorization"}, options.VaryHeaders, "Default vary headers should match")
}

func TestDefaultOptions(t *testing.T) {
	options := DefaultOptions()

	// Verify default operations
	assert.Contains(t, options.AllowedOperations, OperationCreate)
	assert.Contains(t, options.AllowedOperations, OperationList)
	assert.Contains(t, options.AllowedOperations, OperationRead)
	assert.Contains(t, options.AllowedOperations, OperationUpdate)
	assert.Contains(t, options.AllowedOperations, OperationDelete)
	assert.Contains(t, options.AllowedOperations, OperationCount)
	assert.Len(t, options.AllowedOperations, 6, "Should have 6 default operations")

	// Verify other default values
	assert.Empty(t, options.QueryOptions, "QueryOptions should be initialized as empty map")
	assert.Equal(t, naming.SnakeCase, options.NamingConvention, "Default naming convention should be snake_case")

	// Verify default cache options
	assert.True(t, options.Cache.Enabled, "Default cache should be enabled")
	assert.Equal(t, 60, options.Cache.MaxAge, "Default max age should be 60 seconds")
	assert.Equal(t, []string{"Accept", "Accept-Encoding", "Authorization"}, options.Cache.VaryHeaders, "Default vary headers should match")
}

func TestWithOperation(t *testing.T) {
	// Start with empty options
	options := Options{
		AllowedOperations: []Operation{},
	}

	// Add a single operation
	newOptions := options.WithOperation(OperationCreate)

	// Verify the operation was added
	assert.Len(t, newOptions.AllowedOperations, 1)
	assert.Contains(t, newOptions.AllowedOperations, OperationCreate)

	// Original options should be unchanged (fluent interface)
	assert.Empty(t, options.AllowedOperations)
}

func TestWithOperations(t *testing.T) {
	// Start with some initial operations
	options := Options{
		AllowedOperations: []Operation{OperationCreate},
	}

	// Replace with new set of operations
	operations := []Operation{OperationRead, OperationList}
	newOptions := options.WithOperations(operations)

	// Verify operations were replaced
	assert.Len(t, newOptions.AllowedOperations, 2)
	assert.Contains(t, newOptions.AllowedOperations, OperationRead)
	assert.Contains(t, newOptions.AllowedOperations, OperationList)
	assert.NotContains(t, newOptions.AllowedOperations, OperationCreate)

	// Original options should be unchanged
	assert.Len(t, options.AllowedOperations, 1)
	assert.Contains(t, options.AllowedOperations, OperationCreate)
}

func TestWithQueryOption(t *testing.T) {
	// Start with empty options
	options := Options{
		QueryOptions: make(map[string]interface{}),
	}

	// Add a query option
	newOptions := options.WithQueryOption("maxLimit", 100)

	// Verify the option was added
	assert.Len(t, newOptions.QueryOptions, 1)
	assert.Equal(t, 100, newOptions.QueryOptions["maxLimit"])

	// Add another query option
	newerOptions := newOptions.WithQueryOption("defaultSort", "createdAt")

	// Verify both options exist
	assert.Len(t, newerOptions.QueryOptions, 2)
	assert.Equal(t, 100, newerOptions.QueryOptions["maxLimit"])
	assert.Equal(t, "createdAt", newerOptions.QueryOptions["defaultSort"])
}

func TestWithNamingConvention(t *testing.T) {
	// Start with default snake_case
	options := DefaultOptions()
	assert.Equal(t, naming.SnakeCase, options.NamingConvention)

	// Change to camelCase
	newOptions := options.WithNamingConvention(naming.CamelCase)

	// Verify the naming convention was changed
	assert.Equal(t, naming.CamelCase, newOptions.NamingConvention)

	// Original should be unchanged
	assert.Equal(t, naming.SnakeCase, options.NamingConvention)
}

func TestWithCache(t *testing.T) {
	// Start with default options
	options := DefaultOptions()

	// Create custom cache options
	customCache := CacheOptions{
		Enabled:     false,
		MaxAge:      300,
		VaryHeaders: []string{"X-Custom-Header"},
	}

	// Apply custom cache options
	newOptions := options.WithCache(customCache)

	// Verify the cache options were applied
	assert.False(t, newOptions.Cache.Enabled)
	assert.Equal(t, 300, newOptions.Cache.MaxAge)
	assert.Equal(t, []string{"X-Custom-Header"}, newOptions.Cache.VaryHeaders)

	// Original should be unchanged
	assert.True(t, options.Cache.Enabled)
	assert.Equal(t, 60, options.Cache.MaxAge)
	assert.Equal(t, []string{"Accept", "Accept-Encoding", "Authorization"}, options.Cache.VaryHeaders)
}

func TestWithCacheEnabled(t *testing.T) {
	// Start with default options (cache enabled)
	options := DefaultOptions()
	assert.True(t, options.Cache.Enabled)

	// Disable cache
	newOptions := options.WithCacheEnabled(false)

	// Verify cache was disabled
	assert.False(t, newOptions.Cache.Enabled)

	// Original should be unchanged
	assert.True(t, options.Cache.Enabled)
}

func TestWithCacheMaxAge(t *testing.T) {
	// Start with default options
	options := DefaultOptions()
	assert.Equal(t, 60, options.Cache.MaxAge)

	// Change max age
	newOptions := options.WithCacheMaxAge(3600) // 1 hour

	// Verify max age was changed
	assert.Equal(t, 3600, newOptions.Cache.MaxAge)

	// Original should be unchanged
	assert.Equal(t, 60, options.Cache.MaxAge)
}

func TestWithCacheVaryHeaders(t *testing.T) {
	// Start with default options
	options := DefaultOptions()
	assert.Equal(t, []string{"Accept", "Accept-Encoding", "Authorization"}, options.Cache.VaryHeaders)

	// Change vary headers
	newHeaders := []string{"X-Custom-Header", "X-User-ID"}
	newOptions := options.WithCacheVaryHeaders(newHeaders)

	// Verify headers were changed
	assert.Equal(t, newHeaders, newOptions.Cache.VaryHeaders)

	// Original should be unchanged
	assert.Equal(t, []string{"Accept", "Accept-Encoding", "Authorization"}, options.Cache.VaryHeaders)
}

func TestOptionsChaining(t *testing.T) {
	// Test chaining multiple option methods
	options := DefaultOptions()

	// Apply multiple changes in a chain
	result := options.
		WithOperation(OperationCustom).
		WithNamingConvention(naming.CamelCase).
		WithCacheMaxAge(7200).
		WithCacheEnabled(false).
		WithQueryOption("defaultLimit", 50)

	// Verify all changes were applied
	assert.Contains(t, result.AllowedOperations, OperationCustom)
	assert.Equal(t, naming.CamelCase, result.NamingConvention)
	assert.Equal(t, 7200, result.Cache.MaxAge)
	assert.False(t, result.Cache.Enabled)
	assert.Equal(t, 50, result.QueryOptions["defaultLimit"])
}

func TestGetQueryOption(t *testing.T) {
	opts := Options{
		QueryOptions: map[string]interface{}{
			"limit": 10,
		},
	}

	assert.Equal(t, 10, opts.GetQueryOption("limit"))
	assert.Nil(t, opts.GetQueryOption("missing"))
}
