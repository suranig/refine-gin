package handler

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/suranig/refine-gin/pkg/resource"
)

// Models used for relation tests

type RelationChild struct {
	ID   uint
	Name string
}

type RelationParent struct {
	ID      uint
	Items   []*RelationChild
	Profile *RelationChild
	OwnerID uint
	Tags    []uint
}

func TestAttachAndDetachRelations(t *testing.T) {
	// setup common relations
	hasManyRel := resource.Relation{Name: "items", Field: "Items", Type: resource.RelationTypeOneToMany}
	hasOneRel := resource.Relation{Name: "profile", Field: "Profile", Type: resource.RelationTypeOneToOne}
	belongsToRel := resource.Relation{Name: "Owner", Field: "Owner", Type: resource.RelationTypeManyToOne}
	manyToManyRel := resource.Relation{Name: "tags", Field: "Tags", Type: resource.RelationTypeManyToMany}

	t.Run("HasMany", func(t *testing.T) {
		parent := &RelationParent{}
		repo := new(MockRepository)
		repo.On("Get", mock.Anything, "1").Return(&RelationChild{ID: 1, Name: "c1"}, nil).Once()
		repo.On("Get", mock.Anything, "2").Return(&RelationChild{ID: 2, Name: "c2"}, nil).Once()

		err := attachToHasManyRelation(parent, &hasManyRel, []interface{}{uint(1), uint(2)}, repo)
		assert.NoError(t, err)
		assert.Len(t, parent.Items, 2)
		assert.Equal(t, uint(1), parent.Items[0].ID)
		assert.Equal(t, uint(2), parent.Items[1].ID)

		err = detachFromHasManyRelation(parent, &hasManyRel, []interface{}{uint(1)})
		assert.NoError(t, err)
		assert.Len(t, parent.Items, 1)
		assert.Equal(t, uint(2), parent.Items[0].ID)

		repo.AssertExpectations(t)
	})

	t.Run("HasOne", func(t *testing.T) {
		parent := &RelationParent{}
		repo := new(MockRepository)
		repo.On("Get", mock.Anything, "3").Return(&RelationChild{ID: 3, Name: "p"}, nil).Once()

		err := attachToHasOneRelation(parent, &hasOneRel, 3, repo)
		assert.NoError(t, err)
		if assert.NotNil(t, parent.Profile) {
			assert.Equal(t, uint(3), parent.Profile.ID)
		}

		err = detachFromHasOneRelation(parent, &hasOneRel)
		assert.NoError(t, err)
		assert.Nil(t, parent.Profile)

		repo.AssertExpectations(t)
	})

	t.Run("BelongsTo", func(t *testing.T) {
		parent := &RelationParent{}
		repo := new(MockRepository)
		repo.On("Get", mock.Anything, "5").Return(&RelationChild{ID: 5}, nil).Once()

		err := attachToBelongsToRelation(parent, &belongsToRel, 5, repo)
		assert.NoError(t, err)
		assert.Equal(t, uint(5), parent.OwnerID)

		err = detachFromBelongsToRelation(parent, &belongsToRel)
		assert.NoError(t, err)
		assert.Equal(t, uint(0), parent.OwnerID)

		repo.AssertExpectations(t)
	})

	t.Run("ManyToMany", func(t *testing.T) {
		parent := &RelationParent{Tags: []uint{1}}

		err := attachToManyToManyRelation(parent, &manyToManyRel, []interface{}{uint(1), uint(2)}, nil, nil, "")
		assert.NoError(t, err)
		assert.ElementsMatch(t, []uint{1, 2}, parent.Tags)

		err = detachFromManyToManyRelation(parent, &manyToManyRel, []interface{}{uint(1)}, nil, nil, "")
		assert.NoError(t, err)
		assert.ElementsMatch(t, []uint{2}, parent.Tags)
	})
}

