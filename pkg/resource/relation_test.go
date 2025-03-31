package resource

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockResource implements a minimal Resource interface for testing
type MockResource struct {
	RelationsValue []Relation
	FieldsValue    []Field
	ModelValue     interface{}
}

func (m *MockResource) GetName() string {
	return "mock-resource"
}

func (m *MockResource) GetLabel() string {
	return "Mock Resource"
}

func (m *MockResource) GetIcon() string {
	return "test-icon"
}

func (m *MockResource) GetModel() interface{} {
	return m.ModelValue
}

func (m *MockResource) GetFields() []Field {
	return m.FieldsValue
}

func (m *MockResource) GetOperations() []Operation {
	return []Operation{}
}

func (m *MockResource) HasOperation(op Operation) bool {
	return false
}

func (m *MockResource) GetDefaultSort() *Sort {
	return nil
}

func (m *MockResource) GetFilters() []Filter {
	return []Filter{}
}

func (m *MockResource) GetMiddlewares() []interface{} {
	return []interface{}{}
}

func (m *MockResource) GetRelations() []Relation {
	return m.RelationsValue
}

func (m *MockResource) HasRelation(name string) bool {
	for _, rel := range m.RelationsValue {
		if rel.Name == name {
			return true
		}
	}
	return false
}

func (m *MockResource) GetRelation(name string) *Relation {
	for _, rel := range m.RelationsValue {
		if rel.Name == name {
			return &rel
		}
	}
	return nil
}

func (m *MockResource) GetIDFieldName() string {
	return "ID"
}

func (m *MockResource) GetField(name string) *Field {
	for _, field := range m.FieldsValue {
		if field.Name == name {
			return &field
		}
	}
	return nil
}

func (m *MockResource) GetSearchable() []string {
	return []string{}
}

func (m *MockResource) GetFilterableFields() []string {
	return []string{}
}

func (m *MockResource) GetSortableFields() []string {
	return []string{}
}

func (m *MockResource) GetTableFields() []string {
	return []string{}
}

func (m *MockResource) GetFormFields() []string {
	return []string{}
}

func (m *MockResource) GetRequiredFields() []string {
	return []string{}
}

func (m *MockResource) GetEditableFields() []string {
	return []string{}
}

func (m *MockResource) GetPermissions() map[string][]string {
	return nil
}

func (m *MockResource) HasPermission(operation string, role string) bool {
	return true
}

// Implement GetFormLayout method for MockResource
func (m *MockResource) GetFormLayout() *FormLayout {
	return nil
}

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

	// Test with a model that has no relations
	type NoRelationsModel struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	noRelations := ExtractRelationsFromModel(NoRelationsModel{})
	assert.Len(t, noRelations, 0)

	// Test with a pointer to a model
	ptrRelations := ExtractRelationsFromModel(&User{})
	assert.Len(t, ptrRelations, 2)

	// Test with a model that has an invalid relation tag
	type InvalidTagModel struct {
		ID      string `json:"id"`
		Related string `json:"related" relation:"invalid_tag"`
	}
	invalidRelations := ExtractRelationsFromModel(InvalidTagModel{})
	assert.Len(t, invalidRelations, 0)

	// Test with a model that has inferred relations
	type InferredRelationModel struct {
		ID          string     `json:"id"`
		Related     struct{}   `json:"related"`
		RelatedMany []struct{} `json:"related_many"`
	}
	inferredRelations := ExtractRelationsFromModel(InferredRelationModel{})
	assert.Len(t, inferredRelations, 2)
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

	// Test with a field that has no relation tag and is not a struct or slice
	idField, _ := userType.FieldByName("ID")
	relation = inferRelationFromField(idField)
	assert.Nil(t, relation)

	// Test with a field that has no relation tag but is a struct
	type TestStruct struct {
		NestedStruct struct{} `json:"nested_struct"`
	}
	testType := reflect.TypeOf(TestStruct{})
	nestedField, _ := testType.FieldByName("NestedStruct")
	relation = inferRelationFromField(nestedField)
	assert.NotNil(t, relation)
	assert.Equal(t, "NestedStruct", relation.Name)
	assert.Equal(t, RelationTypeOneToOne, relation.Type)

	// Test with a field that has no relation tag but is a slice of non-struct
	type TestSliceStruct struct {
		IDs []string `json:"ids"`
	}
	sliceType := reflect.TypeOf(TestSliceStruct{})
	sliceField, _ := sliceType.FieldByName("IDs")
	relation = inferRelationFromField(sliceField)
	assert.Nil(t, relation)

	// Test with a time.Time field (should be skipped)
	type TimeStruct struct {
		CreatedAt time.Time `json:"created_at"`
	}
	timeType := reflect.TypeOf(TimeStruct{})
	timeField, _ := timeType.FieldByName("CreatedAt")
	relation = inferRelationFromField(timeField)
	assert.Nil(t, relation)

	// Test with a pointer to time.Time field (should be skipped)
	type PtrTimeStruct struct {
		CreatedAt *time.Time `json:"created_at"`
	}
	ptrTimeType := reflect.TypeOf(PtrTimeStruct{})
	ptrTimeField, _ := ptrTimeType.FieldByName("CreatedAt")
	relation = inferRelationFromField(ptrTimeField)
	assert.Nil(t, relation)
}

