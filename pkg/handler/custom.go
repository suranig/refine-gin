package handler

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
	"github.com/suranig/refine-gin/pkg/utils"
)

// CustomAction represents a custom action that can be performed on a resource
type CustomAction struct {
	// Name of the action (used in URL)
	Name string

	// HTTP method for the action
	Method string

	// Whether the action requires an ID parameter
	RequiresID bool

	// Handler function for the action
	Handler func(*gin.Context, resource.Resource, repository.Repository) (interface{}, error)

	// Whether the action is for bulk operations
	IsBulk bool
}

// CustomActionResponse is the standard response for custom actions
type CustomActionResponse struct {
	Data interface{} `json:"data"`
}

// GenerateCustomActionHandler creates a handler for a custom action
func GenerateCustomActionHandler(res resource.Resource, repo repository.Repository, action CustomAction) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Execute the custom action
		result, err := action.Handler(c, res, repo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Set no-cache headers for actions that modify data
		if action.Method != http.MethodGet && action.Method != http.MethodHead {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		}

		// Return the result
		c.JSON(http.StatusOK, CustomActionResponse{
			Data: result,
		})
	}
}

// RegisterCustomActions registers custom actions for a resource
func RegisterCustomActions(router *gin.RouterGroup, res resource.Resource, repo repository.Repository, actions []CustomAction) {
	for _, action := range actions {
		path := "/" + res.GetName()

		// Add ID parameter if required
		if action.RequiresID {
			path += "/:id"
		}

		// Add action name to path
		path += "/actions/" + action.Name

		// Register the route with the appropriate HTTP method
		switch strings.ToUpper(action.Method) {
		case http.MethodGet:
			router.GET(path, GenerateCustomActionHandler(res, repo, action))
		case http.MethodPost:
			router.POST(path, GenerateCustomActionHandler(res, repo, action))
		case http.MethodPut:
			router.PUT(path, GenerateCustomActionHandler(res, repo, action))
		case http.MethodPatch:
			router.PATCH(path, GenerateCustomActionHandler(res, repo, action))
		case http.MethodDelete:
			router.DELETE(path, GenerateCustomActionHandler(res, repo, action))
		default:
			// Default to POST if method is not recognized
			router.POST(path, GenerateCustomActionHandler(res, repo, action))
		}
	}
}

// ActionOperation creates an Operation for a custom action
func ActionOperation(actionName string) resource.Operation {
	return resource.Operation("custom:" + actionName)
}

// RelationRequest represents a request to attach or detach a related resource
type RelationRequest struct {
	IDs []interface{} `json:"ids"`
}

// RelationResponse is the standard response for relation operations
type RelationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// RelationType dla jasności kodu
const (
	HasOne     = resource.RelationTypeOneToOne
	HasMany    = resource.RelationTypeOneToMany
	BelongsTo  = resource.RelationTypeManyToOne
	ManyToMany = resource.RelationTypeManyToMany
)

// AttachAction creates a custom action for attaching related resources
func AttachAction(relationName string) CustomAction {
	return CustomAction{
		Name:       fmt.Sprintf("attach-%s", relationName),
		Method:     http.MethodPost,
		RequiresID: true,
		Handler: func(c *gin.Context, res resource.Resource, repo repository.Repository) (interface{}, error) {
			id := c.Param("id")

			// Parse request body
			var req RelationRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				return nil, err
			}

			if len(req.IDs) == 0 {
				return nil, fmt.Errorf("no IDs provided")
			}

			// Get the parent resource
			parentObj, err := repo.Get(c, id)
			if err != nil {
				return nil, fmt.Errorf("resource not found")
			}

			// Check if relation exists
			relation := getRelationByName(res, relationName)
			if relation == nil {
				return nil, fmt.Errorf("relation %s not found", relationName)
			}

			// Attach related objects
			switch relation.Type {
			case HasMany:
				// For HasMany, we need to get the collection field and append to it
				err = attachToHasManyRelation(parentObj, relation, req.IDs, repo)
			case HasOne:
				// For HasOne, we set the field directly
				err = attachToHasOneRelation(parentObj, relation, req.IDs[0], repo)
			case BelongsTo:
				// For BelongsTo, we set the foreign key field
				err = attachToBelongsToRelation(parentObj, relation, req.IDs[0], repo)
			case ManyToMany:
				// For ManyToMany, we need to use a join table
				err = attachToManyToManyRelation(parentObj, relation, req.IDs, repo, res, id)
			default:
				return nil, fmt.Errorf("unsupported relation type: %s", relation.Type)
			}

			if err != nil {
				return nil, err
			}

			// Update the parent object with the new relations
			_, err = repo.Update(c, id, parentObj)
			if err != nil {
				return nil, err
			}

			return RelationResponse{
				Success: true,
				Message: fmt.Sprintf("Successfully attached %d %s", len(req.IDs), relationName),
			}, nil
		},
		IsBulk: false,
	}
}

