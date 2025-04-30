package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stanxing/refine-gin/pkg/repository"
	"github.com/stanxing/refine-gin/pkg/resource"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates a new in-memory database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	// Setup in-memory database with a unique name for each test
	dbName := fmt.Sprintf("file::memory:%p?cache=shared", t)
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate models
	err = db.AutoMigrate(&Category{}, &Post{}, &Tag{})
	assert.NoError(t, err)

	return db
}

// createTestData creates test data for relation tests
func createTestData(t *testing.T, db *gorm.DB) (Category, Post, Tag) {
	// Create a category
	category := Category{
		Name:      "Technology",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := db.Create(&category).Error
	assert.NoError(t, err)

	// Create a tag
	tag := Tag{
		Name:      "Go",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = db.Create(&tag).Error
	assert.NoError(t, err)

	// Create a post
	post := Post{
		Title:      "Testowy post",
		Content:    "Treść testowego posta",
		CategoryID: &category.ID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err = db.Create(&post).Error
	assert.NoError(t, err)

	// Create post-tag relation
	err = db.Exec("INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?)", post.ID, tag.ID).Error
	assert.NoError(t, err)

	return category, post, tag
}

// TestCategoryRepository_Get tests retrieving a category with relations
func TestCategoryRepository_Get(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	category, post, _ := createTestData(t, db)

	// Create resource
	categoryResource := resource.NewResource(resource.ResourceConfig{
		Name:  "categories",
		Model: Category{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
		},
		Relations: []resource.Relation{
			{
				Name:  "posts",
				Type:  resource.RelationTypeOneToMany,
				Field: "Posts",
			},
		},
	})

	// Create repository
	repo := repository.NewGenericRepositoryWithResource(db, categoryResource)

	// Test: Get category with posts
	result, err := repo.GetWithRelations(context.Background(), category.ID, []string{"Posts"})
	assert.NoError(t, err)

	// Check if category was retrieved correctly
	resultCategory, ok := result.(*Category)
	assert.True(t, ok)
	assert.Equal(t, category.ID, resultCategory.ID)
	assert.Equal(t, category.Name, resultCategory.Name)

	// Check that Posts were loaded
	assert.NotEmpty(t, resultCategory.Posts)
	assert.Equal(t, post.ID, resultCategory.Posts[0].ID)
}

// TestPostRepository_Get tests retrieving a post with relations
func TestPostRepository_Get(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	category, post, tag := createTestData(t, db)

	// Create resource
	postResource := resource.NewResource(resource.ResourceConfig{
		Name:  "posts",
		Model: Post{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
		},
		Relations: []resource.Relation{
			{
				Name:  "category",
				Type:  resource.RelationTypeManyToOne,
				Field: "Category",
			},
			{
				Name:  "tags",
				Type:  resource.RelationTypeManyToMany,
				Field: "Tags",
			},
		},
	})

	// Create repository
	repo := repository.NewGenericRepositoryWithResource(db, postResource)

	// Test: Get post with category and tags
	result, err := repo.GetWithRelations(context.Background(), post.ID, []string{"Category", "Tags"})
	assert.NoError(t, err)

	// Check if post was retrieved correctly
	resultPost, ok := result.(*Post)
	assert.True(t, ok)
	assert.Equal(t, post.ID, resultPost.ID)
	assert.Equal(t, post.Title, resultPost.Title)

	// Check that Category was loaded
	assert.NotNil(t, resultPost.Category)
	assert.Equal(t, category.ID, resultPost.Category.ID)

	// Check that Tags were loaded
	assert.NotEmpty(t, resultPost.Tags)
	assert.Equal(t, tag.ID, resultPost.Tags[0].ID)
}

// TestRelationExtraction tests extracting relations from models
func TestRelationExtraction(t *testing.T) {
	// Create resources
	categoryResource := resource.NewResource(resource.ResourceConfig{
		Name:  "categories",
		Model: Category{},
		Relations: []resource.Relation{
			{
				Name:  "posts",
				Type:  resource.RelationTypeOneToMany,
				Field: "Posts",
			},
		},
	})

	postResource := resource.NewResource(resource.ResourceConfig{
		Name:  "posts",
		Model: Post{},
		Relations: []resource.Relation{
			{
				Name:  "category",
				Type:  resource.RelationTypeManyToOne,
				Field: "Category",
			},
			{
				Name:  "tags",
				Type:  resource.RelationTypeManyToMany,
				Field: "Tags",
			},
		},
	})

	// Test category relations
	categoryRelations := categoryResource.GetRelations()
	assert.Len(t, categoryRelations, 1)
	assert.Equal(t, "posts", categoryRelations[0].Name)
	assert.Equal(t, resource.RelationTypeOneToMany, categoryRelations[0].Type)

	// Test post relations
	postRelations := postResource.GetRelations()
	assert.Len(t, postRelations, 2)

	categoryRelation := findRelationByName(postRelations, "category")
	assert.NotNil(t, categoryRelation)
	assert.Equal(t, resource.RelationTypeManyToOne, categoryRelation.Type)

	tagsRelation := findRelationByName(postRelations, "tags")
	assert.NotNil(t, tagsRelation)
	assert.Equal(t, resource.RelationTypeManyToMany, tagsRelation.Type)
}

func findRelationByName(relations []resource.Relation, name string) *resource.Relation {
	for _, r := range relations {
		if r.Name == name {
			return &r
		}
	}
	return nil
}
