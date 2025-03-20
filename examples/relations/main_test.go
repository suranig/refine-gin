package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
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

// MockCategoryRepository is a repository for Category testing
type MockCategoryRepository struct {
	db *gorm.DB
}

// Get retrieves a category by ID with relations
func (r *MockCategoryRepository) Get(ctx context.Context, id interface{}) (interface{}, error) {
	var category Category

	// Get the resource from context
	res, ok := ctx.Value("resource").(resource.Resource)
	if !ok {
		return nil, fmt.Errorf("resource not found in context")
	}

	// Get includes directly from the resource (for testing)
	var includes []string
	for _, relation := range res.GetRelations() {
		includes = append(includes, relation.Field)
	}

	// Apply includes
	q := r.db
	for _, include := range includes {
		q = q.Preload(include)
	}

	if err := q.First(&category, id).Error; err != nil {
		return nil, err
	}

	return category, nil
}

// List retrieves categories with pagination and relations
func (r *MockCategoryRepository) List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error) {
	var categories []Category
	var total int64

	q := r.db.Model(&Category{})

	// Apply includes from the resource
	var includes []string
	for _, relation := range options.Resource.GetRelations() {
		includes = append(includes, relation.Field)
	}

	for _, include := range includes {
		q = q.Preload(include)
	}

	total, err := options.ApplyWithPagination(q, &categories)
	if err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

// TestCategoryRepository_Get tests retrieving a category with relations
func TestCategoryRepository_Get(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	category, post, _ := createTestData(t, db)
	repo := &MockCategoryRepository{db: db}

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

	// Create Gin context
	gin.SetMode(gin.TestMode)
	ctx := context.WithValue(context.Background(), "resource", categoryResource)

	// Test: Get category with posts
	result, err := repo.Get(ctx, category.ID)
	assert.NoError(t, err)

	// Check if category was retrieved correctly
	resultCategory, ok := result.(Category)
	assert.True(t, ok)
	assert.Equal(t, category.ID, resultCategory.ID)
	assert.Equal(t, category.Name, resultCategory.Name)

	// Check that Posts were loaded
	assert.NotEmpty(t, resultCategory.Posts)
	assert.Equal(t, post.ID, resultCategory.Posts[0].ID)
}

// MockPostRepository is a repository for Post testing
type MockPostRepository struct {
	db *gorm.DB
}

// Get retrieves a post by ID with relations
func (r *MockPostRepository) Get(ctx context.Context, id interface{}) (interface{}, error) {
	var post Post

	// Get the resource from context
	res, ok := ctx.Value("resource").(resource.Resource)
	if !ok {
		return nil, fmt.Errorf("resource not found in context")
	}

	// Get includes directly from the resource (for testing)
	var includes []string
	for _, relation := range res.GetRelations() {
		includes = append(includes, relation.Field)
	}

	// Apply includes
	q := r.db
	for _, include := range includes {
		q = q.Preload(include)
	}

	if err := q.First(&post, id).Error; err != nil {
		return nil, err
	}

	return post, nil
}

// TestPostRepository_Get tests retrieving a post with relations
func TestPostRepository_Get(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	category, post, tag := createTestData(t, db)
	repo := &MockPostRepository{db: db}

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

	// Create Gin context
	gin.SetMode(gin.TestMode)
	ctx := context.WithValue(context.Background(), "resource", postResource)

	// Test: Get post with category and tags
	result, err := repo.Get(ctx, post.ID)
	assert.NoError(t, err)

	// Check if post was retrieved correctly
	resultPost, ok := result.(Post)
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

// Test helper functions for relation extraction
func TestRelationExtraction(t *testing.T) {
	// Set up a resource with relations
	res := resource.NewResource(resource.ResourceConfig{
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

	// Test finding relation by name
	categoryRelation := findRelationByName(res.GetRelations(), "category")
	assert.NotNil(t, categoryRelation)
	assert.Equal(t, "category", categoryRelation.Name)
	assert.Equal(t, resource.RelationTypeManyToOne, categoryRelation.Type)

	tagsRelation := findRelationByName(res.GetRelations(), "tags")
	assert.NotNil(t, tagsRelation)
	assert.Equal(t, "tags", tagsRelation.Name)
	assert.Equal(t, resource.RelationTypeManyToMany, tagsRelation.Type)

	// Test with a non-existent relation
	nonExistentRelation := findRelationByName(res.GetRelations(), "comments")
	assert.Nil(t, nonExistentRelation)
}

// Helper to find a relation by name
func findRelationByName(relations []resource.Relation, name string) *resource.Relation {
	for _, relation := range relations {
		if relation.Name == name {
			return &relation
		}
	}
	return nil
}
