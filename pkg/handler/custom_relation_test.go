package handler

import (
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

func TestGetRelatedObjectBelongsTo(t *testing.T) {
	rel := resource.Relation{Name: "Owner", Field: "Owner", Type: resource.RelationTypeManyToOne}

	t.Run("Struct", func(t *testing.T) {
		parent := &RelationParent{OwnerID: 10}
		repo := new(MockRepository)
		repo.On("Get", mock.Anything, "10").Return(&RelationChild{ID: 10}, nil).Once()

		result, err := getRelatedObject(parent, &rel, repo)
		assert.NoError(t, err)
		if assert.NotNil(t, result) {
			assert.Equal(t, uint(10), result.(*RelationChild).ID)
		}

		repo.AssertExpectations(t)
	})

	t.Run("Map", func(t *testing.T) {
		m := map[string]interface{}{"OwnerID": uint(20)}
		repo := new(MockRepository)
		repo.On("Get", mock.Anything, "20").Return(&RelationChild{ID: 20}, nil).Once()

		result, err := getRelatedObject(m, &rel, repo)
		assert.NoError(t, err)
		if assert.NotNil(t, result) {
			assert.Equal(t, uint(20), result.(*RelationChild).ID)
		}

		repo.AssertExpectations(t)
	})

	t.Run("ForeignKeyMissing", func(t *testing.T) {
		m := map[string]interface{}{}
		_, err := getRelatedObject(m, &rel, nil)
		assert.Error(t, err)
	})
}

func TestGetRelatedObjectHasOne(t *testing.T) {
	rel := resource.Relation{Name: "profile", Field: "Profile", Type: resource.RelationTypeOneToOne}

	t.Run("Struct", func(t *testing.T) {
		parent := &RelationParent{Profile: &RelationChild{ID: 3}}
		result, err := getRelatedObject(parent, &rel, nil)
		assert.NoError(t, err)
		assert.Equal(t, parent.Profile, result)
	})

	t.Run("Map", func(t *testing.T) {
		m := map[string]interface{}{"Profile": &RelationChild{ID: 4}}
		result, err := getRelatedObject(m, &rel, nil)
		assert.NoError(t, err)
		assert.Equal(t, m["Profile"], result)
	})

	t.Run("FieldMissingStruct", func(t *testing.T) {
		parent := &RelationParent{}
		_, err := getRelatedObject(parent, &rel, nil)
		assert.Error(t, err)
	})

	t.Run("FieldMissingMap", func(t *testing.T) {
		m := map[string]interface{}{}
		_, err := getRelatedObject(m, &rel, nil)
		assert.Error(t, err)
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		badRel := resource.Relation{Name: "tags", Field: "Tags", Type: resource.RelationTypeManyToMany}
		_, err := getRelatedObject(&RelationParent{}, &badRel, nil)
		assert.Error(t, err)
	})
}
