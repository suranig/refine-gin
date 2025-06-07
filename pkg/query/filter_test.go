package query

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func createFilterTestContext(queryString string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/?"+queryString, nil)
	c.Request = req

	return c, w
}

func TestParseRefineFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := ResourceFilterConfig{
		Fields: []string{"name", "email", "age"},
		Operators: map[string]string{
			"eq":  "=",
			"gt":  ">",
			"gte": ">=",
		},
	}

	// Query with valid refine filter and unsupported field
	qs := "filters[0][field]=name&filters[0][operator]=eq&filters[0][value]=John&" +
		"filters[1][field]=unknown&filters[1][operator]=eq&filters[1][value]=nope&" +
		"filters[2][field]=email&filters[2][value]=john@example.com"
	c, _ := createFilterTestContext(qs)

	filters := ParseRefineFilters(c, config)
	// Should include only the filters for allowed fields
	assert.Len(t, filters, 2)

	f1 := findFilterByField(filters, "name")
	assert.NotNil(t, f1)
	assert.Equal(t, "eq", f1.Operator)
	assert.Equal(t, "John", f1.Value)

	f2 := findFilterByField(filters, "email")
	assert.NotNil(t, f2)
	// Operator defaults to eq when missing
	assert.Equal(t, "eq", f2.Operator)
	assert.Equal(t, "john@example.com", f2.Value)
}

func TestParseRefineFiltersMixedSyntax(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := ResourceFilterConfig{
		Fields: []string{"name", "email", "age"},
		Operators: map[string]string{
			"eq":  "=",
			"gte": ">=",
		},
	}

	qs := "name_eq=John&email=john@example.com&" +
		"filters[0][field]=age&filters[0][operator]=gte&filters[0][value]=30"
	c, _ := createFilterTestContext(qs)

	filters := ParseRefineFilters(c, config)
	assert.Len(t, filters, 3)

	// name_eq=John
	nf := findFilterByField(filters, "name")
	assert.NotNil(t, nf)
	assert.Equal(t, "eq", nf.Operator)
	assert.Equal(t, "John", nf.Value)

	// email=john@example.com
	ef := findFilterByField(filters, "email")
	assert.NotNil(t, ef)
	assert.Equal(t, "eq", ef.Operator)
	assert.Equal(t, "john@example.com", ef.Value)

	// refine style age>=30
	af := findFilterByField(filters, "age")
	assert.NotNil(t, af)
	assert.Equal(t, "gte", af.Operator)
	assert.Equal(t, "30", af.Value)
}

func TestApplyAdvancedFiltersOperators(t *testing.T) {
	db := setupInMemoryDB(t)
	res := createTestResource()

	t.Run("between", func(t *testing.T) {
		filters := []Filter{{Field: "age", Operator: "between", Value: "25,35"}}
		var results []TestModel
		err := applyAdvancedFilters(db.Model(&TestModel{}), filters, res).Find(&results).Error
		assert.NoError(t, err)
		// unsupported operator falls back to equality
		assert.Len(t, results, 0)
	})

	t.Run("null true", func(t *testing.T) {
		filters := []Filter{{Field: "email", Operator: "null", Value: true}}
		var results []TestModel
		err := applyAdvancedFilters(db.Model(&TestModel{}), filters, res).Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})

	t.Run("null false", func(t *testing.T) {
		filters := []Filter{{Field: "email", Operator: "null", Value: false}}
		var results []TestModel
		err := applyAdvancedFilters(db.Model(&TestModel{}), filters, res).Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 5)
	})

	t.Run("containsi", func(t *testing.T) {
		filters := []Filter{{Field: "name", Operator: "containsi", Value: "alice"}}
		var results []TestModel
		err := applyAdvancedFilters(db.Model(&TestModel{}), filters, res).Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Alice Johnson", results[0].Name)
	})

	t.Run("comparison operators", func(t *testing.T) {
		cases := []Filter{
			{Field: "age", Operator: "gt", Value: 30},
			{Field: "age", Operator: "gte", Value: 30},
			{Field: "age", Operator: "lt", Value: 40},
			{Field: "age", Operator: "lte", Value: 40},
			{Field: "name", Operator: "eq", Value: "John Doe"},
			{Field: "name", Operator: "ne", Value: "Non"},
			{Field: "name", Operator: "startswith", Value: "Ja"},
			{Field: "name", Operator: "endswith", Value: "Doe"},
			{Field: "name", Operator: "in", Value: []string{"Bob Smith", "Alice Johnson"}},
		}
		for _, f := range cases {
			var out []TestModel
			assert.NoError(t, applyAdvancedFilters(db.Model(&TestModel{}), []Filter{f}, res).Find(&out).Error)
		}
	})
}
