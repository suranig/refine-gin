package resource

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestModels for relation tests
type User struct {
	ID      string   `json:"id" gorm:"primaryKey"`
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Posts   []Post   `json:"posts" gorm:"foreignKey:AuthorID" relation:"resource=posts;type=one-to-many;field=author_id;reference=id;include=false"`
	Profile *Profile `json:"profile" gorm:"foreignKey:UserID" relation:"resource=profiles;type=one-to-one;field=user_id;reference=id;include=true"`
}

type Post struct {
	ID       string    `json:"id" gorm:"primaryKey"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	AuthorID string    `json:"author_id" gorm:"index"`
	Author   User      `json:"author" relation:"resource=users;type=many-to-one;field=author_id;reference=id;include=true"`
	Comments []Comment `json:"comments" relation:"resource=comments;type=one-to-many;field=post_id;reference=id;include=false"`
}

type Comment struct {
	ID       string `json:"id" gorm:"primaryKey"`
	Content  string `json:"content"`
	AuthorID string `json:"author_id" gorm:"index"`
	PostID   string `json:"post_id" gorm:"index"`
}

type Profile struct {
	ID     string `json:"id" gorm:"primaryKey"`
	Bio    string `json:"bio"`
	UserID string `json:"user_id" gorm:"uniqueIndex"`
}

func TestExtractRelationsFromModel(t *testing.T) {
	// Test extracting relations from User model
	userRelations := ExtractRelationsFromModel(User{})

	// Should have 2 relations: Posts and Profile
	assert.Len(t, userRelations, 2)

	// Check Posts relation
	postsRelation := findRelationByName(userRelations, "Posts")
	assert.NotNil(t, postsRelation)
	assert.Equal(t, "Posts", postsRelation.Name)
	assert.Equal(t, RelationTypeOneToMany, postsRelation.Type)
	assert.Equal(t, "posts", postsRelation.Resource)
	assert.Equal(t, "author_id", postsRelation.Field)
	assert.Equal(t, "id", postsRelation.ReferenceField)
	assert.False(t, postsRelation.IncludeByDefault)

	// Check Profile relation
	profileRelation := findRelationByName(userRelations, "Profile")
	assert.NotNil(t, profileRelation)
	assert.Equal(t, "Profile", profileRelation.Name)
	assert.Equal(t, RelationTypeOneToOne, profileRelation.Type)
	assert.Equal(t, "profiles", profileRelation.Resource)
	assert.Equal(t, "user_id", profileRelation.Field)
	assert.Equal(t, "id", profileRelation.ReferenceField)
	assert.True(t, profileRelation.IncludeByDefault)

	// Test extracting relations from Post model
	postRelations := ExtractRelationsFromModel(Post{})

	// Should have 2 relations: Author and Comments
	assert.Len(t, postRelations, 2)

	// Check Author relation
	authorRelation := findRelationByName(postRelations, "Author")
	assert.NotNil(t, authorRelation)
	assert.Equal(t, "Author", authorRelation.Name)
	assert.Equal(t, RelationTypeManyToOne, authorRelation.Type)
	assert.Equal(t, "users", authorRelation.Resource)
	assert.Equal(t, "author_id", authorRelation.Field)
	assert.Equal(t, "id", authorRelation.ReferenceField)
	assert.True(t, authorRelation.IncludeByDefault)

	// Check Comments relation
	commentsRelation := findRelationByName(postRelations, "Comments")
	assert.NotNil(t, commentsRelation)
	assert.Equal(t, "Comments", commentsRelation.Name)
	assert.Equal(t, RelationTypeOneToMany, commentsRelation.Type)
	assert.Equal(t, "comments", commentsRelation.Resource)
	assert.Equal(t, "post_id", commentsRelation.Field)
	assert.Equal(t, "id", commentsRelation.ReferenceField)
	assert.False(t, commentsRelation.IncludeByDefault)
}

func TestParseRelationTag(t *testing.T) {
	// Test parsing a complete relation tag
	tag := "resource=users;type=one-to-many;field=user_id;reference=id;include=true;name=custom_name"
	relation := parseRelationTag("DefaultName", tag)

	assert.NotNil(t, relation)
	assert.Equal(t, "custom_name", relation.Name)
	assert.Equal(t, RelationTypeOneToMany, relation.Type)
	assert.Equal(t, "users", relation.Resource)
	assert.Equal(t, "user_id", relation.Field)
	assert.Equal(t, "id", relation.ReferenceField)
	assert.True(t, relation.IncludeByDefault)

	// Test parsing a minimal relation tag
	tag = "resource=users;type=one-to-many"
	relation = parseRelationTag("DefaultName", tag)

	assert.NotNil(t, relation)
	assert.Equal(t, "DefaultName", relation.Name)
	assert.Equal(t, RelationTypeOneToMany, relation.Type)
	assert.Equal(t, "users", relation.Resource)
	assert.Equal(t, "", relation.Field)
	assert.Equal(t, "", relation.ReferenceField)
	assert.False(t, relation.IncludeByDefault)

	// Test parsing an invalid relation tag (missing required fields)
	tag = "field=user_id;reference=id"
	relation = parseRelationTag("DefaultName", tag)

	assert.Nil(t, relation)
}

func TestInferRelationFromField(t *testing.T) {
	// Get field from User struct
	userType := reflect.TypeOf(User{})

	// Test inferring one-to-many relation from Posts field
	postsField, _ := userType.FieldByName("Posts")
	relation := inferRelationFromField(postsField)

	assert.NotNil(t, relation)
	assert.Equal(t, "Posts", relation.Name)
	assert.Equal(t, RelationTypeOneToMany, relation.Type)
	assert.Equal(t, "Post", relation.Resource)

	// Test inferring one-to-one relation from Profile field
	profileField, _ := userType.FieldByName("Profile")
	relation = inferRelationFromField(profileField)

	assert.NotNil(t, relation)
	assert.Equal(t, "Profile", relation.Name)
	assert.Equal(t, RelationTypeOneToOne, relation.Type)
	assert.Equal(t, "Profile", relation.Resource)

	// Get field from Post struct
	postType := reflect.TypeOf(Post{})

	// Test inferring many-to-one relation from Author field
	authorField, _ := postType.FieldByName("Author")
	relation = inferRelationFromField(authorField)

	assert.NotNil(t, relation)
	assert.Equal(t, "Author", relation.Name)
	assert.Equal(t, RelationTypeOneToOne, relation.Type) // Default is one-to-one
	assert.Equal(t, "User", relation.Resource)
}

func TestIncludeRelations(t *testing.T) {
	// Create a mock gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	// Create a resource with relations
	res := &DefaultResource{
		Name: "users",
		Relations: []Relation{
			{
				Name:             "Posts",
				Type:             RelationTypeOneToMany,
				Resource:         "posts",
				IncludeByDefault: false,
			},
			{
				Name:             "Profile",
				Type:             RelationTypeOneToOne,
				Resource:         "profiles",
				IncludeByDefault: true,
			},
		},
	}

	// Test with no include parameter (should return default includes)
	includes := IncludeRelations(c, res)
	assert.Len(t, includes, 1)
	assert.Equal(t, "Profile", includes[0])

	// Test with include parameter
	c.Request = &http.Request{URL: &url.URL{RawQuery: "include=Posts,Profile"}}
	includes = IncludeRelations(c, res)
	assert.Len(t, includes, 2)
	assert.ElementsMatch(t, []string{"Posts", "Profile"}, includes)

	// Test with invalid include parameter
	c.Request = &http.Request{URL: &url.URL{RawQuery: "include=Invalid,Profile"}}
	includes = IncludeRelations(c, res)
	assert.Len(t, includes, 1)
	assert.Equal(t, "Profile", includes[0])
}

func TestLoadRelations(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate models
	err = db.AutoMigrate(&User{}, &Post{}, &Profile{}, &Comment{})
	assert.NoError(t, err)

	// Create test data
	user := User{
		ID:    "1",
		Name:  "John Doe",
		Email: "john@example.com",
	}

	profile := Profile{
		ID:     "1",
		Bio:    "Test bio",
		UserID: "1",
	}

	post1 := Post{
		ID:       "1",
		Title:    "Test Post 1",
		Content:  "Test content 1",
		AuthorID: "1",
	}

	post2 := Post{
		ID:       "2",
		Title:    "Test Post 2",
		Content:  "Test content 2",
		AuthorID: "1",
	}

	// Save data
	db.Create(&user)
	db.Create(&profile)
	db.Create(&post1)
	db.Create(&post2)

	// Create resource
	res := &DefaultResource{
		Name: "users",
		Relations: []Relation{
			{
				Name:             "Posts",
				Type:             RelationTypeOneToMany,
				Resource:         "posts",
				IncludeByDefault: false,
			},
			{
				Name:             "Profile",
				Type:             RelationTypeOneToOne,
				Resource:         "profiles",
				IncludeByDefault: true,
			},
		},
	}

	// Test loading a single record with relations
	var loadedUser User
	err = LoadRelations(db, res, &loadedUser, []string{"Posts", "Profile"})
	assert.NoError(t, err)

	// Verify relations were loaded
	assert.Equal(t, "John Doe", loadedUser.Name)
	assert.Len(t, loadedUser.Posts, 2)
	assert.NotNil(t, loadedUser.Profile)
	assert.Equal(t, "Test bio", loadedUser.Profile.Bio)

	// Test loading multiple records with relations
	var users []User
	err = LoadRelationsForMany(db, res, &users, []string{"Posts", "Profile"})
	assert.NoError(t, err)

	// Verify relations were loaded
	assert.Len(t, users, 1)
	assert.Equal(t, "John Doe", users[0].Name)
	assert.Len(t, users[0].Posts, 2)
	assert.NotNil(t, users[0].Profile)
	assert.Equal(t, "Test bio", users[0].Profile.Bio)
}

// Helper function to find a relation by name
func findRelationByName(relations []Relation, name string) *Relation {
	for _, relation := range relations {
		if relation.Name == name {
			return &relation
		}
	}
	return nil
}
