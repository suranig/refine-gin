package query

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// helper to setup in-memory DB with sample data
func setupInMemoryDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&TestModel{})
	assert.NoError(t, err)

	testData := []TestModel{
		{ID: "1", Name: "John Doe", Email: "john@example.com", Age: 30},
		{ID: "2", Name: "Jane Doe", Email: "jane@example.com", Age: 25},
		{ID: "3", Name: "Bob Smith", Email: "bob@example.com", Age: 40},
		{ID: "4", Name: "Alice Johnson", Email: "alice@example.com", Age: 35},
		{ID: "5", Name: "Charlie Brown", Email: "charlie@example.com", Age: 28},
	}

	for _, m := range testData {
		assert.NoError(t, db.Create(&m).Error)
	}

	return db
}

func TestApplyPaginate(t *testing.T) {
	db := setupInMemoryDB(t)
	var results []TestModel

	query := ApplyPaginate(db.Model(&TestModel{}).Order("id"), PaginateOption{Page: 2, PerPage: 2})
	err := query.Find(&results).Error
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "3", results[0].ID)
	assert.Equal(t, "4", results[1].ID)
}

func TestApplySort(t *testing.T) {
	db := setupInMemoryDB(t)
	var results []TestModel

	query := ApplySort(db.Model(&TestModel{}), SortOption{Field: "age", Order: "desc"})
	err := query.Find(&results).Error
	assert.NoError(t, err)
	assert.Len(t, results, 5)
	for i := 0; i < len(results)-1; i++ {
		assert.GreaterOrEqual(t, results[i].Age, results[i+1].Age)
	}
}

func TestApplyAdvancedFilters(t *testing.T) {
	db := setupInMemoryDB(t)
	res := createTestResource()

	filters := []Filter{
		{Field: "age", Operator: "gte", Value: 30},
		{Field: "name", Operator: "contains", Value: "Doe"},
	}

	var results []TestModel
	query := applyAdvancedFilters(db.Model(&TestModel{}), filters, res)
	err := query.Find(&results).Error
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "John Doe", results[0].Name)
}

func TestParsePaginationResponse(t *testing.T) {
	opts := QueryOptions{Page: 2, PerPage: 10}
	resp := ParsePaginationResponse(opts, 35)

	assert.Equal(t, 2, resp["page"])
	assert.Equal(t, 10, resp["per_page"])
	assert.Equal(t, int64(35), resp["total"])
	assert.Equal(t, 4, resp["last_page"])
}

func TestToResult(t *testing.T) {
	data := []string{"a", "b"}
	meta := map[string]interface{}{"page": 1}

	res := ToResult(data, meta)
	assert.Equal(t, data, res["data"])
	assert.Equal(t, meta, res["meta"])
}
