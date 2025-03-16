package query

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestModel for query tests
type TestModel struct {
	ID    string `json:"id" gorm:"primaryKey" refine:"filterable;sortable;searchable"`
	Name  string `json:"name" refine:"filterable;sortable;searchable"`
	Email string `json:"email" refine:"filterable"`
	Age   int    `json:"age" refine:"filterable"`
}

func TestParseQueryOptions(t *testing.T) {
	// Create a test context
	gin.SetMode(gin.TestMode)

	// Test with no query parameters
	c, _ := createTestContext("")
	res := createTestResource()

	options := ParseQueryOptions(c, res)
	assert.Equal(t, res, options.Resource)
	assert.Equal(t, 1, options.Page)
	assert.Equal(t, 10, options.PerPage)
	assert.Equal(t, "id", options.Sort.Field)
	assert.Equal(t, "asc", options.Sort.Order)
	assert.Empty(t, options.Filters)
	assert.Empty(t, options.Search)

	// Test with pagination parameters
	c, _ = createTestContext("page=2&per_page=20")
	options = ParseQueryOptions(c, res)
	assert.Equal(t, 2, options.Page)
	assert.Equal(t, 20, options.PerPage)

	// Test with sort parameters
	c, _ = createTestContext("sort=name&order=desc")
	options = ParseQueryOptions(c, res)
	assert.Equal(t, "name", options.Sort.Field)
	assert.Equal(t, "desc", options.Sort.Order)

	// Test with filter parameters
	c, _ = createTestContext("name=John&email=example.com&email_operator=contains")
	options = ParseQueryOptions(c, res)
	assert.Len(t, options.Filters, 2)

	// Find name filter
	nameFilter := findFilterByField(options.Filters, "name")
	assert.NotNil(t, nameFilter)
	assert.Equal(t, "name", nameFilter.Field)
	assert.Equal(t, "eq", nameFilter.Operator)
	assert.Equal(t, "John", nameFilter.Value)

	// Find email filter
	emailFilter := findFilterByField(options.Filters, "email")
	assert.NotNil(t, emailFilter)
	assert.Equal(t, "email", emailFilter.Field)
	assert.Equal(t, "contains", emailFilter.Operator)
	assert.Equal(t, "example.com", emailFilter.Value)

	// Test with search parameter
	c, _ = createTestContext("q=searchterm")
	options = ParseQueryOptions(c, res)
	assert.Equal(t, "searchterm", options.Search)
}

// Helper function to create a test context with query parameters
func createTestContext(queryString string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/?"+queryString, nil)
	c.Request = req

	return c, w
}

// Helper function to create a test resource
func createTestResource() resource.Resource {
	return resource.NewResource(resource.ResourceConfig{
		Name:  "tests",
		Model: TestModel{},
		Fields: []resource.Field{
			{Name: "id", Type: "string", Filterable: true, Sortable: true, Searchable: true},
			{Name: "name", Type: "string", Filterable: true, Sortable: true, Searchable: true},
			{Name: "email", Type: "string", Filterable: true},
			{Name: "age", Type: "int", Filterable: true},
		},
		DefaultSort: &resource.Sort{
			Field: "id",
			Order: "asc",
		},
	})
}

