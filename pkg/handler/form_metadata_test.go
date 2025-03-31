package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/suranig/refine-gin/pkg/resource"
)

// FormMockResource is a mock implementation of resource.Resource for testing forms
type FormMockResource struct {
	mock.Mock
}

func (m *FormMockResource) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *FormMockResource) GetLabel() string {
	args := m.Called()
	return args.String(0)
}

func (m *FormMockResource) GetIcon() string {
	args := m.Called()
	return args.String(0)
}

func (m *FormMockResource) GetModel() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *FormMockResource) GetFields() []resource.Field {
	args := m.Called()
	return args.Get(0).([]resource.Field)
}

func (m *FormMockResource) GetOperations() []resource.Operation {
	args := m.Called()
	return args.Get(0).([]resource.Operation)
}

func (m *FormMockResource) HasOperation(op resource.Operation) bool {
	args := m.Called(op)
	return args.Bool(0)
}

func (m *FormMockResource) GetDefaultSort() *resource.Sort {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*resource.Sort)
}

func (m *FormMockResource) GetFilters() []resource.Filter {
	args := m.Called()
	return args.Get(0).([]resource.Filter)
}

func (m *FormMockResource) GetMiddlewares() []interface{} {
	args := m.Called()
	return args.Get(0).([]interface{})
}

func (m *FormMockResource) GetRelations() []resource.Relation {
	args := m.Called()
	return args.Get(0).([]resource.Relation)
}

func (m *FormMockResource) HasRelation(name string) bool {
	args := m.Called(name)
	return args.Bool(0)
}

func (m *FormMockResource) GetRelation(name string) *resource.Relation {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*resource.Relation)
}

func (m *FormMockResource) GetIDFieldName() string {
	args := m.Called()
	return args.String(0)
}

func (m *FormMockResource) GetField(name string) *resource.Field {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*resource.Field)
}

func (m *FormMockResource) GetSearchable() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *FormMockResource) GetFilterableFields() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *FormMockResource) GetSortableFields() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *FormMockResource) GetTableFields() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *FormMockResource) GetFormFields() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *FormMockResource) GetRequiredFields() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *FormMockResource) GetEditableFields() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *FormMockResource) GetPermissions() map[string][]string {
	args := m.Called()
	return args.Get(0).(map[string][]string)
}

func (m *FormMockResource) HasPermission(operation string, role string) bool {
	args := m.Called(operation, role)
	return args.Bool(0)
}

func (m *FormMockResource) GetFormLayout() *resource.FormLayout {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*resource.FormLayout)
}

// FormTestModel is a sample model for testing
type FormTestModel struct {
	ID        uint   `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

// TestGenerateFormMetadataHandler tests the form metadata handler
func TestGenerateFormMetadataHandler(t *testing.T) {
	// Setup gin
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create mock resource
	mockResource := new(FormMockResource)

	// Define model
	model := &FormTestModel{}

	// Define fields
	fields := []resource.Field{
		{Name: "ID", Type: "number"},
		{Name: "FirstName", Type: "string", Validation: &resource.Validation{Required: true}},
		{Name: "LastName", Type: "string"},
		{Name: "Email", Type: "string", Validation: &resource.Validation{Pattern: "^[\\w-\\.]+@([\\w-]+\\.)+[\\w-]{2,4}$"}},
	}

	// Define form layout
	formLayout := &resource.FormLayout{
		Columns: 2,
		Gutter:  16,
		Sections: []*resource.FormSection{
			{
				ID:    "personal",
				Title: "Personal Information",
			},
		},
		FieldLayouts: []*resource.FormFieldLayout{
			{
				Field:     "FirstName",
				SectionID: "personal",
				Column:    0,
				Row:       0,
			},
			{
				Field:     "LastName",
				SectionID: "personal",
				Column:    1,
				Row:       0,
			},
			{
				Field:     "Email",
				SectionID: "personal",
				Column:    0,
				Row:       1,
				ColSpan:   2,
			},
		},
	}

	// Setup mock expectations
	mockResource.On("GetName").Return("user")
	mockResource.On("GetFields").Return(fields)
	mockResource.On("GetModel").Return(model)
	mockResource.On("GetFormLayout").Return(formLayout)

	// Register handler
	router.GET("/form", GenerateFormMetadataHandler(mockResource))

	// Create request
	req, _ := http.NewRequest("GET", "/form", nil)
	resp := httptest.NewRecorder()

	// Serve request
	router.ServeHTTP(resp, req)

	// Assert response
	assert.Equal(t, http.StatusOK, resp.Code)

	// Parse response
	var response FormMetadataResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Assert fields
	assert.Len(t, response.Fields, 4)

	// Assert layout
	assert.NotNil(t, response.Layout)
	assert.Equal(t, 2, response.Layout.Columns)
	assert.Equal(t, 16, response.Layout.Gutter)
	assert.Len(t, response.Layout.Sections, 1)
	assert.Equal(t, "personal", response.Layout.Sections[0].ID)
	assert.Equal(t, "Personal Information", response.Layout.Sections[0].Title)
	assert.Len(t, response.Layout.FieldLayouts, 3)

	// Modify assertion for DefaultValues - it can be nil or empty
	if response.DefaultValues == nil {
		// DefaultValues can be nil, that's fine
	} else {
		// If not nil, it should be empty because our model doesn't have default values
		assert.Empty(t, response.DefaultValues)
	}

	// Verify mock expectations
	mockResource.AssertExpectations(t)
}

// Test extractDefaultValues function
func TestExtractDefaultValues(t *testing.T) {
	// Test case: nil model
	values := extractDefaultValues(nil)
	assert.Nil(t, values)

	// Test case: struct with default values
	type TestStruct struct {
		Name    string `json:"name"`
		Age     int    `json:"age"`
		Private string `json:"-"`
		hidden  string // unexported field
	}

	model := &TestStruct{
		Name:    "John",
		Age:     30,
		Private: "secret",
		hidden:  "hidden",
	}

	values = extractDefaultValues(model)
	assert.NotNil(t, values)
	assert.Equal(t, "John", values["name"])
	assert.Equal(t, 30, values["age"])
	assert.NotContains(t, values, "Private")
	assert.NotContains(t, values, "hidden")
}

// Test extractFieldDependencies function
func TestExtractFieldDependencies(t *testing.T) {
	// Test case: empty fields
	deps := extractFieldDependencies(nil)
	assert.Nil(t, deps)

	// Test case: fields with dependencies
	fields := []resource.Field{
		{
			Name: "country",
			Type: "string",
		},
		{
			Name: "state",
			Type: "string",
			Form: &resource.FormConfig{
				DependentOn: "country",
			},
		},
		{
			Name: "city",
			Type: "string",
			Form: &resource.FormConfig{
				Dependent: &resource.FormDependency{
					Field: "state",
				},
			},
		},
	}

	deps = extractFieldDependencies(fields)
	assert.NotNil(t, deps)
	assert.Len(t, deps, 2)
	assert.Contains(t, deps, "country")
	assert.Contains(t, deps, "state")
	assert.Contains(t, deps["country"], "state")
	assert.Contains(t, deps["state"], "city")
}
