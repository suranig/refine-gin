package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

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
			fmt.Printf("[DEBUG] Using default owner ID: %v\n", r.Resource.GetDefaultOwnerID())
			return r.Resource.GetDefaultOwnerID(), nil
		}
		fmt.Printf("[DEBUG] Ownership not enforced, no default ID\n")
		return nil, nil
	}

	// Extract owner ID from context
	ownerID, err := middleware.GetOwnerID(ctx)
	fmt.Printf("[DEBUG] GetOwnerID result: %v (error: %v)\n", ownerID, err)

	if err != nil {
		// If no owner ID is found in context, use default if provided
		if r.Resource.GetDefaultOwnerID() != nil {
			fmt.Printf("[DEBUG] Using default owner ID after error: %v\n", r.Resource.GetDefaultOwnerID())
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
	if r.Resource == nil || !r.Resource.IsOwnershipEnforced() {
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
	if r.Resource == nil || !r.Resource.IsOwnershipEnforced() {
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
	// Log the incoming request
	fmt.Printf("[DEBUG-REPO] Get request for ID: %v\n", id)

	// Log the context owner ID
	if ownerID, exists := ctx.Value(middleware.OwnerContextKey).(string); exists {
		fmt.Printf("[DEBUG-REPO] Context owner ID: %v\n", ownerID)
	} else {
		fmt.Printf("[DEBUG-REPO] WARNING: No owner ID found in context\n")
	}

	// Get the ID field name from the resource or use the default
	idFieldName := "id" // Default to "id"
	if r.Resource != nil {
		idFieldName = r.Resource.GetIDFieldName()
	}

	// Get the proper column name using GORM's naming strategy
	idColumnName := r.DB.NamingStrategy.ColumnName("", idFieldName)
	fmt.Printf("[DEBUG-REPO] ID field name: %s, column name: %s\n", idFieldName, idColumnName)

	// Get the owner field name
	ownerField := r.Resource.GetOwnerField()
	ownerColumnName := r.DB.NamingStrategy.ColumnName("", ownerField)
	fmt.Printf("[DEBUG-REPO] Owner field name: %s, column name: %s\n", ownerField, ownerColumnName)

	// Extract owner ID
	ownerID, err := r.extractOwnerID(ctx)
	if err != nil {
		fmt.Printf("[DEBUG-REPO] Error extracting owner ID: %v\n", err)
		return nil, err
	}

	// Create a new instance of the model to hold the result
	modelType := reflect.TypeOf(r.Model)
	var result interface{}

	// If the model is a struct (not a pointer), make a new instance of it
	if modelType.Kind() != reflect.Ptr {
		// Create a new instance of the model type
		result = reflect.New(modelType).Interface()
	} else {
		// If it's already a pointer, create a new instance of the pointed-to type
		result = reflect.New(modelType.Elem()).Interface()
	}

	// Build query directly with proper column names
	query := r.DB.Model(r.Model)

	// Add the ID condition - use column name from naming strategy
	query = query.Where(fmt.Sprintf("%s = ?", idColumnName), id)

	// Add owner ID condition if enforced
	if r.Resource != nil && r.Resource.IsOwnershipEnforced() && ownerID != nil {
		query = query.Where(fmt.Sprintf("%s = ?", ownerColumnName), ownerID)
	}

	// Log the SQL query
	stmt := query.Statement
	if stmt.SQL.String() != "" {
		fmt.Printf("[DEBUG-REPO] SQL query: %v\n", stmt.SQL.String())
		fmt.Printf("[DEBUG-REPO] Query values: %v\n", stmt.Vars)
	} else {
		fmt.Printf("[DEBUG-REPO] SQL query not available before execution\n")
	}

	// Execute query
	if err := query.First(result).Error; err != nil {
		// Print the SQL that was executed
		fmt.Printf("[DEBUG-REPO] Executed SQL: %v\n", query.Statement.SQL.String())
		fmt.Printf("[DEBUG-REPO] Query error: %v\n", err)

		// Check if record exists without owner filter
		if r.Resource != nil && r.Resource.IsOwnershipEnforced() && err == gorm.ErrRecordNotFound {
			var exists bool
			checkQuery := r.DB.Model(r.Model).Where(fmt.Sprintf("%s = ?", idColumnName), id)

			if err := checkQuery.Select("1").Limit(1).Find(&exists).Error; err != nil {
				return nil, err
			}

			if exists {
				// Record exists but belongs to a different owner
				return nil, ErrOwnerMismatch
			}
		}

		return nil, err
	}

	// Log the result
	fmt.Printf("[DEBUG-REPO] Query result: %+v\n", result)

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
	// Log the incoming request
	fmt.Printf("[DEBUG-REPO] Update request for ID: %v\n", id)
	fmt.Printf("[DEBUG-REPO] Update data: %+v\n", data)

	// Try to set ID directly on model if it implements IDSetter
	TrySetID(data, id)

	// Log the context owner ID
	if ownerID, exists := ctx.Value(middleware.OwnerContextKey).(string); exists {
		fmt.Printf("[DEBUG-REPO] Context owner ID: %v\n", ownerID)
	} else {
		fmt.Printf("[DEBUG-REPO] WARNING: No owner ID found in context\n")
	}

	// Get the ID field name from the resource or use the default
	idFieldName := "id" // Default to "id"
	if r.Resource != nil {
		idFieldName = r.Resource.GetIDFieldName()
	}

	// Get the proper column name using GORM's naming strategy
	idColumnName := r.DB.NamingStrategy.ColumnName("", idFieldName)
	fmt.Printf("[DEBUG-REPO] ID field name: %s, column name: %s\n", idFieldName, idColumnName)

	// Get the owner field name
	ownerField := r.Resource.GetOwnerField()
	ownerColumnName := r.DB.NamingStrategy.ColumnName("", ownerField)
	fmt.Printf("[DEBUG-REPO] Owner field name: %s, column name: %s\n", ownerField, ownerColumnName)

	// Extract owner ID
	ownerID, err := r.extractOwnerID(ctx)
	if err != nil {
		fmt.Printf("[DEBUG-REPO] Error extracting owner ID: %v\n", err)
		return nil, err
	}

	// First check if the record exists and belongs to the owner
	var exists bool
	checkQuery := r.DB.Model(r.Model).
		Where(fmt.Sprintf("%s = ?", idColumnName), id)

	// Add owner condition if ownership is enforced
	if r.Resource != nil && r.Resource.IsOwnershipEnforced() && ownerID != nil {
		checkQuery = checkQuery.Where(fmt.Sprintf("%s = ?", ownerColumnName), ownerID)
	}

	// Execute the check query
	err = checkQuery.Select("1").Limit(1).Find(&exists).Error
	if err != nil {
		fmt.Printf("[DEBUG-REPO] Error checking record: %v\n", err)
		return nil, err
	}

	// Debug logs
	fmt.Printf("[DEBUG-REPO] Record exists with owner check: %v\n", exists)

	if !exists && r.Resource != nil && r.Resource.IsOwnershipEnforced() {
		// Check if record exists at all
		var recordExists bool
		err = r.DB.Model(r.Model).
			Where(fmt.Sprintf("%s = ?", idColumnName), id).
			Select("1").Limit(1).Find(&recordExists).Error

		if err != nil {
			fmt.Printf("[DEBUG-REPO] Error checking if record exists: %v\n", err)
			return nil, err
		}

		fmt.Printf("[DEBUG-REPO] Record exists at all: %v\n", recordExists)

		if recordExists {
			// Record exists but belongs to another owner
			fmt.Printf("[DEBUG-REPO] Owner mismatch - access forbidden\n")
			return nil, ErrOwnerMismatch
		} else {
			// Record not found
			fmt.Printf("[DEBUG-REPO] Record not found\n")
			return nil, gorm.ErrRecordNotFound
		}
	}

	// Handle different update methods based on data type
	dataMap, isMap := data.(map[string]interface{})
	if isMap && r.Resource != nil && r.Resource.IsOwnershipEnforced() {
		// Check if there's an empty owner ID field in the map, and remove it if it is
		if ownerVal, hasOwner := dataMap[ownerField]; hasOwner {
			// If owner value is empty string or nil, remove it to avoid clearing the owner
			if ownerVal == "" || ownerVal == nil {
				fmt.Printf("[DEBUG-REPO] Removing empty owner field from update data\n")
				delete(dataMap, ownerField)
			}
		}

		// Also check the database column name version
		if ownerVal, hasOwner := dataMap[ownerColumnName]; hasOwner {
			// If owner value is empty string or nil, remove it to avoid clearing the owner
			if ownerVal == "" || ownerVal == nil {
				fmt.Printf("[DEBUG-REPO] Removing empty owner column from update data\n")
				delete(dataMap, ownerColumnName)
			}
		}

		// Print the modified data
		fmt.Printf("[DEBUG-REPO] Modified update data: %+v\n", dataMap)
	}

	// First try to get the existing record
	var result interface{}
	modelType := reflect.TypeOf(r.Model)
	if modelType.Kind() == reflect.Ptr {
		result = reflect.New(modelType.Elem()).Interface()
	} else {
		result = reflect.New(modelType).Interface()
	}
	if err := r.DB.Where(fmt.Sprintf("%s = ?", idColumnName), id).First(result).Error; err != nil {
		fmt.Printf("[DEBUG-REPO] Error fetching existing record: %v\n", err)
		return nil, err
	}

	// For JSON serialized fields, we need special handling
	if isMap {
		// Get the model struct type
		resultVal := reflect.ValueOf(result).Elem()
		resultType := resultVal.Type()

		// Check each field for JSON serializer tag
		for i := 0; i < resultType.NumField(); i++ {
			field := resultType.Field(i)

			// Look for "serializer:json" tag
			if gormTag, ok := field.Tag.Lookup("gorm"); ok && strings.Contains(gormTag, "serializer:json") {
				fieldName := field.Name
				jsonFieldName := field.Tag.Get("json")
				if idx := strings.Index(jsonFieldName, ","); idx > 0 {
					jsonFieldName = jsonFieldName[:idx]
				}

				// If this JSON field is in the data map
				if jsonData, ok := dataMap[jsonFieldName]; ok {
					fmt.Printf("[DEBUG-REPO] Found JSON serialized field %s (%s) in update data\n", fieldName, jsonFieldName)

					// Convert the map to JSON
					jsonBytes, err := json.Marshal(jsonData)
					if err != nil {
						fmt.Printf("[DEBUG-REPO] Error marshaling JSON field: %v\n", err)
						continue
					}

					// Create a new instance of the field type
					fieldValue := resultVal.Field(i)
					fieldType := fieldValue.Type()

					// Create a new instance of the field type
					newValue := reflect.New(fieldType).Interface()

					// Unmarshal JSON into the new value
					if err := json.Unmarshal(jsonBytes, newValue); err != nil {
						fmt.Printf("[DEBUG-REPO] Error unmarshaling JSON to field: %v\n", err)
						continue
					}

					// Set the field value
					fieldValue.Set(reflect.ValueOf(newValue).Elem())

					// Remove from dataMap to avoid double updating
					delete(dataMap, jsonFieldName)

					fmt.Printf("[DEBUG-REPO] Successfully updated JSON field %s\n", fieldName)
				}
			}
		}

		// Convert the map data to JSON for standard update
		jsonData, err := json.Marshal(dataMap)
		if err != nil {
			fmt.Printf("[DEBUG-REPO] Error marshaling update data to JSON: %v\n", err)
			return nil, err
		}

		// Update the existing model with the JSON data for remaining fields
		if err := json.Unmarshal(jsonData, result); err != nil {
			fmt.Printf("[DEBUG-REPO] Error unmarshaling JSON to model: %v\n", err)
		} else {
			fmt.Printf("[DEBUG-REPO] Successfully updated model with JSON data\n")

			// Keep track of the original ID to prevent it from being overwritten
			idField := reflect.ValueOf(result).Elem().FieldByName(idFieldName)
			if idField.IsValid() {
				// Check if ID field exists but we don't need to store its value as it's preserved in the query
				fmt.Printf("[DEBUG-REPO] Original ID field found: %v\n", idField.Interface())
			}

			// Save the updated model
			updateQuery := r.DB.Model(r.Model).
				Where(fmt.Sprintf("%s = ?", idColumnName), id)

			// Add owner condition if ownership is enforced
			if r.Resource != nil && r.Resource.IsOwnershipEnforced() && ownerID != nil {
				updateQuery = updateQuery.Where(fmt.Sprintf("%s = ?", ownerColumnName), ownerID)
			}

			// Save the entire record
			if err := updateQuery.Save(result).Error; err != nil {
				fmt.Printf("[DEBUG-REPO] Error saving updated model: %v\n", err)

				// If saving the whole model fails, try to update with the map data
				if err := updateQuery.Updates(dataMap).Error; err != nil {
					fmt.Printf("[DEBUG-REPO] Error updating with map data: %v\n", err)
					return nil, err
				}
			} else {
				fmt.Printf("[DEBUG-REPO] Successfully saved updated model\n")
				return result, nil
			}
		}
	}

	// Now perform the standard update if JSON approach didn't work
	updateQuery := r.DB.Model(r.Model).
		Where(fmt.Sprintf("%s = ?", idColumnName), id)

	// Add owner condition if ownership is enforced
	if r.Resource != nil && r.Resource.IsOwnershipEnforced() && ownerID != nil {
		updateQuery = updateQuery.Where(fmt.Sprintf("%s = ?", ownerColumnName), ownerID)
	}

	// Log the SQL query
	stmt := updateQuery.Statement
	if stmt.SQL.String() != "" {
		fmt.Printf("[DEBUG-REPO] Update SQL query: %v\n", stmt.SQL.String())
		fmt.Printf("[DEBUG-REPO] Update values: %v\n", stmt.Vars)
	} else {
		fmt.Printf("[DEBUG-REPO] Update SQL query not available before execution\n")
	}

	fmt.Printf("[DEBUG-REPO] Executing update\n")
	var updateErr error
	if isMap {
		updateErr = updateQuery.Updates(dataMap).Error
	} else {
		updateErr = updateQuery.Updates(data).Error
	}

	if updateErr != nil {
		fmt.Printf("[DEBUG-REPO] Update error: %v\n", updateErr)
		return nil, updateErr
	}

	fmt.Printf("[DEBUG-REPO] Update successful\n")

	// Get the updated record directly from the database rather than using r.Get
	// to avoid potential ownership check issues after update
	var fetchResult interface{}
	modelType = reflect.TypeOf(r.Model)
	if modelType.Kind() == reflect.Ptr {
		fetchResult = reflect.New(modelType.Elem()).Interface()
	} else {
		fetchResult = reflect.New(modelType).Interface()
	}
	if err := r.DB.Where(fmt.Sprintf("%s = ?", idColumnName), id).First(fetchResult).Error; err != nil {
		fmt.Printf("[DEBUG-REPO] Error fetching updated record: %v\n", err)
		return nil, err
	}

	fmt.Printf("[DEBUG-REPO] Successfully retrieved updated record: %+v\n", fetchResult)

	return fetchResult, nil
}

// Delete removes a resource after verifying ownership
func (r *OwnerGenericRepository) Delete(ctx context.Context, id interface{}) error {
	// If ownership is not enforced, use standard repository logic
	if r.Resource == nil || !r.Resource.IsOwnershipEnforced() {
		return r.GenericRepository.Delete(ctx, id)
	}

	// Extract owner ID from context
	ownerID, err := r.extractOwnerID(ctx)
	if err != nil {
		return err
	}

	// If no owner ID present, use standard repository logic
	if ownerID == nil {
		return r.GenericRepository.Delete(ctx, id)
	}

	// Get the ID field name
	idFieldName := "id" // Default to "id"
	if r.Resource != nil {
		idFieldName = r.Resource.GetIDFieldName()
	}

	// Get proper column name using GORM's naming strategy
	idColumnName := r.DB.NamingStrategy.ColumnName("", idFieldName)

	// Get the owner field name
	ownerField := r.Resource.GetOwnerField()
	ownerColumnName := r.DB.NamingStrategy.ColumnName("", ownerField)

	// Check if the record exists and belongs to the owner - Start with fresh query
	var exists bool
	result := r.DB.WithContext(ctx).
		Model(r.Model). // Use Model to ensure we reset any previous conditions
		Where(fmt.Sprintf("%s = ?", idColumnName), id).
		Where(fmt.Sprintf("%s = ?", ownerColumnName), ownerID).
		Select("COUNT(*) > 0").
		Find(&exists)

	if result.Error != nil {
		return result.Error
	}

	if !exists {
		// Check if record exists at all - Start with fresh query
		var recordExists bool
		r.DB.WithContext(ctx).
			Model(r.Model). // Use Model to ensure we reset any previous conditions
			Where(fmt.Sprintf("%s = ?", idColumnName), id).
			Select("COUNT(*) > 0").
			Find(&recordExists)

		if recordExists {
			// Record exists but belongs to another owner
			return ErrOwnerMismatch
		}
		// Record doesn't exist
		return gorm.ErrRecordNotFound
	}

	// Delete with both ID and owner filter - Start with fresh query
	return r.DB.WithContext(ctx).
		Model(r.Model). // Use Model to ensure we reset any previous conditions
		Where(fmt.Sprintf("%s = ?", idColumnName), id).
		Where(fmt.Sprintf("%s = ?", ownerColumnName), ownerID).
		Delete(r.Model).Error
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
	origDB := r.DB
	r.DB = tx
	total, err := r.GenericRepository.Count(ctx, options)

	// Reset DB to original
	r.DB = origDB

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
	// If ownership is not enforced, use standard repository logic
	if r.Resource == nil || !r.Resource.IsOwnershipEnforced() {
		return r.GenericRepository.UpdateMany(ctx, ids, data)
	}

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
	// If ownership is not enforced, use standard repository logic
	if r.Resource == nil || !r.Resource.IsOwnershipEnforced() {
		return r.GenericRepository.DeleteMany(ctx, ids)
	}

	// Get the ID field name
	idFieldName := "id" // Default to "id"
	if r.Resource != nil {
		idFieldName = r.Resource.GetIDFieldName()
	}

	// Get proper column name using GORM's naming strategy
	idColumnName := r.DB.NamingStrategy.ColumnName("", idFieldName)

	// Get the owner field name
	ownerField := r.Resource.GetOwnerField()
	ownerColumnName := r.DB.NamingStrategy.ColumnName("", ownerField)

	// Extract owner ID from context
	ownerID, err := r.extractOwnerID(ctx)
	if err != nil {
		return 0, err
	}

	// If no owner ID present, use standard repository logic
	if ownerID == nil {
		return r.GenericRepository.DeleteMany(ctx, ids)
	}

	// Check if all records exist and belong to the owner
	for _, id := range ids {
		var exists bool
		result := r.DB.WithContext(ctx).
			Model(r.Model).
			Where(fmt.Sprintf("%s = ?", idColumnName), id).
			Where(fmt.Sprintf("%s = ?", ownerColumnName), ownerID).
			Select("COUNT(*) > 0").
			Find(&exists)

		if result.Error != nil {
			return 0, result.Error
		}

		if !exists {
			// Check if record exists at all
			var recordExists bool
			r.DB.WithContext(ctx).
				Model(r.Model).
				Where(fmt.Sprintf("%s = ?", idColumnName), id).
				Select("COUNT(*) > 0").
				Find(&recordExists)

			if recordExists {
				// Record exists but belongs to another owner
				return 0, ErrOwnerMismatch
			}
			// Record doesn't exist
			return 0, gorm.ErrRecordNotFound
		}
	}

	// Delete all records with both ID and owner filter
	result := r.DB.WithContext(ctx).
		Model(r.Model).
		Where(fmt.Sprintf("%s IN ?", idColumnName), ids).
		Where(fmt.Sprintf("%s = ?", ownerColumnName), ownerID).
		Delete(r.Model)

	return result.RowsAffected, result.Error
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
	if ownerID != nil && r.Resource != nil && r.Resource.IsOwnershipEnforced() {
		// Get owner field name and convert to column name using GORM's naming strategy
		ownerField := r.Resource.GetOwnerField()
		ownerColumnName := r.DB.NamingStrategy.ColumnName("", ownerField)
		condition[ownerColumnName] = ownerID
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
	if ownerID != nil && r.Resource != nil && r.Resource.IsOwnershipEnforced() {
		// Get owner field name and convert to column name using GORM's naming strategy
		ownerField := r.Resource.GetOwnerField()
		ownerColumnName := r.DB.NamingStrategy.ColumnName("", ownerField)
		condition[ownerColumnName] = ownerID
	}

	return r.GenericRepository.FindAllBy(ctx, condition)
}

func (r *OwnerGenericRepository) GetWithRelations(ctx context.Context, id interface{}, relations []string) (interface{}, error) {
	// If ownership is not enforced, use standard repository logic
	if r.Resource == nil || !r.Resource.IsOwnershipEnforced() {
		return r.GenericRepository.GetWithRelations(ctx, id, relations)
	}

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
	if ownerID != nil && r.Resource != nil && r.Resource.IsOwnershipEnforced() {
		// Get owner field name and convert to column name using GORM's naming strategy
		ownerField := r.Resource.GetOwnerField()
		ownerColumnName := r.DB.NamingStrategy.ColumnName("", ownerField)
		condition[ownerColumnName] = ownerID
	}

	return r.GenericRepository.BulkUpdate(ctx, condition, updates)
}

// GetIDFieldName returns the field name used as primary key
func (r *OwnerGenericRepository) GetIDFieldName() string {
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
