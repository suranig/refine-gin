package repository

import (
	"context"

	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
)

// Repository defines a generic repository interface
type Repository interface {
	// List returns a list of resources with pagination
	List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error)

	// Get returns a single resource by ID
	Get(ctx context.Context, id interface{}) (interface{}, error)

	// Create creates a new resource
	Create(ctx context.Context, data interface{}) (interface{}, error)

	// Update updates an existing resource
	Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error)

	// Delete deletes a resource
	Delete(ctx context.Context, id interface{}) error

	// Count returns the total number of resources matching the query options
	Count(ctx context.Context, options query.QueryOptions) (int64, error)
}

type RepositoryFactory interface {
	CreateRepository(res resource.Resource) Repository
}

type GormRepository struct {
	DB          *gorm.DB
	Model       interface{}
	Resource    resource.Resource
	IDFieldName string // Nazwa pola identyfikatora
}

func (r *GormRepository) List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error) {
	slicePtr := resource.CreateSliceOfType(r.Model)

	total, err := options.ApplyWithPagination(r.DB.Model(r.Model), slicePtr)
	if err != nil {
		return nil, 0, err
	}

	return slicePtr, total, nil
}

func (r *GormRepository) Get(ctx context.Context, id interface{}) (interface{}, error) {
	item := resource.CreateInstanceOfType(r.Model)

	if err := r.DB.First(item, r.IDFieldName+" = ?", id).Error; err != nil {
		return nil, err
	}

	return item, nil
}

func (r *GormRepository) Create(ctx context.Context, data interface{}) (interface{}, error) {
	if err := r.DB.Create(data).Error; err != nil {
		return nil, err
	}

	return data, nil
}

func (r *GormRepository) Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error) {
	// Ustaw ID w danych, jeśli to możliwe
	if err := resource.SetCustomID(data, id, r.IDFieldName); err != nil {
		return nil, err
	}

	if err := r.DB.Save(data).Error; err != nil {
		return nil, err
	}

	return data, nil
}

func (r *GormRepository) Delete(ctx context.Context, id interface{}) error {
	item := resource.CreateInstanceOfType(r.Model)

	if err := resource.SetCustomID(item, id, r.IDFieldName); err != nil {
		return err
	}

	return r.DB.Delete(item).Error
}

func (r *GormRepository) Count(ctx context.Context, options query.QueryOptions) (int64, error) {
	total, err := options.ApplyWithPagination(r.DB.Model(r.Model), nil)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func NewGormRepository(db *gorm.DB, model interface{}) Repository {
	return &GormRepository{
		DB:          db,
		Model:       model,
		IDFieldName: "ID", // Domyślna nazwa pola identyfikatora
	}
}

// NewGormRepositoryWithResource tworzy nowe repozytorium GORM z zasobem
func NewGormRepositoryWithResource(db *gorm.DB, res resource.Resource) Repository {
	return &GormRepository{
		DB:          db,
		Model:       res.GetModel(),
		Resource:    res,
		IDFieldName: res.GetIDFieldName(),
	}
}
