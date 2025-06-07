package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/suranig/refine-gin/pkg/handler"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
)

// MockRepository is a mock implementation of repository.Repository for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, data interface{}) (interface{}, error) {
	args := m.Called(ctx, data)
	return args.Get(0), args.Error(1)
}

func (m *MockRepository) Get(ctx context.Context, id interface{}) (interface{}, error) {
	args := m.Called(ctx, id)
	return args.Get(0), args.Error(1)
}

func (m *MockRepository) List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error) {
	args := m.Called(ctx, options)
	return args.Get(0), int64(args.Int(1)), args.Error(2)
}

func (m *MockRepository) Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error) {
	args := m.Called(ctx, id, data)
	return args.Get(0), args.Error(1)
}

func (m *MockRepository) Delete(ctx context.Context, id interface{}) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) DeleteMany(ctx context.Context, ids []interface{}) (int64, error) {
	args := m.Called(ctx, ids)
	return int64(args.Int(0)), args.Error(1)
}

func (m *MockRepository) Count(ctx context.Context, options query.QueryOptions) (int64, error) {
	args := m.Called(ctx, options)
	return int64(args.Int(0)), args.Error(1)
}

func (m *MockRepository) CreateMany(ctx context.Context, data interface{}) (interface{}, error) {
	args := m.Called(ctx, data)
	return args.Get(0), args.Error(1)
}

func (m *MockRepository) UpdateMany(ctx context.Context, ids []interface{}, data interface{}) (int64, error) {
	args := m.Called(ctx, ids, data)
	return int64(args.Int(0)), args.Error(1)
}

func (m *MockRepository) WithTransaction(fn func(repository.Repository) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockRepository) WithRelations(relations ...string) repository.Repository {
	args := m.Called(relations)
	return args.Get(0).(repository.Repository)
}

func (m *MockRepository) FindOneBy(ctx context.Context, condition map[string]interface{}) (interface{}, error) {
	args := m.Called(ctx, condition)
	return args.Get(0), args.Error(1)
}

func (m *MockRepository) FindAllBy(ctx context.Context, condition map[string]interface{}) (interface{}, error) {
	args := m.Called(ctx, condition)
	return args.Get(0), args.Error(1)
}

func (m *MockRepository) GetWithRelations(ctx context.Context, id interface{}, relations []string) (interface{}, error) {
	args := m.Called(ctx, id, relations)
	return args.Get(0), args.Error(1)
}

func (m *MockRepository) ListWithRelations(ctx context.Context, options query.QueryOptions, relations []string) (interface{}, int64, error) {
	args := m.Called(ctx, options, relations)
	return args.Get(0), int64(args.Int(1)), args.Error(2)
}

func (m *MockRepository) Query(ctx context.Context) *gorm.DB {
	args := m.Called(ctx)
	return args.Get(0).(*gorm.DB)
}

func (m *MockRepository) BulkCreate(ctx context.Context, data interface{}) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockRepository) BulkUpdate(ctx context.Context, condition map[string]interface{}, updates map[string]interface{}) error {
	args := m.Called(ctx, condition, updates)
	return args.Error(0)
}

// GetIDFieldName implements the repository.Repository interface
func (m *MockRepository) GetIDFieldName() string {
	args := m.Called()
	return args.String(0)
}

// MockResource is a mock implementation of resource.Resource for testing
type MockResource struct {
	mock.Mock
}

func (m *MockResource) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockResource) GetLabel() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockResource) GetIcon() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockResource) GetModel() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockResource) GetFields() []resource.Field {
	args := m.Called()
	return args.Get(0).([]resource.Field)
}

func (m *MockResource) GetOperations() []resource.Operation {
	args := m.Called()
	return args.Get(0).([]resource.Operation)
}

func (m *MockResource) HasOperation(op resource.Operation) bool {
	args := m.Called(op)
	return args.Bool(0)
}

func (m *MockResource) GetDefaultSort() *resource.Sort {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*resource.Sort)
}

func (m *MockResource) GetFilters() []resource.Filter {
	args := m.Called()
	return args.Get(0).([]resource.Filter)
}

func (m *MockResource) GetMiddlewares() []interface{} {
	args := m.Called()
	return args.Get(0).([]interface{})
}

func (m *MockResource) GetRelations() []resource.Relation {
	args := m.Called()
	return args.Get(0).([]resource.Relation)
}

