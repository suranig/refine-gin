package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/suranig/refine-gin/pkg/resource"
)

// MockResourceForTest is a local mock for testing relation functions
type MockResourceForTest struct {
	mock.Mock
}

func (m *MockResourceForTest) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockResourceForTest) GetLabel() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockResourceForTest) GetIcon() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockResourceForTest) GetModel() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockResourceForTest) GetFields() []resource.Field {
	args := m.Called()
	return args.Get(0).([]resource.Field)
}

func (m *MockResourceForTest) GetOperations() []resource.Operation {
	args := m.Called()
	return args.Get(0).([]resource.Operation)
}

func (m *MockResourceForTest) HasOperation(op resource.Operation) bool {
	args := m.Called(op)
	return args.Bool(0)
}

func (m *MockResourceForTest) GetDefaultSort() *resource.Sort {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*resource.Sort)
}

func (m *MockResourceForTest) GetFilters() []resource.Filter {
	args := m.Called()
	return args.Get(0).([]resource.Filter)
}

func (m *MockResourceForTest) GetMiddlewares() []interface{} {
	args := m.Called()
	return args.Get(0).([]interface{})
}

func (m *MockResourceForTest) GetRelations() []resource.Relation {
	args := m.Called()
	return args.Get(0).([]resource.Relation)
}

func (m *MockResourceForTest) HasRelation(name string) bool {
	args := m.Called(name)
	return args.Bool(0)
}

func (m *MockResourceForTest) GetRelation(name string) *resource.Relation {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	rel := args.Get(0).(resource.Relation)
	return &rel
}

func (m *MockResourceForTest) GetIDFieldName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockResourceForTest) GetField(name string) *resource.Field {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	field := args.Get(0).(resource.Field)
	return &field
}

func (m *MockResourceForTest) GetSearchable() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func TestAttachAndDetachActions(t *testing.T) {
	// Test the attachment action
	t.Run("AttachAction", func(t *testing.T) {
		attachAction := AttachAction("items")
		assert.Equal(t, "attach-items", attachAction.Name)
		assert.Equal(t, http.MethodPost, attachAction.Method)
		assert.True(t, attachAction.RequiresID)
		assert.NotNil(t, attachAction.Handler)
	})

	// Test the detachment action
	t.Run("DetachAction", func(t *testing.T) {
		detachAction := DetachAction("items")
		assert.Equal(t, "detach-items", detachAction.Name)
		assert.Equal(t, http.MethodPost, detachAction.Method)
		assert.True(t, detachAction.RequiresID)
		assert.NotNil(t, detachAction.Handler)
	})

	// Test the list relation action
	t.Run("ListRelationAction", func(t *testing.T) {
		listAction := ListRelationAction("items")
		assert.Equal(t, "list-items", listAction.Name)
		assert.Equal(t, http.MethodGet, listAction.Method)
		assert.True(t, listAction.RequiresID)
		assert.NotNil(t, listAction.Handler)
	})
}

func TestActionOperation(t *testing.T) {
	op := ActionOperation("test-action")
	assert.Equal(t, resource.Operation("custom:test-action"), op)
}

func TestGetRelationByName(t *testing.T) {
	mockResource := new(MockResourceForTest)

	// Test with existing relation
	relation := resource.Relation{
		Name:  "related",
		Type:  HasMany,
		Field: "RelatedItems",
	}
	mockResource.On("GetRelations").Return([]resource.Relation{relation})
	mockResource.On("HasRelation", "related").Return(true)
	mockResource.On("GetRelation", "related").Return(relation)

	result := getRelationByName(mockResource, "related")
	assert.NotNil(t, result)
	assert.Equal(t, "related", result.Name)

	// Test with non-existent relation
	mockResource = new(MockResourceForTest)
	mockResource.On("GetRelations").Return([]resource.Relation{})
	mockResource.On("HasRelation", "nonexistent").Return(false)

	result = getRelationByName(mockResource, "nonexistent")
	assert.Nil(t, result)
}
