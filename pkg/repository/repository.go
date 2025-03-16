package repository

import (
	"context"

	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
)

// Repository reprezentuje repozytorium dla zasobu
type Repository interface {
	// List pobiera listę elementów z opcjami zapytania
	List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error)

	// Get pobiera element po ID
	Get(ctx context.Context, id interface{}) (interface{}, error)

	// Create tworzy nowy element
	Create(ctx context.Context, data interface{}) (interface{}, error)

	// Update aktualizuje istniejący element
	Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error)

	// Delete usuwa element
	Delete(ctx context.Context, id interface{}) error
}

// RepositoryFactory tworzy repozytorium dla zasobu
type RepositoryFactory interface {
	// CreateRepository tworzy repozytorium dla zasobu
	CreateRepository(res resource.Resource) Repository
}