func (m *MockResource) HasRelation(name string) bool {
	args := m.Called(name)
	return args.Bool(0)
}

func (m *MockResource) GetRelation(name string) *resource.Relation {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	rel := args.Get(0).(resource.Relation)
	return &rel
}

func (m *MockResource) GetIDFieldName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockResource) GetField(name string) *resource.Field {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	field := args.Get(0).(resource.Field)
	return &field
}

func (m *MockResource) GetSearchable() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockResource) GetFilterableFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *MockResource) GetSortableFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *MockResource) GetRequiredFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *MockResource) GetTableFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *MockResource) GetFormFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *MockResource) GetEditableFields() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockResource) GetPermissions() map[string][]string {
	args := m.Called()
	if result := args.Get(0); result != nil {
		return result.(map[string][]string)
	}
	return nil
}

func (m *MockResource) HasPermission(operation string, role string) bool {
	args := m.Called(operation, role)
	return args.Bool(0)
}

// GetFormLayout returns the form layout configuration
func (r *MockResource) GetFormLayout() *resource.FormLayout {
	args := r.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*resource.FormLayout)
}

func TestRegisterResourceForRefineWithRelations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	router := r.Group("/api")
	mockRepo := new(MockRepository)
	mockResource := new(MockResource)

	// Setup resource
	mockResource.On("GetName").Return("tests")
	mockResource.On("GetIDFieldName").Return("ID")
	mockRelations := []resource.Relation{
		{
			Name:  "comments",
			Type:  handler.HasMany,
			Field: "Comments",
		},
		{
			Name:  "author",
			Type:  handler.BelongsTo,
			Field: "Author",
		},
	}
	mockResource.On("GetRelations").Return(mockRelations)
	mockResource.On("HasRelation", "comments").Return(true)
	mockResource.On("HasRelation", "author").Return(true)

	commentsRelation := mockRelations[0]
	authorRelation := mockRelations[1]
	mockResource.On("GetRelation", "comments").Return(commentsRelation)
	mockResource.On("GetRelation", "author").Return(authorRelation)

	// Add other required mock methods for resource and repository
	mockResource.On("GetLabel").Return("Tests")
	mockResource.On("GetIcon").Return("")
	mockResource.On("GetModel").Return(struct{}{})
	mockResource.On("GetFields").Return([]resource.Field{})
	mockResource.On("GetOperations").Return([]resource.Operation{})
	mockResource.On("GetDefaultSort").Return(nil)
	mockResource.On("GetFilters").Return([]resource.Filter{})
	mockResource.On("GetMiddlewares").Return([]interface{}{})
	mockResource.On("GetSearchable").Return([]string{})

	// Match any operation name with this catch-all mock
	mockResource.On("HasOperation", mock.AnythingOfType("resource.Operation")).Return(true)

	// Add repository method mocks
	mockRepo.On("List", mock.Anything, mock.AnythingOfType("query.QueryOptions")).Return([]interface{}{}, 0, nil)
	mockRepo.On("Get", mock.Anything, mock.Anything).Return(map[string]interface{}{"id": "123"}, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(map[string]interface{}{"id": "123"}, nil)
	mockRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{"id": "123"}, nil)
	mockRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("Count", mock.Anything, mock.AnythingOfType("query.QueryOptions")).Return(int64(0), nil)
	mockRepo.On("FindOneBy", mock.Anything, mock.Anything).Return(map[string]interface{}{"id": "123"}, nil)
	mockRepo.On("GetWithRelations", mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{"id": "123"}, nil)
	mockRepo.On("ListWithRelations", mock.Anything, mock.AnythingOfType("query.QueryOptions"), mock.Anything).Return([]interface{}{}, 0, nil)

	// Register resource with relations
	relationNames := []string{"comments", "author"}
	handler.RegisterResourceForRefineWithRelations(router, mockResource, mockRepo, "id", relationNames)

	// Verify that routes for relations were registered
	expectedRoutes := []struct {
		method string
		path   string
	}{
		// Regular CRUD routes
		{http.MethodGet, "/api/tests"},
		{http.MethodGet, "/api/tests/:id"},
		{http.MethodPost, "/api/tests"},
		{http.MethodPut, "/api/tests/:id"},
		{http.MethodDelete, "/api/tests/:id"},

		// Relation routes for comments
		{http.MethodPost, "/api/tests/:id/actions/attach-comments"},
		{http.MethodPost, "/api/tests/:id/actions/detach-comments"},
		{http.MethodGet, "/api/tests/:id/actions/list-comments"},

		// Relation routes for author
		{http.MethodPost, "/api/tests/:id/actions/attach-author"},
		{http.MethodPost, "/api/tests/:id/actions/detach-author"},
		{http.MethodGet, "/api/tests/:id/actions/list-author"},
	}

	// Validate that all expected routes were registered
	// This is a bit of a hack since Gin doesn't expose routes directly
	for _, route := range expectedRoutes {
		req := httptest.NewRequest(route.method, route.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// We're not testing the actual handler behavior here,
		// just that a route was registered (404 means not registered)
		assert.NotEqual(t, http.StatusNotFound, w.Code,
			fmt.Sprintf("Route %s %s was not registered", route.method, route.path))
	}
}

