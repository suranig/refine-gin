package repository

import (
	"context"

	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
)

type Repository interface {
	List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error)

	Get(ctx context.Context, id interface{}) (interface{}, error)

	Create(ctx context.Context, data interface{}) (interface{}, error)

	Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error)

	Delete(ctx context.Context, id interface{}) error
}

type RepositoryFactory interface {
	CreateRepository(res resource.Resource) Repository
}

type GormRepository struct {
	DB    *gorm.DB
	Model interface{}
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

	if err := r.DB.First(item, "id = ?", id).Error; err != nil {
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
	if err := resource.SetID(data, id); err != nil {
		return nil, err
	}

	if err := r.DB.Save(data).Error; err != nil {
		return nil, err
	}

	return data, nil
}

func (r *GormRepository) Delete(ctx context.Context, id interface{}) error {
	item := resource.CreateInstanceOfType(r.Model)

	if err := resource.SetID(item, id); err != nil {
		return err
	}

	return r.DB.Delete(item).Error
}

func NewGormRepository(db *gorm.DB, model interface{}) Repository {
	return &GormRepository{
		DB:    db,
		Model: model,
	}
}
