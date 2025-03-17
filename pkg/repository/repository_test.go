package repository

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestModel for repository tests
type TestModel struct {
	ID    string `json:"id" gorm:"primaryKey"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// TestModelWithCustomID for custom ID field tests
type TestModelWithCustomID struct {
	UID   string `json:"uid" gorm:"primaryKey"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func TestGormRepository(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate model
	err = db.AutoMigrate(&TestModel{})
	assert.NoError(t, err)

	// Create repository
	repo := NewGormRepository(db, &TestModel{})

	// Create a test resource
	res := resource.NewResource(resource.ResourceConfig{
		Name:  "tests",
		Model: TestModel{},
	})

	// Create a test context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	ctx := context.WithValue(c, "resource", res)

	// Test Create
	model := &TestModel{
		ID:    "1",
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	createdModel, err := repo.Create(ctx, model)
	assert.NoError(t, err)
	assert.Equal(t, model, createdModel)

	// Test Get
	retrievedModel, err := repo.Get(ctx, "1")
	assert.NoError(t, err)
	assert.Equal(t, model, retrievedModel)

	// Test List
	options := query.QueryOptions{
		Resource: res,
		Page:     1,
		PerPage:  10,
		Sort: &resource.Sort{
			Field: "id",
			Order: "asc",
		},
	}

	models, total, err := repo.List(ctx, options)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)

	// The List method returns a slice of the model type, not a slice of pointers to the model type
	modelsList, ok := models.(*[]TestModel)
	assert.True(t, ok)
	assert.Len(t, *modelsList, 1)
	assert.Equal(t, model.ID, (*modelsList)[0].ID)

	// Test Update
	model.Name = "Jane Doe"
	updatedModel, err := repo.Update(ctx, "1", model)
	assert.NoError(t, err)
	assert.Equal(t, model, updatedModel)

	// Verify update
	retrievedModel, err = repo.Get(ctx, "1")
	assert.NoError(t, err)
	assert.Equal(t, "Jane Doe", retrievedModel.(*TestModel).Name)

	// Test Delete
	err = repo.Delete(ctx, "1")
	assert.NoError(t, err)

	// Verify delete
	_, err = repo.Get(ctx, "1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
}

func TestGormRepositoryWithCustomID(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate model
	err = db.AutoMigrate(&TestModelWithCustomID{})
	assert.NoError(t, err)

	// Create a test resource with custom ID field
	res := resource.NewResource(resource.ResourceConfig{
		Name:        "tests",
		Model:       TestModelWithCustomID{},
		IDFieldName: "UID",
	})

	// Create repository with resource
	repo := NewGormRepositoryWithResource(db, res)

	// Create a test context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	ctx := context.WithValue(c, "resource", res)

	// Test Create
	model := &TestModelWithCustomID{
		UID:   "custom-1",
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	createdModel, err := repo.Create(ctx, model)
	assert.NoError(t, err)
	assert.Equal(t, model, createdModel)

	// Test Get
	retrievedModel, err := repo.Get(ctx, "custom-1")
	assert.NoError(t, err)
	assert.Equal(t, model, retrievedModel)

	// Test List
	options := query.QueryOptions{
		Resource: res,
		Page:     1,
		PerPage:  10,
		Sort: &resource.Sort{
			Field: "uid",
			Order: "asc",
		},
	}

	models, total, err := repo.List(ctx, options)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)

	// The List method returns a slice of the model type, not a slice of pointers to the model type
	modelsList, ok := models.(*[]TestModelWithCustomID)
	assert.True(t, ok)
	assert.Len(t, *modelsList, 1)
	assert.Equal(t, model.UID, (*modelsList)[0].UID)

	// Test Update
	model.Name = "Jane Doe"
	updatedModel, err := repo.Update(ctx, "custom-1", model)
	assert.NoError(t, err)
	assert.Equal(t, model, updatedModel)

	// Verify update
	retrievedModel, err = repo.Get(ctx, "custom-1")
	assert.NoError(t, err)
	assert.Equal(t, "Jane Doe", retrievedModel.(*TestModelWithCustomID).Name)

	// Test Delete
	err = repo.Delete(ctx, "custom-1")
	assert.NoError(t, err)

	// Verify delete
	_, err = repo.Get(ctx, "custom-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
}

func TestGormRepositoryFactory(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Create factory
	factory := NewGormRepositoryFactory(db)

	// Create a test resource with custom ID field
	res := resource.NewResource(resource.ResourceConfig{
		Name:        "tests",
		Model:       TestModelWithCustomID{},
		IDFieldName: "UID",
	})

	// Create repository using factory
	repo := factory.CreateRepository(res)
	assert.NotNil(t, repo)

	// Sprawdź, czy repozytorium używa niestandardowego pola identyfikatora
	gormRepo, ok := repo.(*GormRepository)
	assert.True(t, ok)
	assert.Equal(t, "UID", gormRepo.IDFieldName)
}