func TestApplyQueryOptions(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate model
	err = db.AutoMigrate(&TestModel{})
	assert.NoError(t, err)

	// Create test data
	testData := []TestModel{
		{ID: "1", Name: "John Doe", Email: "john@example.com", Age: 30},
		{ID: "2", Name: "Jane Doe", Email: "jane@example.com", Age: 25},
		{ID: "3", Name: "Bob Smith", Email: "bob@example.com", Age: 40},
		{ID: "4", Name: "Alice Johnson", Email: "alice@example.com", Age: 35},
		{ID: "5", Name: "Charlie Brown", Email: "charlie@example.com", Age: 28},
	}

	for _, model := range testData {
		err := db.Create(&model).Error
		assert.NoError(t, err)
	}

	// Create a test resource
	res := resource.NewResource(resource.ResourceConfig{
		Name:  "tests",
		Model: TestModel{},
		Fields: []resource.Field{
			{Name: "id", Type: "string", Filterable: true, Sortable: true, Searchable: true},
			{Name: "name", Type: "string", Filterable: true, Sortable: true, Searchable: true},
			{Name: "email", Type: "string", Filterable: true},
			{Name: "age", Type: "int", Filterable: true},
		},
	})

	// Test with no options
	options := QueryOptions{
		Resource: res,
		Page:     1,
		PerPage:  10,
		Sort: &resource.Sort{
			Field: "id",
			Order: "asc",
		},
	}

	var results []TestModel
	total, err := options.ApplyWithPagination(db.Model(&TestModel{}), &results)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, results, 5)
	assert.Equal(t, "1", results[0].ID)
	assert.Equal(t, "5", results[4].ID)

	// Test with pagination
	options = QueryOptions{
		Resource: res,
		Page:     2,
		PerPage:  2,
		Sort: &resource.Sort{
			Field: "id",
			Order: "asc",
		},
	}

	results = []TestModel{}
	total, err = options.ApplyWithPagination(db.Model(&TestModel{}), &results)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, results, 2)
	assert.Equal(t, "3", results[0].ID)
	assert.Equal(t, "4", results[1].ID)

	// Test with sorting
	options = QueryOptions{
		Resource: res,
		Page:     1,
		PerPage:  10,
		Sort: &resource.Sort{
			Field: "age",
			Order: "desc",
		},
	}

	results = []TestModel{}
	total, err = options.ApplyWithPagination(db.Model(&TestModel{}), &results)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, results, 5)
	assert.Equal(t, 40, results[0].Age)
	assert.Equal(t, 25, results[4].Age)

	// Test with filtering
	options = QueryOptions{
		Resource: res,
		Page:     1,
		PerPage:  10,
		Sort: &resource.Sort{
			Field: "id",
			Order: "asc",
		},
		Filters: []Filter{
			{Field: "name", Operator: "contains", Value: "Doe"},
		},
	}

	results = []TestModel{}
	total, err = options.ApplyWithPagination(db.Model(&TestModel{}), &results)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)
	assert.Equal(t, "John Doe", results[0].Name)
	assert.Equal(t, "Jane Doe", results[1].Name)

	// Test with search
	options = QueryOptions{
		Resource: res,
		Page:     1,
		PerPage:  10,
		Sort: &resource.Sort{
			Field: "id",
			Order: "asc",
		},
		Search: "Alice",
	}

	results = []TestModel{}
	total, err = options.ApplyWithPagination(db.Model(&TestModel{}), &results)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)
	assert.Equal(t, "Alice Johnson", results[0].Name)
}

// Helper function to find a filter by field
func findFilterByField(filters []Filter, field string) *Filter {
	for _, filter := range filters {
		if filter.Field == field {
			return &filter
		}
	}
	return nil
}

