package resource

import (
	"github.com/suranig/refine-gin/pkg/naming"
)

// Options holds global configuration for resource
type Options struct {
	// Operations that are allowed for this resource
	AllowedOperations []Operation
	// QueryOptions for resource
	QueryOptions map[string]interface{}
	// NamingConvention for JSON fields
	NamingConvention naming.NamingConvention
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
