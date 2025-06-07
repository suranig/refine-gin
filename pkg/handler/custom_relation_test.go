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