// DetachAction creates a custom action for detaching related resources
func DetachAction(relationName string) CustomAction {
	return CustomAction{
		Name:       fmt.Sprintf("detach-%s", relationName),
		Method:     http.MethodPost,
		RequiresID: true,
		Handler: func(c *gin.Context, res resource.Resource, repo repository.Repository) (interface{}, error) {
			id := c.Param("id")

			// Check if relation exists
			relation := getRelationByName(res, relationName)
			if relation == nil {
				return nil, fmt.Errorf("relation %s not found", relationName)
			}

			// Parse request body
			var req RelationRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				return nil, err
			}

			if len(req.IDs) == 0 {
				return nil, fmt.Errorf("no IDs provided to detach")
			}

			// Get the parent resource
			parentObj, err := repo.Get(c, id)
			if err != nil {
				return nil, err
			}

			// Detach related objects
			switch relation.Type {
			case HasMany:
				// For HasMany, we need to remove items from the collection
				err = detachFromHasManyRelation(parentObj, relation, req.IDs)
			case HasOne:
				// For HasOne, we set the field to nil
				err = detachFromHasOneRelation(parentObj, relation)
			case BelongsTo:
				// For BelongsTo, we set the foreign key field to nil
				err = detachFromBelongsToRelation(parentObj, relation)
			case ManyToMany:
				// For ManyToMany, we need to remove from the join table
				err = detachFromManyToManyRelation(parentObj, relation, req.IDs, repo, res, id)
			default:
				return nil, fmt.Errorf("unsupported relation type: %s", relation.Type)
			}

			if err != nil {
				return nil, err
			}

			// Update the parent object with the relations removed
			_, err = repo.Update(c, id, parentObj)
			if err != nil {
				return nil, err
			}

			return RelationResponse{
				Success: true,
				Message: fmt.Sprintf("Successfully detached %d %s", len(req.IDs), relationName),
			}, nil
		},
		IsBulk: false,
	}
}

// ListRelationAction creates a custom action for listing related resources
func ListRelationAction(relationName string) CustomAction {
	return CustomAction{
		Name:       fmt.Sprintf("list-%s", relationName),
		Method:     http.MethodGet,
		RequiresID: true,
		Handler: func(c *gin.Context, res resource.Resource, repo repository.Repository) (interface{}, error) {
			id := c.Param("id")

			// Check if relation exists
			relation := getRelationByName(res, relationName)
			if relation == nil {
				return nil, fmt.Errorf("relation %s not found", relationName)
			}

			// Get the parent resource
			parentObj, err := repo.Get(c, id)
			if err != nil {
				return nil, err
			}

			// Get related objects based on relation type
			var relatedObjects interface{}

			switch relation.Type {
			case HasMany, ManyToMany:
				// For HasMany and ManyToMany, we return a collection
				relatedObjects, err = getRelatedCollection(parentObj, relation, repo)
			case HasOne, BelongsTo:
				// For HasOne and BelongsTo, we return a single object
				relatedObjects, err = getRelatedObject(parentObj, relation, repo)
			default:
				return nil, fmt.Errorf("unsupported relation type: %s", relation.Type)
			}

			if err != nil {
				return nil, err
			}

			return relatedObjects, nil
		},
		IsBulk: false,
	}
}

// RegisterResourceForRefineWithRelations registers a resource with relation actions for Refine.dev
func RegisterResourceForRefineWithRelations(
	router *gin.RouterGroup,
	res resource.Resource,
	repo repository.Repository,
	idParamName string,
	relationNames []string,
) {
	// Register standard CRUD operations from register.go
	RegisterResourceForRefine(router, res, repo, idParamName)

	// Register relation actions for each relation
	var actions []CustomAction

	for _, relationName := range relationNames {
		// Add attach action
		actions = append(actions, AttachAction(relationName))

		// Add detach action
		actions = append(actions, DetachAction(relationName))

		// Add list action
		actions = append(actions, ListRelationAction(relationName))
	}

	// Register all actions
	RegisterCustomActions(router, res, repo, actions)
}

// Helper functions for handling relations