func TestApply(t *testing.T) {
	// Setup in-memory database with a unique name to avoid conflicts
	db, err := gorm.Open(sqlite.Open("file::memory:test_apply?mode=memory&cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate model
	err = db.AutoMigrate(&TestModel{})
	assert.NoError(t, err)

	// Clean up any existing data
	db.Exec("DELETE FROM test_models")

	// Create test data
	testData := []TestModel{
		{ID: "1", Name: "John Doe", Email: "john@example.com", Age: 30},
		{ID: "2", Name: "Jane Doe", Email: "jane@example.com", Age: 25},
		{ID: "3", Name: "Bob Smith", Email: "bob@example.com", Age: 40},
	}

	for _, model := range testData {
		err := db.Create(&model).Error
		assert.NoError(t, err)
	}

	// Create a test resource
	res := createTestResource()

	// Test with various options
	options := QueryOptions{
		Resource: res,
		Page:     2,
		PerPage:  1,
		Sort: &resource.Sort{
			Field: "age",
			Order: "desc",
		},
		Filters: []Filter{
			{Field: "name", Operator: "contains", Value: "Doe"},
		},
	}

	// Apply options to query
	query := db.Model(&TestModel{})
	query = options.Apply(query)

	// Execute query
	var results []TestModel
	err = query.Find(&results).Error
	assert.NoError(t, err)

	// Verify results (should be Jane Doe due to pagination and sorting)
	assert.Len(t, results, 1)
	assert.Equal(t, "Jane Doe", results[0].Name)
}

func TestExtractSort(t *testing.T) {
	// Test with sort and order parameters
	c, _ := createTestContext("sort=name&order=desc")
	sortOption := ExtractSort(c, nil)
	assert.Equal(t, "name", sortOption.Field)
	assert.Equal(t, "desc", sortOption.Order)

	// Test with sort parameter only (default order should be asc)
	c, _ = createTestContext("sort=age")
	sortOption = ExtractSort(c, nil)
	assert.Equal(t, "age", sortOption.Field)
	assert.Equal(t, "asc", sortOption.Order)

	// Test with no parameters but with default sort
	c, _ = createTestContext("")
	defaultSort := &SortOption{Field: "id", Order: "asc"}
	sortOption = ExtractSort(c, defaultSort)
	assert.Equal(t, "id", sortOption.Field)
	assert.Equal(t, "asc", sortOption.Order)

	// Test with no parameters and no default sort
	c, _ = createTestContext("")
	sortOption = ExtractSort(c, nil)
	assert.Equal(t, "", sortOption.Field)
	assert.Equal(t, "asc", sortOption.Order)
}

func TestExtractPaginate(t *testing.T) {
	// Test with page and per_page parameters
	c, _ := createTestContext("page=2&per_page=20")
	paginateOption := ExtractPaginate(c)
	assert.NotNil(t, paginateOption)
	assert.Equal(t, 2, paginateOption.Page)
	assert.Equal(t, 20, paginateOption.PerPage)

	// Test with page parameter only (default per_page should be 10)
	c, _ = createTestContext("page=3")
	paginateOption = ExtractPaginate(c)
	assert.NotNil(t, paginateOption)
	assert.Equal(t, 3, paginateOption.Page)
	assert.Equal(t, 10, paginateOption.PerPage)

	// Test with per_page parameter only (default page should be 1)
	c, _ = createTestContext("per_page=30")
	paginateOption = ExtractPaginate(c)
	assert.NotNil(t, paginateOption)
	assert.Equal(t, 1, paginateOption.Page)
	assert.Equal(t, 30, paginateOption.PerPage)

	// Test with invalid parameters (should use defaults)
	c, _ = createTestContext("page=invalid&per_page=invalid")
	paginateOption = ExtractPaginate(c)
	assert.NotNil(t, paginateOption)
	assert.Equal(t, 1, paginateOption.Page)
	assert.Equal(t, 10, paginateOption.PerPage)

	// Test with no parameters (should use defaults)
	c, _ = createTestContext("")
	paginateOption = ExtractPaginate(c)
	assert.NotNil(t, paginateOption)
	assert.Equal(t, 1, paginateOption.Page)
	assert.Equal(t, 10, paginateOption.PerPage)
}

func TestExtractFilters(t *testing.T) {
	// Create a test context with filter parameters
	c, _ := createTestContext("name=John&email_operator=contains&email=example.com&age_operator=gt&age=25")

	// Define filter configuration
	config := ResourceFilterConfig{
		Fields: []string{"name", "email", "age"},
		Operators: map[string]string{
			"eq":       "=",
			"neq":      "!=",
			"gt":       ">",
			"gte":      ">=",
			"lt":       "<",
			"lte":      "<=",
			"contains": "LIKE",
		},
		DefaultField: "name",
	}

	// Extract filters
	filters := ExtractFilters(c, config)

	// Verify filters
	assert.Len(t, filters, 3)

	// Find and verify each filter
	nameFilter := findFilterOptionByField(filters, "name")
	assert.NotNil(t, nameFilter)
	assert.Equal(t, "name", nameFilter.Field)
	assert.Equal(t, "eq", nameFilter.Operator)
	assert.Equal(t, "John", nameFilter.Value)

	emailFilter := findFilterOptionByField(filters, "email")
	assert.NotNil(t, emailFilter)
	assert.Equal(t, "email", emailFilter.Field)
	assert.Equal(t, "contains", emailFilter.Operator)
	assert.Equal(t, "example.com", emailFilter.Value)

	ageFilter := findFilterOptionByField(filters, "age")
	assert.NotNil(t, ageFilter)
	assert.Equal(t, "age", ageFilter.Field)
	assert.Equal(t, "gt", ageFilter.Operator)
	assert.Equal(t, "25", ageFilter.Value)

	// Test with search parameter
	c, _ = createTestContext("q=searchterm")
	filters = ExtractFilters(c, config)
	assert.Len(t, filters, 1)
	assert.Equal(t, config.DefaultField, filters[0].Field)
	assert.Equal(t, "like", filters[0].Operator)
	assert.Equal(t, "searchterm", filters[0].Value)

	// Test with no parameters
	c, _ = createTestContext("")
	filters = ExtractFilters(c, config)
	assert.Len(t, filters, 0)
}

// Helper function to find a filter option by field
func findFilterOptionByField(filters []FilterOption, field string) *FilterOption {
	for _, filter := range filters {
		if filter.Field == field {
			return &filter
		}
	}
	return nil
}

func TestApplyFilters(t *testing.T) {
	// Setup in-memory database with a unique name to avoid conflicts
	db, err := gorm.Open(sqlite.Open("file::memory:test_filters?mode=memory&cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate model
	err = db.AutoMigrate(&TestModel{})
	assert.NoError(t, err)

	// Clean up any existing data
	db.Exec("DELETE FROM test_models")

	// Create test data
	testData := []TestModel{
		{ID: "1", Name: "John Doe", Email: "john@example.com", Age: 30},
		{ID: "2", Name: "Jane Doe", Email: "jane@example.com", Age: 25},
		{ID: "3", Name: "Bob Smith", Email: "bob@example.com", Age: 40},
		{ID: "4", Name: "Alice Johnson", Email: "alice@example.com", Age: 35},
		{ID: "5", Name: "Charlie Brown", Email: "charlie@example.com", Age: 28},
	}

	for _, model := range testData {
		err := db.Create(&model).Error
		assert.NoError(t, err)
	}

	// Test cases for different filter operators
	testCases := []struct {
		name     string
		filters  []Filter
		expected int
	}{
		{
			name: "Equal operator",
			filters: []Filter{
				{Field: "name", Operator: "eq", Value: "John Doe"},
			},
			expected: 1,
		},
		{
			name: "Not equal operator",
			filters: []Filter{
				{Field: "name", Operator: "neq", Value: "John Doe"},
			},
			expected: 4,
		},
		{
			name: "Greater than operator",
			filters: []Filter{
				{Field: "age", Operator: "gt", Value: 30},
			},
			expected: 2,
		},
		{
			name: "Greater than or equal operator",
			filters: []Filter{
				{Field: "age", Operator: "gte", Value: 30},
			},
			expected: 3,
		},
		{
			name: "Less than operator",
			filters: []Filter{
				{Field: "age", Operator: "lt", Value: 30},
			},
			expected: 2,
		},
		{
			name: "Less than or equal operator",
			filters: []Filter{
				{Field: "age", Operator: "lte", Value: 30},
			},
			expected: 3,
		},
		{
			name: "Like/Contains operator",
			filters: []Filter{
				{Field: "name", Operator: "contains", Value: "Doe"},
			},
			expected: 2,
		},
		{
			name: "Starts with operator",
			filters: []Filter{
				{Field: "name", Operator: "startswith", Value: "John"},
			},
			expected: 1,
		},
		{
			name: "Ends with operator",
			filters: []Filter{
				{Field: "name", Operator: "endswith", Value: "Smith"},
			},
			expected: 1,
		},
		{
			name: "Multiple filters (AND)",
			filters: []Filter{
				{Field: "name", Operator: "contains", Value: "Doe"},
				{Field: "age", Operator: "lt", Value: 30},
			},
			expected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := db.Model(&TestModel{})
			query = ApplyFilters(query, tc.filters)

			var results []TestModel
			err := query.Find(&results).Error
			assert.NoError(t, err)

			assert.Len(t, results, tc.expected, "Expected %d results for %s", tc.expected, tc.name)
		})
	}
}
