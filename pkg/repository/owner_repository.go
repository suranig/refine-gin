package repository

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/suranig/refine-gin/pkg/middleware"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
)

// Error constants for owner repository
var (
	ErrOwnerMismatch    = errors.New("owner ID mismatch")
	ErrOwnerIDNotFound  = errors.New("owner ID not found in context")
	ErrNoOwnerResource  = errors.New("not an owner resource")
	ErrNoOwnerFieldName = errors.New("owner field name not specified")
)

// OwnerRepository extends the Repository interface with ownership functionality
type OwnerRepository interface {
	Repository
}

// OwnerGenericRepository provides a complete implementation of the OwnerRepository interface
type OwnerGenericRepository struct {
	GenericRepository
	Resource resource.OwnerResource
}

// NewOwnerRepository creates a new OwnerGenericRepository instance
func NewOwnerRepository(db *gorm.DB, res resource.Resource) (Repository, error) {
	// Check if resource is an owner resource or can be promoted
	var ownerRes resource.OwnerResource
	var ok bool

	if ownerRes, ok = res.(resource.OwnerResource); !ok {
		// Try to promote to owner resource
		ownerRes = resource.PromoteToOwnerResource(res)
	}

	// Check if owner field is specified
	if ownerRes.GetOwnerField() == "" {
		return nil, ErrNoOwnerFieldName
	}

	return &OwnerGenericRepository{
		GenericRepository: GenericRepository{
			DB:       db,
			Model:    ownerRes.GetModel(),
			Resource: ownerRes,
		},
		Resource: ownerRes,
	}, nil
}

// extractOwnerID gets the owner ID from context
func (r *OwnerGenericRepository) extractOwnerID(ctx context.Context) (interface{}, error) {
	if !r.Resource.IsOwnershipEnforced() {
		// If enforcement is disabled, use default owner ID if provided
		if r.Resource.GetDefaultOwnerID() != nil {
			return r.Resource.GetDefaultOwnerID(), nil
		}
		return nil, nil
	}

	// Extract owner ID from context
	ownerID, err := middleware.GetOwnerID(ctx)
	if err != nil {
		// If no owner ID is found in context, use default if provided
		if r.Resource.GetDefaultOwnerID() != nil {
			return r.Resource.GetDefaultOwnerID(), nil
		}
		return nil, ErrOwnerIDNotFound
	}

	return ownerID, nil
}

// applyOwnerFilter adds owner filtering to a query
func (r *OwnerGenericRepository) applyOwnerFilter(ctx context.Context, tx *gorm.DB) (*gorm.DB, error) {
	// If ownership is not enforced, return original query unmodified
	if !r.Resource.IsOwnershipEnforced() {
		return tx, nil
	}

	// Extract owner ID from context
	ownerID, err := r.extractOwnerID(ctx)
	if err != nil {
		return nil, err
	}

	// If no owner ID present, return original query (no filtering)
	if ownerID == nil {
		return tx, nil
	}

	// Get the owner field and convert to column name
	ownerField := r.Resource.GetOwnerField()
	columnName := tx.Config.NamingStrategy.ColumnName("", ownerField)

	// Apply the owner filter
	return tx.Where(columnName+" = ?", ownerID), nil
}

// verifyOwnership checks if the user owns a record
func (r *OwnerGenericRepository) verifyOwnership(ctx context.Context, id interface{}) error {
	if !r.Resource.IsOwnershipEnforced() {
		return nil
	}

	ownerID, err := r.extractOwnerID(ctx)
	if err != nil {
		return err
	}

	// If no owner ID, skip verification
	if ownerID == nil {
		return nil
	}

	// First check if the record exists
	record, err := r.GenericRepository.Get(ctx, id)
	if err != nil {
		return err // Record not found or other error
	}

	// Get owner field value from record
	recordValue := reflect.ValueOf(record)
	if recordValue.Kind() == reflect.Ptr {
		recordValue = recordValue.Elem()
	}

	ownerField := r.Resource.GetOwnerField()
	field := recordValue.FieldByName(ownerField)
	if !field.IsValid() {
		return fmt.Errorf("owner field '%s' not found in record", ownerField)
	}

	// Compare owner IDs
	// Convert both to strings for comparison (simple approach)
	recordOwnerID := fmt.Sprintf("%v", field.Interface())
	contextOwnerID := fmt.Sprintf("%v", ownerID)

	if recordOwnerID != contextOwnerID {
		return ErrOwnerMismatch
	}

	return nil
}

