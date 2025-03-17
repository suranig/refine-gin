package main

import (
	"context"
	"fmt"
	"net/http"
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
	err = db.AutoMigrate(&User{}, &Post{}, &Comment{}, &Profile{})
	assert.NoError(t, err)

	return db
}

// createTestData creates test data with unique IDs based on the test name
func createTestData(t *testing.T, db *gorm.DB) (User, Post, Comment, Profile) {
	// Generate unique IDs based on test name
	prefix := fmt.Sprintf("%p", t)
	userID := fmt.Sprintf("user_%s", prefix)
	profileID := fmt.Sprintf("profile_%s", prefix)
	postID := fmt.Sprintf("post_%s", prefix)
	commentID := fmt.Sprintf("comment_%s", prefix)

	// Create a user
	user := User{
		ID:        userID,
		Name:      "Jan Kowalski",
		Email:     fmt.Sprintf("jan_%s@example.com", prefix),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := db.Create(&user).Error
	assert.NoError(t, err)

	// Create a profile
	profile := Profile{
		ID:        profileID,
		Bio:       "Programista Go",
		UserID:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = db.Create(&profile).Error
	assert.NoError(t, err)

	// Create a post
	post := Post{
		ID:        postID,
		Title:     "Testowy post",
		Content:   "Treść testowego posta",
		AuthorID:  user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = db.Create(&post).Error
	assert.NoError(t, err)

	// Create a comment
	comment := Comment{
		ID:        commentID,
		Content:   "Testowy komentarz",
		AuthorID:  user.ID,
		PostID:    post.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = db.Create(&comment).Error
	assert.NoError(t, err)

	return user, post, comment, profile
}

// MockUserRepository is a modified version of UserRepository for testing
type MockUserRepository struct {
	db *gorm.DB
}

// Get retrieves a user by ID with relations
func (r *MockUserRepository) Get(ctx context.Context, id interface{}) (interface{}, error) {
	var user User

	// Get the resource from context
	res, ok := ctx.Value("resource").(resource.Resource)
	if !ok {
		return nil, fmt.Errorf("resource not found in context")
	}

	// Get includes directly from the resource (for testing)
	var includes []string
	for _, relation := range res.GetRelations() {
		if relation.IncludeByDefault {
			includes = append(includes, relation.Name)
		}
	}

	// Apply includes
	q := r.db
	for _, include := range includes {
		q = q.Preload(include)
	}

	if err := q.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// List retrieves users with pagination and relations
func (r *MockUserRepository) List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error) {
	var users []User
	var total int64

	q := r.db.Model(&User{})

	// Apply includes from the resource
	var includes []string
	for _, relation := range options.Resource.GetRelations() {
		if relation.IncludeByDefault {
			includes = append(includes, relation.Name)
		}
	}

	for _, include := range includes {
		q = q.Preload(include)
	}

	total, err := options.ApplyWithPagination(q, &users)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func TestUserRepository_Get(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	user, _, _, _ := createTestData(t, db)
	repo := &MockUserRepository{db: db}

	// Create resource
	userResource := resource.NewResource(resource.ResourceConfig{
		Name:  "users",
		Model: User{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
		},
	})

	// Create Gin context
	gin.SetMode(gin.TestMode)
	ctx := context.WithValue(context.Background(), "resource", userResource)

	// Test: Get user without additional relations
	result, err := repo.Get(ctx, user.ID)
	assert.NoError(t, err)

	// Check if user was retrieved correctly
	resultUser, ok := result.(User)
	assert.True(t, ok)
	assert.Equal(t, user.ID, resultUser.ID)
	assert.Equal(t, user.Name, resultUser.Name)
	assert.Equal(t, user.Email, resultUser.Email)

	// Profile should be loaded by default (include=true)
	assert.NotNil(t, resultUser.Profile)
	assert.Equal(t, "Programista Go", resultUser.Profile.Bio)

	// Posts should not be loaded by default (include=false)
	assert.Empty(t, resultUser.Posts)
}

func TestUserRepository_List(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	user, _, _, _ := createTestData(t, db)
	repo := &MockUserRepository{db: db}

	// Create resource
	userResource := resource.NewResource(resource.ResourceConfig{
		Name:  "users",
		Model: User{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
		},
	})

	// Create Gin context
	gin.SetMode(gin.TestMode)

	// Create query options
	options := query.QueryOptions{
		Resource: userResource,
		Filters:  []query.Filter{},
		Sort:     nil,
		Page:     1,
		PerPage:  10,
	}

	// Test: List users without additional relations
	result, total, err := repo.List(context.Background(), options)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)

	// Check if users were retrieved correctly
	resultUsers, ok := result.([]User)
	assert.True(t, ok)
	assert.Len(t, resultUsers, 1)
	assert.Equal(t, user.ID, resultUsers[0].ID)
	assert.Equal(t, user.Name, resultUsers[0].Name)
	assert.Equal(t, user.Email, resultUsers[0].Email)

	// Profile should be loaded by default (include=true)
	assert.NotNil(t, resultUsers[0].Profile)
	assert.Equal(t, "Programista Go", resultUsers[0].Profile.Bio)

	// Posts should not be loaded by default (include=false)
	assert.Empty(t, resultUsers[0].Posts)
}

// MockPostRepository is a simplified repository for posts
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
		if relation.IncludeByDefault {
			includes = append(includes, relation.Name)
		}
	}

	// Apply includes
	q := r.db
	for _, include := range includes {
		q = q.Preload(include)
	}

	if err := q.First(&post, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return post, nil
}

func TestPostRepository_Get(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	_, post, _, _ := createTestData(t, db)
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
	})

	// Create Gin context
	gin.SetMode(gin.TestMode)
	ctx := context.WithValue(context.Background(), "resource", postResource)

	// Test: Get post without additional relations
	result, err := repo.Get(ctx, post.ID)
	assert.NoError(t, err)

	// Check if post was retrieved correctly
	resultPost, ok := result.(Post)
	assert.True(t, ok)
	assert.Equal(t, post.ID, resultPost.ID)
	assert.Equal(t, post.Title, resultPost.Title)
	assert.Equal(t, post.Content, resultPost.Content)

	// Author should be loaded by default (include=true)
	assert.NotEmpty(t, resultPost.Author)
	assert.Equal(t, post.AuthorID, resultPost.Author.ID)

	// Comments should not be loaded by default (include=false)
	assert.Empty(t, resultPost.Comments)
}

func TestRelationExtraction(t *testing.T) {
	// Test relation extraction from User model
	userRelations := resource.ExtractRelationsFromModel(User{})

	// Should have 2 relations: Posts and Profile
	assert.Len(t, userRelations, 2)

	// Check Posts relation
	postsRelation := findRelationByName(userRelations, "Posts")
	assert.NotNil(t, postsRelation)
	assert.Equal(t, "Posts", postsRelation.Name)
	assert.Equal(t, resource.RelationTypeOneToMany, postsRelation.Type)
	assert.Equal(t, "posts", postsRelation.Resource)
	assert.Equal(t, "author_id", postsRelation.Field)
	assert.Equal(t, "id", postsRelation.ReferenceField)
	assert.False(t, postsRelation.IncludeByDefault)

	// Check Profile relation
	profileRelation := findRelationByName(userRelations, "Profile")
	assert.NotNil(t, profileRelation)
	assert.Equal(t, "Profile", profileRelation.Name)
	assert.Equal(t, resource.RelationTypeOneToOne, profileRelation.Type)
	assert.Equal(t, "profiles", profileRelation.Resource)
	assert.Equal(t, "user_id", profileRelation.Field)
	assert.Equal(t, "id", profileRelation.ReferenceField)
	assert.True(t, profileRelation.IncludeByDefault)

	// Test relation extraction from Post model
	postRelations := resource.ExtractRelationsFromModel(Post{})

	// Should have 2 relations: Author and Comments
	assert.Len(t, postRelations, 2)

	// Check Author relation
	authorRelation := findRelationByName(postRelations, "Author")
	assert.NotNil(t, authorRelation)
	assert.Equal(t, "Author", authorRelation.Name)
	assert.Equal(t, resource.RelationTypeManyToOne, authorRelation.Type)
	assert.Equal(t, "users", authorRelation.Resource)
	assert.Equal(t, "author_id", authorRelation.Field)
	assert.Equal(t, "id", authorRelation.ReferenceField)
	assert.True(t, authorRelation.IncludeByDefault)
}

// Helper function to find relation by name
func findRelationByName(relations []resource.Relation, name string) *resource.Relation {
	for _, relation := range relations {
		if relation.Name == name {
			return &relation
		}
	}
	return nil
}

func TestIncludeRelations(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	_, _, _, _ = createTestData(t, db)

	// Create resource
	userResource := resource.NewResource(resource.ResourceConfig{
		Name:  "users",
		Model: User{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
		},
	})

	// Create Gin context
	gin.SetMode(gin.TestMode)

	// Test: Without include parameter (should return default relations)
	c1, _ := gin.CreateTestContext(nil)
	includes1 := resource.IncludeRelations(c1, userResource)

	// Profile should be loaded by default (include=true)
	assert.Contains(t, includes1, "Profile")
	// Posts should not be loaded by default (include=false)
	assert.NotContains(t, includes1, "Posts")

	// Test: With include parameter
	c2, _ := gin.CreateTestContext(nil)
	req, _ := http.NewRequest("GET", "/?include=Posts,Profile", nil)
	c2.Request = req
	includes2 := resource.IncludeRelations(c2, userResource)

	// Both relations should be loaded
	assert.Contains(t, includes2, "Posts")
	assert.Contains(t, includes2, "Profile")
}

func TestLoadRelations(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	user, post, _, _ := createTestData(t, db)

	// Test: Load relations for user
	var loadedUser User
	err := db.First(&loadedUser, "id = ?", user.ID).Error
	assert.NoError(t, err)

	// Load Posts and Profile relations
	err = resource.LoadRelations(db, resource.NewResource(resource.ResourceConfig{
		Name:  "users",
		Model: User{},
	}), &loadedUser, []string{"Posts", "Profile"})
	assert.NoError(t, err)

	// Check if relations were loaded
	assert.NotNil(t, loadedUser.Profile)
	assert.Equal(t, "Programista Go", loadedUser.Profile.Bio)
	assert.Len(t, loadedUser.Posts, 1)
	assert.Equal(t, post.ID, loadedUser.Posts[0].ID)

	// Test: Load relations for post
	var loadedPost Post
	err = db.First(&loadedPost, "id = ?", post.ID).Error
	assert.NoError(t, err)

	// Load Author and Comments relations
	err = resource.LoadRelations(db, resource.NewResource(resource.ResourceConfig{
		Name:  "posts",
		Model: Post{},
	}), &loadedPost, []string{"Author", "Comments"})
	assert.NoError(t, err)

	// Check if relations were loaded
	assert.Equal(t, user.ID, loadedPost.Author.ID)
	assert.Equal(t, user.Name, loadedPost.Author.Name)
	assert.Len(t, loadedPost.Comments, 1)
	assert.Equal(t, "Testowy komentarz", loadedPost.Comments[0].Content)
}