func TestGetRelatedCollection(t *testing.T) {
	rel := resource.Relation{Name: "items", Field: "Items", Type: resource.RelationTypeOneToMany}

	t.Run("StructSlice", func(t *testing.T) {
		parent := &RelationParent{Items: []*RelationChild{{ID: 1}, {ID: 2}}}
		result, err := getRelatedCollection(parent, &rel, nil)
		assert.NoError(t, err)
		assert.Equal(t, parent.Items, result)
	})

	t.Run("MapSlice", func(t *testing.T) {
		m := map[string]interface{}{"Items": []int{1, 2}}
		result, err := getRelatedCollection(m, &rel, nil)
		assert.NoError(t, err)
		assert.Equal(t, m["Items"], result)
	})

	t.Run("FieldNotFound", func(t *testing.T) {
		parent := &RelationParent{}
		badRel := resource.Relation{Name: "missing", Field: "Missing", Type: resource.RelationTypeOneToMany}
		_, err := getRelatedCollection(parent, &badRel, nil)
		assert.Error(t, err)
	})

	t.Run("NotSlice", func(t *testing.T) {
		parent := &RelationParent{Profile: &RelationChild{ID: 1}}
		badRel := resource.Relation{Name: "profile", Field: "Profile", Type: resource.RelationTypeOneToMany}
		_, err := getRelatedCollection(parent, &badRel, nil)
		assert.Error(t, err)
	})
}