// setOwnership sets owner field on new records
func (r *OwnerGenericRepository) setOwnership(ctx context.Context, data interface{}) error {
	if !r.Resource.IsOwnershipEnforced() {
		return nil
	}

	ownerID, err := r.extractOwnerID(ctx)
	if err != nil {
		return err
	}

	// If no owner ID, skip setting
	if ownerID == nil {
		return nil
	}

	// Set owner field on record
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() == reflect.Ptr {
		dataValue = dataValue.Elem()
	}

	// Handle slice of records
	if dataValue.Kind() == reflect.Slice {
		for i := 0; i < dataValue.Len(); i++ {
			item := dataValue.Index(i)
			if item.Kind() == reflect.Ptr {
				item = item.Elem()
			}

			// Set owner field
			ownerField := r.Resource.GetOwnerField()
			field := item.FieldByName(ownerField)
			if !field.IsValid() {
				return fmt.Errorf("owner field '%s' not found in record", ownerField)
			}

			if field.CanSet() {
				// Convert owner ID to the field's type
				converted, err := convertToFieldType(ownerID, field.Type())
				if err != nil {
					return err
				}
				field.Set(reflect.ValueOf(converted))
			}
		}
		return nil
	}

	// Handle single record
	ownerField := r.Resource.GetOwnerField()
	field := dataValue.FieldByName(ownerField)
	if !field.IsValid() {
		return fmt.Errorf("owner field '%s' not found in record", ownerField)
	}

	if field.CanSet() {
		// Convert owner ID to the field's type
		converted, err := convertToFieldType(ownerID, field.Type())
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(converted))
	}

	return nil
}

// convertToFieldType converts a value to the target field type
func convertToFieldType(value interface{}, targetType reflect.Type) (interface{}, error) {
	valueType := reflect.TypeOf(value)

	// If types are the same, return as is
	if valueType == targetType {
		return value, nil
	}

	// Handle string conversions
	valueStr := fmt.Sprintf("%v", value)
	switch targetType.Kind() {
	case reflect.String:
		return valueStr, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var result int64
		_, err := fmt.Sscanf(valueStr, "%d", &result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert %v to int: %w", value, err)
		}
		return reflect.ValueOf(result).Convert(targetType).Interface(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var result uint64
		_, err := fmt.Sscanf(valueStr, "%d", &result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert %v to uint: %w", value, err)
		}
		return reflect.ValueOf(result).Convert(targetType).Interface(), nil
	case reflect.Float32, reflect.Float64:
		var result float64
		_, err := fmt.Sscanf(valueStr, "%f", &result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert %v to float: %w", value, err)
		}
		return reflect.ValueOf(result).Convert(targetType).Interface(), nil
	case reflect.Bool:
		var result bool
		_, err := fmt.Sscanf(valueStr, "%t", &result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert %v to bool: %w", value, err)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported conversion from %v to %v", valueType, targetType)
	}
}

// Override the Repository interface methods to add ownership checks

// List returns a paginated list of resources filtered by owner
func (r *OwnerGenericRepository) List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error) {
	// Apply owner filter to DB
	tx := r.DB.WithContext(ctx)
	var err error
	tx, err = r.applyOwnerFilter(ctx, tx)
	if err != nil {
		return nil, 0, err
	}

	// Use GenericRepository to complete the operation with modified query
	r.DB = tx
	result, total, err := r.GenericRepository.List(ctx, options)

	// Reset DB to original
	r.DB = r.GenericRepository.DB

	return result, total, err
}

// Get retrieves a single resource and verifies ownership
func (r *OwnerGenericRepository) Get(ctx context.Context, id interface{}) (interface{}, error) {
	result, err := r.GenericRepository.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if err := r.verifyOwnership(ctx, id); err != nil {
		return nil, err
	}

	return result, nil
}

// Create inserts a new resource and sets ownership
func (r *OwnerGenericRepository) Create(ctx context.Context, data interface{}) (interface{}, error) {
	// Set owner field
	if err := r.setOwnership(ctx, data); err != nil {
		return nil, err
	}

	return r.GenericRepository.Create(ctx, data)
}

// Update modifies an existing resource after verifying ownership
func (r *OwnerGenericRepository) Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error) {
	// Verify ownership
	if err := r.verifyOwnership(ctx, id); err != nil {
		return nil, err
	}

	// We apply the owner filter directly to the model being modified
	return r.GenericRepository.Update(ctx, id, data)
}

// Delete removes a resource after verifying ownership
func (r *OwnerGenericRepository) Delete(ctx context.Context, id interface{}) error {
	// Verify ownership
	if err := r.verifyOwnership(ctx, id); err != nil {
		return err
	}

	// We apply owner verification directly so no need to filter query
	return r.GenericRepository.Delete(ctx, id)
}

// Count returns the total number of resources filtered by owner
func (r *OwnerGenericRepository) Count(ctx context.Context, options query.QueryOptions) (int64, error) {
	// Apply owner filter to DB
	tx := r.DB.WithContext(ctx)
	var err error
	tx, err = r.applyOwnerFilter(ctx, tx)
	if err != nil {
		return 0, err
	}

	// Use GenericRepository to complete the operation with modified query
	r.DB = tx
	total, err := r.GenericRepository.Count(ctx, options)

	// Reset DB to original
	r.DB = r.GenericRepository.DB

	return total, err
}

