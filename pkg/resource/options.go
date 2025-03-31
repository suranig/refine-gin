package resource

import (
	"github.com/suranig/refine-gin/pkg/naming"
)

// CacheOptions holds configuration for resource caching
type CacheOptions struct {
	// Enabled określa czy cachowanie jest włączone
	Enabled bool
	// MaxAge określa czas ważności cache w sekundach
	MaxAge int
	// VaryHeaders określa nagłówki, które wpływają na cache
	VaryHeaders []string
}

// DefaultCacheOptions zwraca domyślne opcje cachowania
func DefaultCacheOptions() CacheOptions {
	return CacheOptions{
		Enabled:     true,
		MaxAge:      60, // 1 minuta
		VaryHeaders: []string{"Accept", "Accept-Encoding", "Authorization"},
	}
}

// Options holds global configuration for resource
type Options struct {
	// Operations that are allowed for this resource
	AllowedOperations []Operation
	// QueryOptions for resource
	QueryOptions map[string]interface{}
	// NamingConvention for JSON fields
	NamingConvention naming.NamingConvention
	// Cache options for resource
	Cache CacheOptions
}

// DefaultOptions returns default options
func DefaultOptions() Options {
	return Options{
		AllowedOperations: []Operation{
			OperationCreate,
			OperationList,
			OperationRead,
			OperationUpdate,
			OperationDelete,
			OperationCount,
		},
		QueryOptions:     make(map[string]interface{}),
		NamingConvention: naming.SnakeCase,
		Cache:            DefaultCacheOptions(),
	}
}

// WithOperation adds operation to allowed operations
func (o Options) WithOperation(op Operation) Options {
	o.AllowedOperations = append(o.AllowedOperations, op)
	return o
}

// WithOperations sets allowed operations
func (o Options) WithOperations(ops []Operation) Options {
	o.AllowedOperations = ops
	return o
}

// WithQueryOption adds query option
func (o Options) WithQueryOption(key string, value interface{}) Options {
	o.QueryOptions[key] = value
	return o
}

// WithNamingConvention sets naming convention for JSON fields
func (o Options) WithNamingConvention(convention naming.NamingConvention) Options {
	o.NamingConvention = convention
	return o
}

// WithCache sets cache options
func (o Options) WithCache(options CacheOptions) Options {
	o.Cache = options
	return o
}

// WithCacheEnabled enables or disables caching
func (o Options) WithCacheEnabled(enabled bool) Options {
	o.Cache.Enabled = enabled
	return o
}

// WithCacheMaxAge sets maximum age for cache in seconds
func (o Options) WithCacheMaxAge(seconds int) Options {
	o.Cache.MaxAge = seconds
	return o
}

// WithCacheVaryHeaders sets headers that affect caching
func (o Options) WithCacheVaryHeaders(headers []string) Options {
	o.Cache.VaryHeaders = headers
	return o
}

// GetQueryOption returns the value of a query option, or nil if not set
func (o Options) GetQueryOption(key string) interface{} {
	if value, exists := o.QueryOptions[key]; exists {
		return value
	}
	return nil
}