func TestGetRelatedObject_Complete(t *testing.T) {
	// Test comprehensive coverage of getRelatedObject function for 100% coverage

	t.Run("Map_OneToOne_FieldExists", func(t *testing.T) {
		relation := resource.Relation{
			Name:  "profile",
			Field: "Profile",
			Type:  resource.RelationTypeOneToOne,
		}

		parentMap := map[string]interface{}{
			"Profile": &RelationChild{ID: 123, Name: "Profile Data"},
		}

		result, err := getRelatedObject(parentMap, &relation, nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		profile := result.(*RelationChild)
		assert.Equal(t, uint(123), profile.ID)
		assert.Equal(t, "Profile Data", profile.Name)
	})

	t.Run("Map_OneToOne_FieldNotFound", func(t *testing.T) {
		relation := resource.Relation{
			Name:  "profile",
			Field: "Profile",
			Type:  resource.RelationTypeOneToOne,
		}

		parentMap := map[string]interface{}{
			"OtherField": "value",
		}

		result, err := getRelatedObject(parentMap, &relation, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "field Profile not found in map")
		assert.Nil(t, result)
	})

	t.Run("Map_ManyToOne_ForeignKeyExists", func(t *testing.T) {
		relation := resource.Relation{
			Name:  "Owner",
			Field: "Owner",
			Type:  resource.RelationTypeManyToOne,
		}

		parentMap := map[string]interface{}{
			"OwnerID": uint(456),
		}

		mockRepo := new(MockRepository)
		expectedOwner := &RelationChild{ID: 456, Name: "Owner Data"}
		mockRepo.On("Get", mock.Anything, "456").Return(expectedOwner, nil)

		result, err := getRelatedObject(parentMap, &relation, mockRepo)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		owner := result.(*RelationChild)
		assert.Equal(t, uint(456), owner.ID)
		assert.Equal(t, "Owner Data", owner.Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Map_ManyToOne_ForeignKeyNil", func(t *testing.T) {
		relation := resource.Relation{
			Name:  "Owner",
			Field: "Owner",
			Type:  resource.RelationTypeManyToOne,
		}

		parentMap := map[string]interface{}{
			"OwnerID": nil,
		}

		result, err := getRelatedObject(parentMap, &relation, nil)

		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Map_ManyToOne_ForeignKeyZero", func(t *testing.T) {
		relation := resource.Relation{
			Name:  "Owner",
			Field: "Owner",
			Type:  resource.RelationTypeManyToOne,
		}

		parentMap := map[string]interface{}{
			"OwnerID": 0,
		}

		result, err := getRelatedObject(parentMap, &relation, nil)

		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Map_ManyToOne_ForeignKeyNotFound", func(t *testing.T) {
		relation := resource.Relation{
			Name:  "Owner",
			Field: "Owner",
			Type:  resource.RelationTypeManyToOne,
		}

		parentMap := map[string]interface{}{
			"OtherField": "value",
		}

		result, err := getRelatedObject(parentMap, &relation, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "foreign key field OwnerID not found in map")
		assert.Nil(t, result)
	})

	t.Run("Map_ManyToOne_RepositoryError", func(t *testing.T) {
		relation := resource.Relation{
			Name:  "Owner",
			Field: "Owner",
			Type:  resource.RelationTypeManyToOne,
		}

		parentMap := map[string]interface{}{
			"OwnerID": uint(999),
		}

		mockRepo := new(MockRepository)
		mockRepo.On("Get", mock.Anything, "999").Return(nil, fmt.Errorf("not found"))

		result, err := getRelatedObject(parentMap, &relation, mockRepo)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Struct_OneToOne_Success", func(t *testing.T) {
		relation := resource.Relation{
			Name:  "profile",
			Field: "Profile",
			Type:  resource.RelationTypeOneToOne,
		}

		parent := &RelationParent{
			Profile: &RelationChild{ID: 789, Name: "Struct Profile"},
		}

		result, err := getRelatedObject(parent, &relation, nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		profile := result.(*RelationChild)
		assert.Equal(t, uint(789), profile.ID)
		assert.Equal(t, "Struct Profile", profile.Name)
	})

	t.Run("Struct_ManyToOne_Success", func(t *testing.T) {
		relation := resource.Relation{
			Name:  "Owner",
			Field: "Owner",
			Type:  resource.RelationTypeManyToOne,
		}

		parent := &RelationParent{
			OwnerID: 101,
		}

		mockRepo := new(MockRepository)
		expectedOwner := &RelationChild{ID: 101, Name: "Struct Owner"}
		mockRepo.On("Get", mock.Anything, "101").Return(expectedOwner, nil)

		result, err := getRelatedObject(parent, &relation, mockRepo)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		owner := result.(*RelationChild)
		assert.Equal(t, uint(101), owner.ID)
		assert.Equal(t, "Struct Owner", owner.Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Struct_ManyToOne_ZeroForeignKey", func(t *testing.T) {
		relation := resource.Relation{
			Name:  "Owner",
			Field: "Owner",
			Type:  resource.RelationTypeManyToOne,
		}

		parent := &RelationParent{
			OwnerID: 0, // Zero value
		}

		result, err := getRelatedObject(parent, &relation, nil)

		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Struct_ManyToOne_FieldNotFound", func(t *testing.T) {
		relation := resource.Relation{
			Name:  "NonExistent",
			Field: "NonExistent",
			Type:  resource.RelationTypeManyToOne,
		}

		parent := &RelationParent{}

		result, err := getRelatedObject(parent, &relation, nil)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Struct_ManyToOne_RepositoryError", func(t *testing.T) {
		relation := resource.Relation{
			Name:  "Owner",
			Field: "Owner",
			Type:  resource.RelationTypeManyToOne,
		}

		parent := &RelationParent{
			OwnerID: 999,
		}

		mockRepo := new(MockRepository)
		mockRepo.On("Get", mock.Anything, "999").Return(nil, fmt.Errorf("repository error"))

		result, err := getRelatedObject(parent, &relation, mockRepo)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository error")
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("UnsupportedRelationType", func(t *testing.T) {
		relation := resource.Relation{
			Name:  "unsupported",
			Field: "Unsupported",
			Type:  "UnsupportedType",
		}

		parent := &RelationParent{}

		result, err := getRelatedObject(parent, &relation, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported relation type")
		assert.Nil(t, result)
	})
}
