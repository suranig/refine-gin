package repository

import (
	"context"

	"github.com/suranig/refine-gin/pkg/query"
	"gorm.io/gorm"
)

// Repository defines the interface for database operations
type Repository interface {
	// List returns a paginated list of resources based on query options
	List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error)

	// Get retrieves a single resource by its ID
	Get(ctx context.Context, id interface{}) (interface{}, error)

	// Create inserts a new resource into the database
	Create(ctx context.Context, data interface{}) (interface{}, error)

	// Update modifies an existing resource identified by ID
	Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error)

	// Delete removes a resource from the database
	Delete(ctx context.Context, id interface{}) error

	// Count returns the total number of resources matching the query options
	Count(ctx context.Context, options query.QueryOptions) (int64, error)

	// CreateMany inserts multiple resources in a single transaction
	CreateMany(ctx context.Context, data interface{}) (interface{}, error)

	// UpdateMany modifies multiple resources in a single transaction
	UpdateMany(ctx context.Context, ids []interface{}, data interface{}) (int64, error)

	// DeleteMany removes multiple resources in a single transaction
	DeleteMany(ctx context.Context, ids []interface{}) (int64, error)

	// WithTransaction executes operations within a database transaction
	WithTransaction(fn func(Repository) error) error

	// WithRelations specifies which relations should be loaded with the query
	WithRelations(relations ...string) Repository

	// FindOneBy finds the first record that matches the condition
	FindOneBy(ctx context.Context, condition map[string]interface{}) (interface{}, error)

	// FindAllBy finds all records that match the condition
	FindAllBy(ctx context.Context, condition map[string]interface{}) (interface{}, error)

	// GetWithRelations retrieves a single resource with specified relations loaded
	GetWithRelations(ctx context.Context, id interface{}, relations []string) (interface{}, error)

	// ListWithRelations lists resources with specified relations loaded
	ListWithRelations(ctx context.Context, options query.QueryOptions, relations []string) (interface{}, int64, error)

	// Query returns a GORM DB instance for custom queries
	Query(ctx context.Context) *gorm.DB

	// BulkCreate creates multiple resources at once
	BulkCreate(ctx context.Context, data interface{}) error

	// BulkUpdate updates multiple resources matching a condition
	BulkUpdate(ctx context.Context, condition map[string]interface{}, updates map[string]interface{}) error
}