// getRelationByName finds a relation by name in a resource
func getRelationByName(res resource.Resource, name string) *resource.Relation {
	relations := res.GetRelations()
	for i := range relations {
		if relations[i].Name == name {
			return &relations[i]
		}
	}
	return nil
}

// attachToHasManyRelation attaches objects to a HasMany relation
func attachToHasManyRelation(parentObj interface{}, relation *resource.Relation, ids []interface{}, repo repository.Repository) error {
	field, err := utils.GetSliceField(parentObj, relation.Field)
	if err != nil {
		return err
	}

	// For each ID, create or get the related object and append to the slice
	for _, id := range ids {
		// Get the related object using empty query options
		relatedObj, err := repo.Get(context.Background(), fmt.Sprintf("%v", id))
		if err != nil {
			return err
		}

		// Create a value for the related object
		relatedVal := reflect.ValueOf(relatedObj)

		// Append to the slice
		field.Set(reflect.Append(field, relatedVal))
	}

	return nil
}

// attachToHasOneRelation attaches an object to a HasOne relation
func attachToHasOneRelation(parentObj interface{}, relation *resource.Relation, id interface{}, repo repository.Repository) error {
	// Get the related object using empty query options
	relatedObj, err := repo.Get(context.Background(), fmt.Sprintf("%v", id))
	if err != nil {
		return err
	}

	// Set the field
	return utils.SetFieldValue(parentObj, relation.Field, relatedObj)
}

// attachToBelongsToRelation attaches an object to a BelongsTo relation
func attachToBelongsToRelation(parentObj interface{}, relation *resource.Relation, id interface{}, repo repository.Repository) error {
	// Get the related object to ensure it exists using empty query options
	_, err := repo.Get(context.Background(), fmt.Sprintf("%v", id))
	if err != nil {
		return err
	}

	// For BelongsTo, we assume a foreign key field with the same name as the relation plus "ID"
	foreignKeyField := relation.Name + "ID"

	// Set the foreign key field
	return utils.SetFieldValue(parentObj, foreignKeyField, id)
}

// attachToManyToManyRelation attaches objects to a ManyToMany relation
func attachToManyToManyRelation(parentObj interface{}, relation *resource.Relation, ids []interface{}, repo repository.Repository, res resource.Resource, parentID string) error {
	field, err := utils.GetSliceField(parentObj, relation.Field)
	if err != nil {
		return err
	}

	// For each ID, append to the slice if not already there
	for _, id := range ids {
		// Convert to the appropriate type
		idVal := reflect.ValueOf(id)

		// Check if ID already exists in the slice
		exists := false
		for i := 0; i < field.Len(); i++ {
			if reflect.DeepEqual(field.Index(i).Interface(), idVal.Interface()) {
				exists = true
				break
			}
		}

		// If not exists, append
		if !exists {
			field.Set(reflect.Append(field, idVal))
		}
	}

	return nil
}

// detachFromHasManyRelation detaches objects from a HasMany relation
func detachFromHasManyRelation(parentObj interface{}, relation *resource.Relation, ids []interface{}) error {
	field, err := utils.GetSliceField(parentObj, relation.Field)
	if err != nil {
		return err
	}

	// Create a new slice without the specified IDs
	newSlice := reflect.MakeSlice(field.Type(), 0, field.Len())

	// Create a map of IDs to remove for faster lookup. Convert to strings so
	// numeric types from JSON (float64) match struct field types like int/uint.
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[fmt.Sprintf("%v", id)] = true
	}

	// Copy all elements except those with matching IDs
	for i := 0; i < field.Len(); i++ {
		item := field.Index(i)

		// Get the ID of the item
		var itemIDStr string
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}

		// Assume the ID field is named "ID" - this should be customizable
		idField := item.FieldByName("ID")
		if !idField.IsValid() {
			continue
		}

		itemIDStr = fmt.Sprintf("%v", idField.Interface())

		// If not in the remove list, keep it
		if !idMap[itemIDStr] {
			newSlice = reflect.Append(newSlice, field.Index(i))
		}
	}

	// Set the new slice
	field.Set(newSlice)

	return nil
}

