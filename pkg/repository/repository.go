package repository

import (
	"context"

	"github.com/stanxing/refine-gin/pkg/query"
	"gorm.io/gorm"
)

// Repository defines the standard interface for database operations
type Repository interface {
	// Basic operations
	Get(ctx context.Context, id interface{}) (interface{}, error)
	List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error)
	Create(ctx context.Context, data interface{}) (interface{}, error)
	Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error)
	Delete(ctx context.Context, id interface{}) error

	// Bulk operations
	Count(ctx context.Context, options query.QueryOptions) (int64, error)
	CreateMany(ctx context.Context, data interface{}) (interface{}, error)
	UpdateMany(ctx context.Context, ids []interface{}, data interface{}) (int64, error)
	DeleteMany(ctx context.Context, ids []interface{}) (int64, error)

	// Retrieval with relations
	WithRelations(relations ...string) Repository
	GetWithRelations(ctx context.Context, id interface{}, relations []string) (interface{}, error)
	ListWithRelations(ctx context.Context, options query.QueryOptions, relations []string) (interface{}, int64, error)

	// Query helpers
	Query(ctx context.Context) *gorm.DB
	FindOneBy(ctx context.Context, condition map[string]interface{}) (interface{}, error)
	FindAllBy(ctx context.Context, condition map[string]interface{}) (interface{}, error)

	// Transaction support
	WithTransaction(fn func(Repository) error) error

	// GORM-specific operations
	BulkCreate(ctx context.Context, data interface{}) error
	BulkUpdate(ctx context.Context, condition map[string]interface{}, updates map[string]interface{}) error

	// Utilities
	GetIDFieldName() string
}