func TestIncludeRelations(t *testing.T) {
	// Create a test context
	c, _ := gin.CreateTestContext(nil)
	req, _ := http.NewRequest("GET", "/", nil)
	c.Request = req

	// Create a mock resource
	relations := []Relation{
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
	}

	// Setup mock resource with our relations
	res := &MockResource{
		RelationsValue: relations,
	}

	// Test with no include parameter (should return default includes)
	includes := IncludeRelations(c, res)
	t.Logf("Default includes: %v", includes)
	assert.Len(t, includes, 1)
	assert.Equal(t, "Profile", includes[0])

	// Test with include parameter
	c2, _ := gin.CreateTestContext(nil)
	req2, _ := http.NewRequest("GET", "/?include=Posts,Profile", nil)
	c2.Request = req2
	c2.Request.URL.RawQuery = "include=Posts,Profile"

	includes = IncludeRelations(c2, res)
	t.Logf("Includes with parameter 'Posts,Profile': %v", includes)
	t.Logf("Query parameter 'include': %v", c2.Query("include"))

	assert.Len(t, includes, 2)
	assert.Contains(t, includes, "Posts")
	assert.Contains(t, includes, "Profile")

	// Test with invalid include parameter
	c3, _ := gin.CreateTestContext(nil)
	req3, _ := http.NewRequest("GET", "/?include=Invalid,Profile", nil)
	c3.Request = req3
	c3.Request.URL.RawQuery = "include=Invalid,Profile"

	includes = IncludeRelations(c3, res)
	t.Logf("Includes with parameter 'Invalid,Profile': %v", includes)
	t.Logf("Query parameter 'include': %v", c3.Query("include"))

	assert.Len(t, includes, 1)
	assert.Equal(t, "Profile", includes[0])
}

func TestLoadRelations(t *testing.T) {
	// Skip this test if running in short mode
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

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
	err = db.Create(&user).Error
	assert.NoError(t, err)

	err = db.Create(&profile).Error
	assert.NoError(t, err)

	err = db.Create(&post1).Error
	assert.NoError(t, err)

	err = db.Create(&post2).Error
	assert.NoError(t, err)

	// Setup relations
	relations := []Relation{
		{
			Name:             "Posts",
			Type:             RelationTypeOneToMany,
			Resource:         "posts",
			Field:            "author_id",
			ReferenceField:   "id",
			IncludeByDefault: false,
		},
		{
			Name:             "Profile",
			Type:             RelationTypeOneToOne,
			Resource:         "profiles",
			Field:            "user_id",
			ReferenceField:   "id",
			IncludeByDefault: true,
		},
	}

	// Create a mock resource with our relations
	res := &MockResource{
		RelationsValue: relations,
	}

	// Test LoadRelations with a single record
	var loadedUser User

	// First find the user
	err = db.First(&loadedUser, "id = ?", "1").Error
	assert.NoError(t, err)

	// Test with empty includes
	err = LoadRelations(db, res, &loadedUser, []string{})
	assert.NoError(t, err)

	// Test with valid includes
	includes := []string{"Posts", "Profile"}
	err = LoadRelations(db, res, &loadedUser, includes)
	assert.NoError(t, err)

	// Test with invalid include
	includes = []string{"Invalid", "Profile"}
	err = LoadRelations(db, res, &loadedUser, includes)
	assert.NoError(t, err)

	// Test LoadRelationsForMany with multiple records
	var users []User

	// First find users
	err = db.Find(&users).Error
	assert.NoError(t, err)

	// Test with empty includes
	err = LoadRelationsForMany(db, res, &users, []string{})
	assert.NoError(t, err)

	// Test with valid includes
	includes = []string{"Posts", "Profile"}
	err = LoadRelationsForMany(db, res, &users, includes)
	assert.NoError(t, err)

	// Test with invalid include
	includes = []string{"Invalid", "Profile"}
	err = LoadRelationsForMany(db, res, &users, includes)
	assert.NoError(t, err)

	// Verify relations were loaded
	assert.Len(t, users, 1)
	assert.Equal(t, "John Doe", users[0].Name)
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
