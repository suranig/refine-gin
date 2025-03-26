package resource

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

// RelationValidator validates relations between resources
type RelationValidator struct {
	// The relation being validated
	Relation Relation

	// Minimum number of related items (for to-many relations)
	MinItems int

	// Maximum number of related items (for to-many relations)
	MaxItems int

	// Whether the relation is required
	Required bool

	// Custom error message
	Message string

	// Database connection for validation
	DB *gorm.DB

	// The related resource
	RelatedResource Resource
}

// Validate validates a relation value
func (v RelationValidator) Validate(value interface{}) error {
	if value == nil {
		if v.Required {
			if v.Message != "" {
				return fmt.Errorf(v.Message)
			}
			return fmt.Errorf("relation %s is required", v.Relation.Name)
		}
		return nil
	}

	// Get the value reflection
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Check relation type
	switch v.Relation.Type {
	case RelationTypeOneToOne, RelationTypeManyToOne:
		return v.validateToOneRelation(val)
	case RelationTypeOneToMany, RelationTypeManyToMany:
		return v.validateToManyRelation(val)
	default:
		return fmt.Errorf("unsupported relation type: %s", v.Relation.Type)
	}
}

// validateToOneRelation validates a one-to-one or many-to-one relation
func (v RelationValidator) validateToOneRelation(val reflect.Value) error {
	// Skip validation if DB is not provided
	if v.DB == nil {
		return nil
	}

	// Get the reference field value
	var refValue interface{}
	if val.Kind() == reflect.Struct {
		field := val.FieldByName(v.Relation.ReferenceField)
		if field.IsValid() {
			refValue = field.Interface()
		}
	} else {
		refValue = val.Interface()
	}

	if refValue == nil {
		if v.Required {
			if v.Message != "" {
				return fmt.Errorf(v.Message)
			}
			return fmt.Errorf("relation %s is required", v.Relation.Name)
		}
		return nil
	}

	// Check if the related record exists
	var count int64
	result := v.DB.Model(v.RelatedResource.GetModel()).
		Where(fmt.Sprintf("%s = ?", v.Relation.ReferenceField), refValue).
		Count(&count)

	if result.Error != nil {
		return fmt.Errorf("error validating relation %s: %w", v.Relation.Name, result.Error)
	}

	if count == 0 {
		if v.Message != "" {
			return fmt.Errorf(v.Message)
		}
		return fmt.Errorf("referenced %s with %s = %v does not exist",
			v.Relation.Resource, v.Relation.ReferenceField, refValue)
	}

	return nil
}

// validateToManyRelation validates a one-to-many or many-to-many relation
func (v RelationValidator) validateToManyRelation(val reflect.Value) error {
	// Check if value is a slice
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return fmt.Errorf("expected slice or array for to-many relation %s, got %s",
			v.Relation.Name, val.Kind())
	}

	// Check number of items
	numItems := val.Len()

	if v.Required && numItems == 0 {
		if v.Message != "" {
			return fmt.Errorf(v.Message)
		}
		return fmt.Errorf("relation %s is required", v.Relation.Name)
	}

	if v.MinItems > 0 && numItems < v.MinItems {
		if v.Message != "" {
			return fmt.Errorf(v.Message)
		}
		return fmt.Errorf("relation %s requires at least %d items, got %d",
			v.Relation.Name, v.MinItems, numItems)
	}

	if v.MaxItems > 0 && numItems > v.MaxItems {
		if v.Message != "" {
			return fmt.Errorf(v.Message)
		}
		return fmt.Errorf("relation %s allows at most %d items, got %d",
			v.Relation.Name, v.MaxItems, numItems)
	}

	// Skip further validation if DB is not provided
	if v.DB == nil {
		return nil
	}

	// Check if referenced records exist
	for i := 0; i < numItems; i++ {
		item := val.Index(i)
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}

		// Get the reference field value
		var refValue interface{}
		if item.Kind() == reflect.Struct {
			field := item.FieldByName(v.Relation.ReferenceField)
			if field.IsValid() {
				refValue = field.Interface()
			}
		} else {
			refValue = item.Interface()
		}

		if refValue == nil {
			continue
		}

		// Check if the related record exists
		var count int64
		result := v.DB.Model(v.RelatedResource.GetModel()).
			Where(fmt.Sprintf("%s = ?", v.Relation.ReferenceField), refValue).
			Count(&count)

		if result.Error != nil {
			return fmt.Errorf("error validating relation %s: %w", v.Relation.Name, result.Error)
		}

		if count == 0 {
			if v.Message != "" {
				return fmt.Errorf(v.Message)
			}
			return fmt.Errorf("referenced %s with %s = %v does not exist",
				v.Relation.Resource, v.Relation.ReferenceField, refValue)
		}
	}

	return nil
}

// ValidateRelations validates all relations in a model
func ValidateRelations(db *gorm.DB, obj interface{}) error {
	objType := reflect.TypeOf(obj)
	objVal := reflect.ValueOf(obj)

	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
		objVal = objVal.Elem()
	}

	// Get the resource from registry based on model type
	registry := GlobalResourceRegistry
	var res Resource
	var found bool

	// Loop through all registered resources
	for _, r := range registry.GetAll() {
		resourceModel := r.GetModel()
		resourceType := reflect.TypeOf(resourceModel)
		if resourceType.Kind() == reflect.Ptr {
			resourceType = resourceType.Elem()
		}

		// If resource model type matches the object type
		if resourceType == objType {
			res = r
			found = true
			break
		}
	}

	if !found {
		return nil // No resource found for this model, skip validation
	}

	// Validate each relation
	for _, relation := range res.GetRelations() {
		// Get the field value
		field := objVal.FieldByName(relation.Field)
		if !field.IsValid() {
			continue
		}

		// Create validator for the relation
		validator := RelationValidator{
			Relation: relation,
			Required: relation.Required,
			MinItems: relation.MinItems,
			MaxItems: relation.MaxItems,
			DB:       db,
		}

		if err := validator.Validate(field.Interface()); err != nil {
			return err
		}
	}

	return nil
}
