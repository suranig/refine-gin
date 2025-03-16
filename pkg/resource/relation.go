package resource

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RelationType defines the type of relation
type RelationType string

const (
	// RelationTypeOneToOne represents a one-to-one relation
	RelationTypeOneToOne RelationType = "one-to-one"

	// RelationTypeOneToMany represents a one-to-many relation
	RelationTypeOneToMany RelationType = "one-to-many"

	// RelationTypeManyToOne represents a many-to-one relation
	RelationTypeManyToOne RelationType = "many-to-one"

	// RelationTypeManyToMany represents a many-to-many relation
	RelationTypeManyToMany RelationType = "many-to-many"
)

// Relation defines a relation between resources
type Relation struct {
	// Name of the relation
	Name string

	// Type of the relation
	Type RelationType

	// Resource name that this relation refers to
	Resource string

	// Field in the current resource that holds the relation
	Field string

	// Field in the related resource that this relation refers to
	ReferenceField string

	// Whether to include this relation in responses by default
	IncludeByDefault bool
}

// ExtractRelationsFromModel extracts relations from a model using reflection
func ExtractRelationsFromModel(model interface{}) []Relation {
	var relations []Relation

	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Check for relation tag
		if tag, ok := field.Tag.Lookup("relation"); ok {
			relation := parseRelationTag(field.Name, tag)
			if relation != nil {
				relations = append(relations, *relation)
			}
		} else {
			// Try to infer relation from field type
			relation := inferRelationFromField(field)
			if relation != nil {
				relations = append(relations, *relation)
			}
		}
	}

	return relations
}

// parseRelationTag parses a relation tag
func parseRelationTag(fieldName string, tag string) *Relation {
	// Format: resource=users;type=one-to-many;field=user_id;reference=id;include=true
	parts := map[string]string{}

	for _, part := range strings.Split(tag, ";") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			parts[kv[0]] = kv[1]
		}
	}

	// Required fields
	resource, ok1 := parts["resource"]
	typeStr, ok2 := parts["type"]

	if !ok1 || !ok2 {
		return nil
	}

	// Optional fields
	field := parts["field"]
	reference := parts["reference"]
	include := parts["include"] == "true"

	// Default name to field name if not specified
	name := parts["name"]
	if name == "" {
		name = fieldName
	}

	return &Relation{
		Name:             name,
		Type:             RelationType(typeStr),
		Resource:         resource,
		Field:            field,
		ReferenceField:   reference,
		IncludeByDefault: include,
	}
}

// inferRelationFromField tries to infer a relation from a field type
func inferRelationFromField(field reflect.StructField) *Relation {
	fieldType := field.Type

	// Check for slice or array (one-to-many)
	if fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {
		elemType := fieldType.Elem()

		// If it's a slice of structs or pointers to structs
		if elemType.Kind() == reflect.Struct || (elemType.Kind() == reflect.Ptr && elemType.Elem().Kind() == reflect.Struct) {
			if elemType.Kind() == reflect.Ptr {
				elemType = elemType.Elem()
			}

			return &Relation{
				Name:             field.Name,
				Type:             RelationTypeOneToMany,
				Resource:         elemType.Name(),
				Field:            "",
				ReferenceField:   "",
				IncludeByDefault: false,
			}
		}
	}

	// Check for struct or pointer to struct (one-to-one or many-to-one)
	if fieldType.Kind() == reflect.Struct || (fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Struct) {
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		// Skip basic types like time.Time
		if fieldType.PkgPath() == "time" && fieldType.Name() == "Time" {
			return nil
		}

		return &Relation{
			Name:             field.Name,
			Type:             RelationTypeOneToOne, // Default to one-to-one
			Resource:         fieldType.Name(),
			Field:            "",
			ReferenceField:   "",
			IncludeByDefault: false,
		}
	}

	return nil
}

// Helper function to include relations in query
func IncludeRelations(c *gin.Context, res Resource) []string {
	// Check for include parameter
	includeParam := c.Query("include")
	if includeParam == "" {
		// If no include parameter, use default includes
		var defaultIncludes []string
		for _, relation := range res.GetRelations() {
			if relation.IncludeByDefault {
				defaultIncludes = append(defaultIncludes, relation.Name)
			}
		}
		return defaultIncludes
	}

	// Parse include parameter
	includes := strings.Split(includeParam, ",")
	var validIncludes []string

	for _, include := range includes {
		include = strings.TrimSpace(include)
		if res.HasRelation(include) {
			validIncludes = append(validIncludes, include)
		}
	}

	return validIncludes
}

// Helper function to load relations for a record
func LoadRelations(db *gorm.DB, res Resource, record interface{}, includes []string) error {
	if len(includes) == 0 {
		return nil
	}

	for _, include := range includes {
		relation := res.GetRelation(include)
		if relation == nil {
			continue
		}

		// Preload the relation
		db = db.Preload(relation.Name)
	}

	return db.First(record).Error
}

// Helper function to load relations for multiple records
func LoadRelationsForMany(db *gorm.DB, res Resource, records interface{}, includes []string) error {
	if len(includes) == 0 {
		return nil
	}

	for _, include := range includes {
		relation := res.GetRelation(include)
		if relation == nil {
			continue
		}

		// Preload the relation
		db = db.Preload(relation.Name)
	}

	return db.Find(records).Error
}