// CreateMany inserts multiple resources and sets ownership on all
func (r *OwnerGenericRepository) CreateMany(ctx context.Context, data interface{}) (interface{}, error) {
	// Set owner field on all records
	if err := r.setOwnership(ctx, data); err != nil {
		return nil, err
	}

	return r.GenericRepository.CreateMany(ctx, data)
}

// UpdateMany modifies multiple resources after verifying ownership for all
func (r *OwnerGenericRepository) UpdateMany(ctx context.Context, ids []interface{}, data interface{}) (int64, error) {
	// Verify ownership for each ID
	for _, id := range ids {
		if err := r.verifyOwnership(ctx, id); err != nil {
			return 0, err
		}
	}

	return r.GenericRepository.UpdateMany(ctx, ids, data)
}

// DeleteMany removes multiple resources after verifying ownership for all
func (r *OwnerGenericRepository) DeleteMany(ctx context.Context, ids []interface{}) (int64, error) {
	// Verify ownership for each ID
	for _, id := range ids {
		if err := r.verifyOwnership(ctx, id); err != nil {
			return 0, err
		}
	}

	return r.GenericRepository.DeleteMany(ctx, ids)
}

// We delegate these methods to the GenericRepository, as they don't require ownership checks
func (r *OwnerGenericRepository) WithTransaction(fn func(Repository) error) error {
	return r.GenericRepository.WithTransaction(fn)
}

func (r *OwnerGenericRepository) WithRelations(relations ...string) Repository {
	return &OwnerGenericRepository{
		GenericRepository: *r.GenericRepository.WithRelations(relations...).(*GenericRepository),
		Resource:          r.Resource,
	}
}

func (r *OwnerGenericRepository) FindOneBy(ctx context.Context, condition map[string]interface{}) (interface{}, error) {
	// Add owner condition if enforced
	var err error
	ownerID, err := r.extractOwnerID(ctx)
	if err != nil {
		return nil, err
	}

	// Add owner condition if enforced
	if ownerID != nil && r.Resource.IsOwnershipEnforced() {
		condition[r.Resource.GetOwnerField()] = ownerID
	}

	return r.GenericRepository.FindOneBy(ctx, condition)
}

func (r *OwnerGenericRepository) FindAllBy(ctx context.Context, condition map[string]interface{}) (interface{}, error) {
	// Add owner condition if enforced
	var err error
	ownerID, err := r.extractOwnerID(ctx)
	if err != nil {
		return nil, err
	}

	// Add owner condition if enforced
	if ownerID != nil && r.Resource.IsOwnershipEnforced() {
		condition[r.Resource.GetOwnerField()] = ownerID
	}

	return r.GenericRepository.FindAllBy(ctx, condition)
}

func (r *OwnerGenericRepository) GetWithRelations(ctx context.Context, id interface{}, relations []string) (interface{}, error) {
	result, err := r.GenericRepository.GetWithRelations(ctx, id, relations)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if err := r.verifyOwnership(ctx, id); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *OwnerGenericRepository) ListWithRelations(ctx context.Context, options query.QueryOptions, relations []string) (interface{}, int64, error) {
	// Apply owner filter to DB
	tx := r.DB.WithContext(ctx)
	var err error
	tx, err = r.applyOwnerFilter(ctx, tx)
	if err != nil {
		return nil, 0, err
	}

	// Use GenericRepository to complete the operation with modified query
	r.DB = tx
	result, total, err := r.GenericRepository.ListWithRelations(ctx, options, relations)

	// Reset DB to original
	r.DB = r.GenericRepository.DB

	return result, total, err
}

func (r *OwnerGenericRepository) Query(ctx context.Context) *gorm.DB {
	// Apply owner filter to query
	tx := r.GenericRepository.Query(ctx)
	tx, err := r.applyOwnerFilter(ctx, tx)
	if err != nil {
		// Since we can't return error, we'll return a query that will return no results
		return tx.Where("1 = 0") // Always false condition
	}
	return tx
}

func (r *OwnerGenericRepository) BulkCreate(ctx context.Context, data interface{}) error {
	// Set owner field
	if err := r.setOwnership(ctx, data); err != nil {
		return err
	}

	return r.GenericRepository.BulkCreate(ctx, data)
}

func (r *OwnerGenericRepository) BulkUpdate(ctx context.Context, condition map[string]interface{}, updates map[string]interface{}) error {
	var err error
	ownerID, err := r.extractOwnerID(ctx)
	if err != nil {
		return err
	}

	// Add owner condition if enforced
	if ownerID != nil && r.Resource.IsOwnershipEnforced() {
		condition[r.Resource.GetOwnerField()] = ownerID
	}

	return r.GenericRepository.BulkUpdate(ctx, condition, updates)
}