func TestAttachAndDetachActions(t *testing.T) {
	// Test the attachment action
	t.Run("AttachAction", func(t *testing.T) {
		attachAction := handler.AttachAction("items")
		assert.Equal(t, "attach-items", attachAction.Name)
		assert.Equal(t, http.MethodPost, attachAction.Method)
		assert.True(t, attachAction.RequiresID)
		assert.NotNil(t, attachAction.Handler)
	})

	// Test the detachment action
	t.Run("DetachAction", func(t *testing.T) {
		detachAction := handler.DetachAction("items")
		assert.Equal(t, "detach-items", detachAction.Name)
		assert.Equal(t, http.MethodPost, detachAction.Method)
		assert.True(t, detachAction.RequiresID)
		assert.NotNil(t, detachAction.Handler)
	})

	// Test the list relation action
	t.Run("ListRelationAction", func(t *testing.T) {
		listAction := handler.ListRelationAction("items")
		assert.Equal(t, "list-items", listAction.Name)
		assert.Equal(t, http.MethodGet, listAction.Method)
		assert.True(t, listAction.RequiresID)
		assert.NotNil(t, listAction.Handler)
	})
}

func TestGenerateCustomActionHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	res := new(MockResource)
	repo := new(MockRepository)

	res.On("GetName").Return("tests")
	res.On("GetIDFieldName").Return("ID")

	t.Run("POST sets no cache and returns data", func(t *testing.T) {
		r := gin.New()

		action := handler.CustomAction{
			Name:       "do",
			Method:     http.MethodPost,
			RequiresID: false,
			Handler: func(c *gin.Context, r resource.Resource, repo repository.Repository) (interface{}, error) {
				return map[string]string{"ok": "yes"}, nil
			},
		}

		r.POST("/tests/actions/do", handler.GenerateCustomActionHandler(res, repo, action))

		req := httptest.NewRequest(http.MethodPost, "/tests/actions/do", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "no-store, no-cache, must-revalidate", w.Header().Get("Cache-Control"))

		var resp handler.CustomActionResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		expected := map[string]interface{}{"ok": "yes"}
		assert.Equal(t, expected, resp.Data)
	})

	t.Run("error returns 500 JSON", func(t *testing.T) {
		r := gin.New()

		action := handler.CustomAction{
			Name:       "fail",
			Method:     http.MethodPost,
			RequiresID: false,
			Handler: func(c *gin.Context, r resource.Resource, repo repository.Repository) (interface{}, error) {
				return nil, fmt.Errorf("boom")
			},
		}

		r.POST("/tests/actions/fail", handler.GenerateCustomActionHandler(res, repo, action))

		req := httptest.NewRequest(http.MethodPost, "/tests/actions/fail", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "boom", resp["error"])
	})
}

func TestCustomActionHandlerErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockRepo := new(MockRepository)
	mockResource := new(MockResource)

	// Setup resource
	mockResource.On("GetName").Return("tests")
	mockResource.On("GetIDFieldName").Return("ID")

	// Test relation not found error
	t.Run("Relation not found", func(t *testing.T) {
		// Setup a resource with no relations
		mockResource.On("GetRelations").Return([]resource.Relation{})
		mockResource.On("HasRelation", "nonexistent").Return(false)
		mockResource.On("GetRelation", "nonexistent").Return(nil)

		// Mock repository Get method
		mockRepo.On("Get", mock.Anything, "123").Return(nil, nil)

		// Create an attach action with a non-existent relation
		action := handler.AttachAction("nonexistent")
		r.POST("/tests/:id/actions/attach-nonexistent", handler.GenerateCustomActionHandler(mockResource, mockRepo, action))

		// Make request
		req := httptest.NewRequest(http.MethodPost, "/tests/123/actions/attach-nonexistent",
			strings.NewReader(`{"ids": [1, 2, 3]}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "relation nonexistent not found")
	})

	// Test invalid request body
	t.Run("Invalid request body", func(t *testing.T) {
		// Setup a resource with a relation
		relation := resource.Relation{
			Name:  "valid",
			Type:  handler.HasMany,
			Field: "ValidItems",
		}
		mockResource.On("GetRelation", "valid").Return(relation)
		mockResource.On("HasRelation", "valid").Return(true)

		// Create an attach action
		action := handler.AttachAction("valid")
		r.POST("/tests/:id/actions/attach-valid", handler.GenerateCustomActionHandler(mockResource, mockRepo, action))

		// Make request with invalid JSON
		req := httptest.NewRequest(http.MethodPost, "/tests/123/actions/attach-valid",
			strings.NewReader(`{invalid json`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})

	// Test empty IDs array
	t.Run("Empty IDs array", func(t *testing.T) {
		// Setup a resource with a relation
		relation := resource.Relation{
			Name:  "empty",
			Type:  handler.HasMany,
			Field: "EmptyItems",
		}
		mockResource.On("GetRelation", "empty").Return(&relation)
		mockResource.On("HasRelation", "empty").Return(true)

		// Mock repository Get method
		mockRepo.On("Get", mock.Anything, "123").Return(nil, nil)

		// Create an attach action
		action := handler.AttachAction("empty")
		r.POST("/tests/:id/actions/attach-empty", handler.GenerateCustomActionHandler(mockResource, mockRepo, action))

		// Make request with empty IDs array
		req := httptest.NewRequest(http.MethodPost, "/tests/123/actions/attach-empty",
			strings.NewReader(`{"ids": []}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "no IDs provided")
	})

	// Test repository Get error
	t.Run("Repository Get error", func(t *testing.T) {
		// Setup a resource with a relation
		relation := resource.Relation{
			Name:  "repo_error",
			Type:  handler.HasMany,
			Field: "RepoErrorItems",
		}
		mockResource.On("GetRelation", "repo_error").Return(&relation)
		mockResource.On("HasRelation", "repo_error").Return(true)

		// Mock repository error
		mockRepo.On("Get", mock.Anything, "999").Return(nil, fmt.Errorf("resource not found"))

		// Create an attach action
		action := handler.AttachAction("repo_error")
		r.POST("/tests/:id/actions/attach-repo_error", handler.GenerateCustomActionHandler(mockResource, mockRepo, action))

		// Make request
		req := httptest.NewRequest(http.MethodPost, "/tests/999/actions/attach-repo_error",
			strings.NewReader(`{"ids": [1, 2, 3]}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "resource not found")
	})
}

// --- Additional test helpers and structs for detach tests ---

// Detach models used in detach action tests
type TestComment struct{ ID int }
type TestProfile struct{ ID int }
type TestUser struct {
	ID       int
	Comments []TestComment
	Profile  *TestProfile
}

// RecordingRepository is a simple repository that records Update calls
type RecordingRepository struct {
	data         map[string]*TestUser
	UpdatedData  interface{}
	UpdateCalled bool
}

func NewRecordingRepository(user *TestUser) *RecordingRepository {
	return &RecordingRepository{data: map[string]*TestUser{fmt.Sprintf("%v", user.ID): user}}
}

func (r *RecordingRepository) Create(ctx context.Context, data interface{}) (interface{}, error) {
	return nil, nil
}

func (r *RecordingRepository) Get(ctx context.Context, id interface{}) (interface{}, error) {
	if u, ok := r.data[fmt.Sprintf("%v", id)]; ok {
		return u, nil
	}
	return nil, fmt.Errorf("not found")
}

func (r *RecordingRepository) List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error) {
	return nil, 0, nil
}

func (r *RecordingRepository) Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error) {
	r.UpdateCalled = true
	r.UpdatedData = data
	if u, ok := data.(*TestUser); ok {
		r.data[fmt.Sprintf("%v", id)] = u
	}
	return data, nil
}

func (r *RecordingRepository) Delete(ctx context.Context, id interface{}) error { return nil }
func (r *RecordingRepository) DeleteMany(ctx context.Context, ids []interface{}) (int64, error) {
	return 0, nil
}
func (r *RecordingRepository) Count(ctx context.Context, options query.QueryOptions) (int64, error) {
	return 0, nil
}
func (r *RecordingRepository) CreateMany(ctx context.Context, data interface{}) (interface{}, error) {
	return nil, nil
}
func (r *RecordingRepository) UpdateMany(ctx context.Context, ids []interface{}, data interface{}) (int64, error) {
	return 0, nil
}
func (r *RecordingRepository) WithTransaction(fn func(repository.Repository) error) error {
	return fn(r)
}
func (r *RecordingRepository) WithRelations(relations ...string) repository.Repository { return r }
func (r *RecordingRepository) FindOneBy(ctx context.Context, condition map[string]interface{}) (interface{}, error) {
	return nil, nil
}
func (r *RecordingRepository) FindAllBy(ctx context.Context, condition map[string]interface{}) (interface{}, error) {
	return nil, nil
}
func (r *RecordingRepository) GetWithRelations(ctx context.Context, id interface{}, relations []string) (interface{}, error) {
	return r.Get(ctx, id)
}
func (r *RecordingRepository) ListWithRelations(ctx context.Context, options query.QueryOptions, relations []string) (interface{}, int64, error) {
	return nil, 0, nil
}
func (r *RecordingRepository) Query(ctx context.Context) *gorm.DB                     { return nil }
func (r *RecordingRepository) BulkCreate(ctx context.Context, data interface{}) error { return nil }
func (r *RecordingRepository) BulkUpdate(ctx context.Context, condition map[string]interface{}, updates map[string]interface{}) error {
	return nil
}
func (r *RecordingRepository) GetIDFieldName() string { return "ID" }

func TestDetachActionHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Prepare sample data
	user := &TestUser{
		ID:       1,
		Comments: []TestComment{{ID: 1}, {ID: 2}, {ID: 3}},
		Profile:  &TestProfile{ID: 10},
	}

	repo := NewRecordingRepository(user)

	res := new(MockResource)
	relations := []resource.Relation{
		{Name: "comments", Type: handler.HasMany, Field: "Comments"},
		{Name: "profile", Type: handler.HasOne, Field: "Profile"},
	}
	res.On("GetRelations").Return(relations)
	res.On("HasRelation", "comments").Return(true)
	res.On("HasRelation", "profile").Return(true)
	res.On("GetRelation", "comments").Return(relations[0])
	res.On("GetRelation", "profile").Return(relations[1])

	r := gin.New()
	r.POST("/users/:id/actions/detach-comments", handler.GenerateCustomActionHandler(res, repo, handler.DetachAction("comments")))
	r.POST("/users/:id/actions/detach-profile", handler.GenerateCustomActionHandler(res, repo, handler.DetachAction("profile")))

	t.Run("successful detach", func(t *testing.T) {
		body := strings.NewReader(`{"ids": [2]}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/users/1/actions/detach-comments", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, repo.UpdateCalled)
		assert.Len(t, repo.data["1"].Comments, 2)
		for _, c := range repo.data["1"].Comments {
			assert.NotEqual(t, 2, c.ID)
		}
	})

	t.Run("detach hasone", func(t *testing.T) {
		repo.UpdateCalled = false
		body := strings.NewReader(`{"ids": [10]}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/users/1/actions/detach-profile", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, repo.UpdateCalled)
		assert.Nil(t, repo.data["1"].Profile)
	})

	t.Run("missing ids", func(t *testing.T) {
		repo.UpdateCalled = false
		body := strings.NewReader(`{"ids": []}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/users/1/actions/detach-comments", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.False(t, repo.UpdateCalled)
	})

	t.Run("relation not found", func(t *testing.T) {
		action := handler.DetachAction("ghost")
		r.POST("/users/:id/actions/detach-ghost", handler.GenerateCustomActionHandler(res, repo, action))

		body := strings.NewReader(`{"ids": [1]}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/users/1/actions/detach-ghost", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "relation ghost not found")
		assert.False(t, repo.UpdateCalled)
	})
}
