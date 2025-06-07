package resource

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// helper to setup in-memory database
func setupRelationValidatorDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	return db
}

// simple models used in tests
type Category struct {
	ID   uint
	Name string
}

type Tag struct {
	ID   uint
	Name string
}

func TestValidateToOneRelationWithDB(t *testing.T) {
	db := setupRelationValidatorDB(t)
	assert.NoError(t, db.AutoMigrate(&Category{}))
	assert.NoError(t, db.Create(&Category{ID: 1, Name: "cat"}).Error)

	relation := Relation{
		Name:           "category",
		Type:           RelationTypeManyToOne,
		Resource:       "categories",
		ReferenceField: "ID",
	}
	related := &DefaultResource{Name: "categories", Model: Category{}}
	validator := RelationValidator{Relation: relation, DB: db, RelatedResource: related}

	err := validator.validateToOneRelation(reflect.ValueOf(uint(1)))
	assert.NoError(t, err)

	err = validator.validateToOneRelation(reflect.ValueOf(uint(2)))
	assert.Error(t, err)

	cat := Category{ID: 1}
	err = validator.validateToOneRelation(reflect.ValueOf(cat))
	assert.NoError(t, err)
}

func TestValidateToManyRelationWithDB(t *testing.T) {
	db := setupRelationValidatorDB(t)
	assert.NoError(t, db.AutoMigrate(&Tag{}))
	assert.NoError(t, db.Create(&Tag{ID: 1, Name: "t1"}).Error)
	assert.NoError(t, db.Create(&Tag{ID: 2, Name: "t2"}).Error)

	relation := Relation{
		Name:           "tags",
		Type:           RelationTypeOneToMany,
		Resource:       "tags",
		ReferenceField: "ID",
	}
	related := &DefaultResource{Name: "tags", Model: Tag{}}
	validator := RelationValidator{Relation: relation, DB: db, RelatedResource: related}

	err := validator.validateToManyRelation(reflect.ValueOf([]uint{1, 2}))
	assert.NoError(t, err)

	err = validator.validateToManyRelation(reflect.ValueOf([]uint{1, 3}))
	assert.Error(t, err)

	tags := []Tag{{ID: 1}, {ID: 2}}
	err = validator.validateToManyRelation(reflect.ValueOf(tags))
	assert.NoError(t, err)

	err = validator.validateToManyRelation(reflect.ValueOf(5))
	assert.Error(t, err)
}