// detachFromHasOneRelation detaches an object from a HasOne relation
func detachFromHasOneRelation(parentObj interface{}, relation *resource.Relation) error {
	v := reflect.ValueOf(parentObj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	field := v.FieldByName(relation.Field)
	if !field.IsValid() {
		return fmt.Errorf("field %s not found", relation.Field)
	}

	// Check if the field can be set to nil
	if field.Kind() == reflect.Ptr {
		field.Set(reflect.Zero(field.Type()))
		return nil
	}

	return fmt.Errorf("cannot set non-pointer field %s to nil", relation.Field)
}

// detachFromBelongsToRelation detaches an object from a BelongsTo relation
func detachFromBelongsToRelation(parentObj interface{}, relation *resource.Relation) error {
	// For BelongsTo, we assume a foreign key field with the same name as the relation plus "ID"
	foreignKeyField := relation.Name + "ID"

	v := reflect.ValueOf(parentObj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	field := v.FieldByName(foreignKeyField)
	if !field.IsValid() {
		return fmt.Errorf("foreign key field %s not found", foreignKeyField)
	}

	// Set to zero value
	field.Set(reflect.Zero(field.Type()))

	return nil
}

// detachFromManyToManyRelation detaches objects from a ManyToMany relation
func detachFromManyToManyRelation(parentObj interface{}, relation *resource.Relation, ids []interface{}, repo repository.Repository, res resource.Resource, parentID string) error {
	field, err := utils.GetSliceField(parentObj, relation.Field)
	if err != nil {
		return err
	}

	// Create a new slice without the specified IDs
	newSlice := reflect.MakeSlice(field.Type(), 0, field.Len())

	// Create a map of IDs to remove for faster lookup
	idMap := make(map[interface{}]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	// Copy all elements except those with matching IDs
	for i := 0; i < field.Len(); i++ {
		item := field.Index(i).Interface()

		// If not in the remove list, keep it
		if !idMap[item] {
			newSlice = reflect.Append(newSlice, field.Index(i))
		}
	}

	// Set the new slice
	field.Set(newSlice)

	return nil
}

// getRelatedCollection gets a collection of related objects
func getRelatedCollection(parentObj interface{}, relation *resource.Relation, repo repository.Repository) (interface{}, error) {
	// If parentObj is a map, try to get the field from the map
	if m, ok := parentObj.(map[string]interface{}); ok {
		if field, exists := m[relation.Field]; exists {
			return field, nil
		}
		return nil, fmt.Errorf("field %s not found in map", relation.Field)
	}

	// For struct types, use reflection
	value, err := utils.GetFieldValue(parentObj, relation.Field)
	if err != nil {
		return nil, err
	}

	if !utils.IsSlice(value) {
		return nil, fmt.Errorf("field %s is not a slice", relation.Field)
	}

	return value, nil
}

// getRelatedObject gets a single related object
func getRelatedObject(parentObj interface{}, relation *resource.Relation, repo repository.Repository) (interface{}, error) {
	// If parentObj is a map, try to get the field from the map
	if m, ok := parentObj.(map[string]interface{}); ok {
		if relation.Type == resource.RelationTypeOneToOne {
			if field, exists := m[relation.Field]; exists {
				return field, nil
			}
			return nil, fmt.Errorf("field %s not found in map", relation.Field)
		}

		if relation.Type == resource.RelationTypeManyToOne {
			foreignKeyField := relation.Name + "ID"
			if foreignKeyValue, exists := m[foreignKeyField]; exists {
				if foreignKeyValue == nil || reflect.ValueOf(foreignKeyValue).IsZero() {
					return nil, nil
				}
				// Get the related object using empty query options
				relatedObj, err := repo.Get(context.Background(), fmt.Sprintf("%v", foreignKeyValue))
				if err != nil {
					return nil, err
				}
				return relatedObj, nil
			}
			return nil, fmt.Errorf("foreign key field %s not found in map", foreignKeyField)
		}
	}

	// For HasOne, get the field directly
	if relation.Type == resource.RelationTypeOneToOne {
		return utils.GetFieldValue(parentObj, relation.Field)
	}

	// For BelongsTo, get the foreign key and load the related object
	if relation.Type == resource.RelationTypeManyToOne {
		// For BelongsTo, we assume a foreign key field with the same name as the relation plus "ID"
		foreignKeyField := relation.Name + "ID"

		foreignKeyValue, err := utils.GetFieldValue(parentObj, foreignKeyField)
		if err != nil {
			return nil, err
		}

		// If nil or zero, return nil
		if reflect.ValueOf(foreignKeyValue).IsZero() {
			return nil, nil
		}

		// Get the related object using empty query options
		relatedObj, err := repo.Get(context.Background(), fmt.Sprintf("%v", foreignKeyValue))
		if err != nil {
			return nil, err
		}

		return relatedObj, nil
	}

	return nil, fmt.Errorf("unsupported relation type: %s", relation.Type)
}
