package repository

import (
	"context"

	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
)

// GenericRepository is an extended version of GormRepository with additional functionality
type GenericRepository struct {
	GormRepository
}

// NewGenericRepository creates a new generic repository based on a resource
func NewGenericRepository(db *gorm.DB, res resource.Resource) *GenericRepository {
	gormRepo := &GormRepository{
		DB:          db,
		Model:       resource.CreateInstanceOfType(res.GetModel()),
		Resource:    res,
		IDFieldName: res.GetIDFieldName(),
	}

	return &GenericRepository{
		GormRepository: *gormRepo,
	}
}

// ListWithRelations retrieves a list of resources with relations included
func (r *GenericRepository) ListWithRelations(ctx context.Context, options query.QueryOptions, relations []string) (interface{}, int64, error) {
	slicePtr := resource.CreateSliceOfType(r.Model)

	// Prepare query
	query := r.DB.Model(r.Model)

	// Load relations
	for _, relation := range relations {
		query = query.Preload(relation)
	}

	// Apply query options
	total, err := options.ApplyWithPagination(query, slicePtr)
	if err != nil {
		return nil, 0, err
	}

	return slicePtr, total, nil
}

// GetWithRelations retrieves a single resource with relations included
func (r *GenericRepository) GetWithRelations(ctx context.Context, id interface{}, relations []string) (interface{}, error) {
	item := resource.CreateInstanceOfType(r.Model)

	// Prepare query
	query := r.DB.Model(r.Model)

	// Load relations
	for _, relation := range relations {
		query = query.Preload(relation)
	}

	// Execute query
	if err := query.First(item, r.IDFieldName+" = ?", id).Error; err != nil {
		return nil, err
	}

	return item, nil
}

// WithTransaction executes operations in a transaction
func (r *GenericRepository) WithTransaction(callback func(repo Repository) error) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// Create a copy of the repository with transactional DB
		txRepo := &GenericRepository{
			GormRepository: GormRepository{
				DB:          tx,
				Model:       r.Model,
				Resource:    r.Resource,
				IDFieldName: r.IDFieldName,
			},
		}
		return callback(txRepo)
	})
}

// GenericRepositoryFactory is a factory for creating GenericRepository instances
type GenericRepositoryFactory struct {
	DB *gorm.DB
}

// CreateRepository creates a new generic repository for a resource
func (f *GenericRepositoryFactory) CreateRepository(res resource.Resource) Repository {
	return NewGenericRepository(f.DB, res)
}

// NewGenericRepositoryFactory creates a new factory for generic repositories
func NewGenericRepositoryFactory(db *gorm.DB) RepositoryFactory {
	return &GenericRepositoryFactory{
		DB: db,
	}
}

// BulkCreate creates multiple records at once
func (r *GenericRepository) BulkCreate(ctx context.Context, items interface{}) error {
	return r.DB.Create(items).Error
}

// BulkUpdate updates multiple records at once based on a condition
func (r *GenericRepository) BulkUpdate(ctx context.Context, condition map[string]interface{}, updates map[string]interface{}) error {
	return r.DB.Model(r.Model).Where(condition).Updates(updates).Error
}

// Query returns a query builder for custom queries
func (r *GenericRepository) Query(ctx context.Context) *gorm.DB {
	return r.DB.Model(r.Model)
}

// FindOneBy finds the first record that matches the condition
func (r *GenericRepository) FindOneBy(ctx context.Context, condition map[string]interface{}) (interface{}, error) {
	item := resource.CreateInstanceOfType(r.Model)

	if err := r.DB.Where(condition).First(item).Error; err != nil {
		return nil, err
	}

	return item, nil
}

// FindAllBy finds all records that match the condition
func (r *GenericRepository) FindAllBy(ctx context.Context, condition map[string]interface{}) (interface{}, error) {
	slicePtr := resource.CreateSliceOfType(r.Model)

	if err := r.DB.Where(condition).Find(slicePtr).Error; err != nil {
		return nil, err
	}

	return slicePtr, nil
}

// CreateMany creates multiple resources at once, implementing the Repository interface
func (r *GenericRepository) CreateMany(ctx context.Context, data interface{}) (interface{}, error) {
	// Check if the data is a slice
	if !resource.IsSlice(data) {
		return nil, resource.ErrInvalidType
	}

	// Begin transaction
	tx := r.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create resources in the database
	if err := tx.Create(data).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return data, nil
}

// UpdateMany updates multiple resources at once, implementing the Repository interface
func (r *GenericRepository) UpdateMany(ctx context.Context, ids []interface{}, data interface{}) (int64, error) {
	// Begin transaction
	tx := r.DB.Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update resources in the database
	result := tx.Model(r.Model).Where(r.IDFieldName+" IN ?", ids).Updates(data)
	if result.Error != nil {
		tx.Rollback()
		return 0, result.Error
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return result.RowsAffected, nil
}

// DeleteMany deletes multiple resources at once, implementing the Repository interface
func (r *GenericRepository) DeleteMany(ctx context.Context, ids []interface{}) (int64, error) {
	// Begin transaction
	tx := r.DB.Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete resources from the database
	result := tx.Where(r.IDFieldName+" IN ?", ids).Delete(r.Model)
	if result.Error != nil {
		tx.Rollback()
		return 0, result.Error
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return result.RowsAffected, nil
}
