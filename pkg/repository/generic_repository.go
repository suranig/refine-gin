package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
)

// GenericRepository provides a complete implementation of the Repository interface
// with support for relations, transactions and all CRUD operations
type GenericRepository struct {
	DB       *gorm.DB
	Model    interface{}
	Resource resource.Resource
}

// NewGenericRepository creates a new GenericRepository instance
func NewGenericRepository(db *gorm.DB, modelOrResource interface{}) Repository {
	// Check if we're given a resource or just a model
	if res, ok := modelOrResource.(resource.Resource); ok {
		return &GenericRepository{
			DB:       db,
			Model:    res.GetModel(),
			Resource: res,
		}
	}

	// Otherwise, it's just a model
	return &GenericRepository{
		DB:    db,
		Model: modelOrResource,
	}
}

// NewGenericRepositoryWithResource creates a new GenericRepository instance with a resource
func NewGenericRepositoryWithResource(db *gorm.DB, res resource.Resource) Repository {
	return &GenericRepository{
		DB:       db,
		Model:    res.GetModel(),
		Resource: res,
	}
}

// List returns a paginated list of resources based on query options
func (r *GenericRepository) List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error) {
	// For models stored as pointers, return slice of pointers
	// For models stored as values, return slice of values
	modelType := reflect.TypeOf(r.Model)
	var elemType reflect.Type

	if modelType.Kind() == reflect.Ptr {
		// If model is a pointer, use the element type (*T -> T)
		elemType = modelType.Elem()
	} else {
		// If model is a value, use as is
		elemType = modelType
	}

	// Create a slice of the appropriate type
	sliceType := reflect.SliceOf(elemType)
	result := reflect.New(sliceType).Interface()

	tx := r.DB.WithContext(ctx)

	// Apply query options (filters, sorting, etc.)
	tx = options.Apply(tx)

	// Get total count before pagination
	var total int64
	if err := tx.Model(r.Model).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination if enabled
	if !options.DisablePagination {
		offset := (options.Page - 1) * options.PerPage
		tx = tx.Offset(offset).Limit(options.PerPage)
	}

	if err := tx.Find(result).Error; err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

// Get retrieves a single resource by its ID
func (r *GenericRepository) Get(ctx context.Context, id interface{}) (interface{}, error) {
	// Create a new instance of the model type
	modelType := reflect.TypeOf(r.Model)
	var result interface{}

	// If the model is already a pointer, we need to create a pointer to the element type
	if modelType.Kind() == reflect.Ptr {
		result = reflect.New(modelType.Elem()).Interface()
	} else {
		// If the model is not a pointer, we need to create a pointer to the model type
		result = reflect.New(modelType).Interface()
	}

	idFieldName := "id" // Default to "id"
	if r.Resource != nil {
		idFieldName = r.Resource.GetIDFieldName()
	}

	// Get the proper column name using GORM's naming strategy
	idColumnName := r.DB.NamingStrategy.ColumnName("", idFieldName)

	if err := r.DB.WithContext(ctx).Where(idColumnName+" = ?", id).First(result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

// Create inserts a new resource into the database
func (r *GenericRepository) Create(ctx context.Context, data interface{}) (interface{}, error) {
	if err := r.DB.WithContext(ctx).Create(data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

// Update modifies an existing resource identified by ID
func (r *GenericRepository) Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error) {
	// Try to set ID directly on model if it implements IDSetter
	TrySetID(data, id)

	// First, get the existing record
	existingRecord, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if data is a map or a struct
	isMap := false
	updateData := data
	if _, ok := data.(map[string]interface{}); ok {
		isMap = true
	}

	// If data is a map, we need to convert it to the model's struct type for Save to work properly
	if isMap {
		// Create a new instance of the model type to contain our updates
		modelType := reflect.TypeOf(r.Model)
		if modelType.Kind() == reflect.Ptr {
			// If the model is a pointer, create a new pointer to the element type
			modelValue := reflect.New(modelType.Elem())
			updateData = modelValue.Interface()
		} else {
			// If the model is not a pointer, create a new pointer to the model type
			modelValue := reflect.New(modelType)
			updateData = modelValue.Interface()
		}

		// Convert the existing record to JSON
		jsonData, err := json.Marshal(existingRecord)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal existing record: %w", err)
		}

		// Unmarshal the JSON into the new model instance to initialize it with existing values
		if err := json.Unmarshal(jsonData, updateData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal existing record: %w", err)
		}

		// Now apply the updates from the data map to the model
		updatesMap, _ := data.(map[string]interface{})
		updatesJSON, err := json.Marshal(updatesMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal updates: %w", err)
		}

		// This will only update the fields present in the JSON
		if err := json.Unmarshal(updatesJSON, updateData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal updates: %w", err)
		}
	}

	// Ensure ID field is set correctly in the update data
	if idSetter, ok := updateData.(IDSetter); ok {
		idSetter.SetID(id)
	}

	// Save the modified record - this will correctly handle JSON serialization
	if err := r.DB.WithContext(ctx).Save(updateData).Error; err != nil {
		return nil, err
	}

	// Return the updated record
	return r.Get(ctx, id)
}

// Delete removes a resource from the database
func (r *GenericRepository) Delete(ctx context.Context, id interface{}) error {
	tx := r.DB.WithContext(ctx)

	// If id is a map, use it directly as a condition
	if conditions, ok := id.(map[string]interface{}); ok {
		return tx.Delete(r.Model, conditions).Error
	}

	// Otherwise, use the ID field name
	idFieldName := "id" // Default to "id"
	if r.Resource != nil {
		idFieldName = r.Resource.GetIDFieldName()
	}

	// Get the proper column name using GORM's naming strategy
	idColumnName := r.DB.NamingStrategy.ColumnName("", idFieldName)

	return tx.Where(idColumnName+" = ?", id).Delete(r.Model).Error
}

// Count returns the total number of resources matching the query options
func (r *GenericRepository) Count(ctx context.Context, options query.QueryOptions) (int64, error) {
	var total int64
	tx := r.DB.WithContext(ctx).Model(r.Model)

	// Apply query options (filters only)
	tx = options.Apply(tx)

	if err := tx.Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

// CreateMany inserts multiple resources in a single transaction
func (r *GenericRepository) CreateMany(ctx context.Context, data interface{}) (interface{}, error) {
	// Check if data is a slice or a pointer to a slice
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Slice {
		return data, fmt.Errorf("unsupported data type: %v: Table not set, please set it like: db.Model(&user) or db.Table(\"users\")", data)
	}

	err := r.DB.WithContext(ctx).Create(data).Error
	if err != nil {
		// Return a nil slice of the same type as the input
		return reflect.Zero(val.Type()).Interface(), err
	}
	return data, nil
}

// UpdateMany modifies multiple resources in a single transaction
func (r *GenericRepository) UpdateMany(ctx context.Context, ids []interface{}, data interface{}) (int64, error) {
	idFieldName := "id" // Default to "id"
	if r.Resource != nil {
		idFieldName = r.Resource.GetIDFieldName()
	}

	// Get the proper column name using GORM's naming strategy
	idColumnName := r.DB.NamingStrategy.ColumnName("", idFieldName)

	result := r.DB.WithContext(ctx).Model(r.Model).Where(idColumnName+" IN ?", ids).Updates(data)
	return result.RowsAffected, result.Error
}

// DeleteMany removes multiple resources in a single transaction
func (r *GenericRepository) DeleteMany(ctx context.Context, ids []interface{}) (int64, error) {
	idFieldName := "id" // Default to "id"
	if r.Resource != nil {
		idFieldName = r.Resource.GetIDFieldName()
	}

	// Get the proper column name using GORM's naming strategy
	idColumnName := r.DB.NamingStrategy.ColumnName("", idFieldName)

	result := r.DB.WithContext(ctx).Where(idColumnName+" IN ?", ids).Delete(r.Model)
	return result.RowsAffected, result.Error
}

// WithTransaction executes operations within a database transaction
func (r *GenericRepository) WithTransaction(fn func(Repository) error) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		txRepo := &GenericRepository{
			DB:       tx,
			Model:    r.Model,
			Resource: r.Resource,
		}
		return fn(txRepo)
	})
}

// WithRelations specifies which relations should be loaded with the query
func (r *GenericRepository) WithRelations(relations ...string) Repository {
	newRepo := &GenericRepository{
		DB:       r.DB.Preload(strings.Join(relations, ".")),
		Model:    r.Model,
		Resource: r.Resource,
	}
	return newRepo
}

// BulkCreate creates multiple records at once
func (r *GenericRepository) BulkCreate(ctx context.Context, items interface{}) error {
	return r.DB.WithContext(ctx).Create(items).Error
}

// BulkUpdate updates multiple records at once based on a condition
func (r *GenericRepository) BulkUpdate(ctx context.Context, condition map[string]interface{}, updates map[string]interface{}) error {
	return r.DB.WithContext(ctx).Model(r.Model).Where(condition).Updates(updates).Error
}

// Query returns a query builder for custom queries
func (r *GenericRepository) Query(ctx context.Context) *gorm.DB {
	return r.DB.WithContext(ctx).Model(r.Model)
}

// GetWithRelations retrieves a single resource by its ID with related entities
func (r *GenericRepository) GetWithRelations(ctx context.Context, id interface{}, relations []string) (interface{}, error) {
	// Create a new instance of the model type
	modelType := reflect.TypeOf(r.Model)
	var result interface{}

	// If the model is already a pointer, we need to create a pointer to the element type
	if modelType.Kind() == reflect.Ptr {
		result = reflect.New(modelType.Elem()).Interface()
	} else {
		// If the model is not a pointer, we need to create a pointer to the model type
		result = reflect.New(modelType).Interface()
	}

	query := r.DB.WithContext(ctx)

	// Add preloads for all relations
	for _, relation := range relations {
		query = query.Preload(relation)
	}

	if err := query.First(result, id).Error; err != nil {
		return nil, err
	}

	return result, nil
}

// ListWithRelations retrieves a list of resources with related entities
func (r *GenericRepository) ListWithRelations(ctx context.Context, options query.QueryOptions, relations []string) (interface{}, int64, error) {
	// For models stored as pointers, return slice of pointers
	// For models stored as values, return slice of values
	modelType := reflect.TypeOf(r.Model)
	var elemType reflect.Type

	if modelType.Kind() == reflect.Ptr {
		// If model is a pointer, use the element type (*T -> T)
		elemType = modelType.Elem()
	} else {
		// If model is a value, use as is
		elemType = modelType
	}

	// Create a slice of the appropriate type
	sliceType := reflect.SliceOf(elemType)
	result := reflect.New(sliceType).Interface()

	tx := r.DB.WithContext(ctx)

	// Add preloads for all relations
	for _, relation := range relations {
		tx = tx.Preload(relation)
	}

	// Apply query options (filters, sorting, etc.)
	tx = options.Apply(tx)

	// Get total count before pagination
	var total int64
	if err := tx.Model(r.Model).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination if enabled
	if !options.DisablePagination {
		offset := (options.Page - 1) * options.PerPage
		tx = tx.Offset(offset).Limit(options.PerPage)
	}

	if err := tx.Find(result).Error; err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

// FindOneBy finds the first record that matches the condition
func (r *GenericRepository) FindOneBy(ctx context.Context, condition map[string]interface{}) (interface{}, error) {
	// Create a new instance of the model type
	modelType := reflect.TypeOf(r.Model)
	var result interface{}

	// If the model is already a pointer, we need to create a pointer to the element type
	if modelType.Kind() == reflect.Ptr {
		result = reflect.New(modelType.Elem()).Interface()
	} else {
		// If the model is not a pointer, we need to create a pointer to the model type
		result = reflect.New(modelType).Interface()
	}

	if err := r.DB.WithContext(ctx).Where(condition).First(result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

// FindAllBy finds all records that match the condition
func (r *GenericRepository) FindAllBy(ctx context.Context, condition map[string]interface{}) (interface{}, error) {
	// For models stored as pointers, return slice of pointers
	// For models stored as values, return slice of values
	modelType := reflect.TypeOf(r.Model)
	var elemType reflect.Type

	if modelType.Kind() == reflect.Ptr {
		// If model is a pointer, use the element type (*T -> T)
		elemType = modelType.Elem()
	} else {
		// If model is a value, use as is
		elemType = modelType
	}

	// Create a slice of the appropriate type
	sliceType := reflect.SliceOf(elemType)
	result := reflect.New(sliceType).Interface()

	if err := r.DB.WithContext(ctx).Where(condition).Find(result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

// GetIDFieldName returns the field name used as primary key
func (r *GenericRepository) GetIDFieldName() string {
	// If resource is set and has a custom ID field name, use it
	if r.Resource != nil {
		idFieldName := r.Resource.GetIDFieldName()
		if idFieldName != "" {
			return idFieldName
		}
	}
	// Default to "id"
	return "id"
}
