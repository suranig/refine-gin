package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suranig/refine-gin/pkg/middleware"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Test model with owner field
type OwnerTestEntity struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `json:"name"`
	OwnerID string `json:"ownerId" gorm:"column:owner_id"`
}

// Setup test database and resources
func setupOwnerTest(t *testing.T) (*gorm.DB, resource.Resource, resource.OwnerResource) {
	// Create in-memory SQLite database with unique identifier
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Drop table if exists
	db.Exec("DROP TABLE IF EXISTS owner_test_entities")

	// Create tables
	err = db.AutoMigrate(&OwnerTestEntity{})
	require.NoError(t, err)

	// Verify we start with an empty database
	var count int64
	db.Model(&OwnerTestEntity{}).Count(&count)
	require.Equal(t, int64(0), count, "Should start with an empty database")

	// Create regular resource
	res := resource.NewResource(resource.ResourceConfig{
		Name:  "owner-test-entity",
		Model: &OwnerTestEntity{},
	})

	// Create owner resource
	ownerRes := resource.NewOwnerResource(res, resource.DefaultOwnerConfig())

	return db, res, ownerRes
}

// createOwnerTestData inserts test data into the database
func createOwnerTestData(t *testing.T, db *gorm.DB) {
	// Create test data for different owners
	testData := []OwnerTestEntity{
		{Name: "Item 1 - Owner A", OwnerID: "owner-a"},
		{Name: "Item 2 - Owner A", OwnerID: "owner-a"},
		{Name: "Item 3 - Owner B", OwnerID: "owner-b"},
		{Name: "Item 4 - Owner B", OwnerID: "owner-b"},
		{Name: "Item 5 - Owner C", OwnerID: "owner-c"},
	}

	result := db.Create(&testData)
	require.NoError(t, result.Error)
	require.Equal(t, int64(5), result.RowsAffected)
}

// ownerContext creates a context with owner ID
func ownerContext(ownerID string) context.Context {
	return context.WithValue(context.Background(), middleware.OwnerContextKey, ownerID)
}

// Test NewOwnerRepository function
func TestNewOwnerRepository(t *testing.T) {
	db, res, ownerRes := setupOwnerTest(t)

	// Test with owner resource
	t.Run("With owner resource", func(t *testing.T) {
		repo, err := NewOwnerRepository(db, ownerRes)
		require.NoError(t, err)
		require.NotNil(t, repo)

		// Check that it's an OwnerGenericRepository
		_, ok := repo.(*OwnerGenericRepository)
		require.True(t, ok)
	})

	// Test with regular resource (should be promoted)
	t.Run("With regular resource", func(t *testing.T) {
		repo, err := NewOwnerRepository(db, res)
		require.NoError(t, err)
		require.NotNil(t, repo)

		// Check that it's an OwnerGenericRepository
		_, ok := repo.(*OwnerGenericRepository)
		require.True(t, ok)
	})

	// Test with empty owner field
	t.Run("With empty owner field", func(t *testing.T) {
		emptyOwnerRes := resource.NewOwnerResource(res, resource.OwnerConfig{
			OwnerField:       "",
			EnforceOwnership: true,
		})

		_, err := NewOwnerRepository(db, emptyOwnerRes)
		require.Error(t, err)
		require.Equal(t, ErrNoOwnerFieldName, err)
	})
}

// Test basic ownership filtering functionality
func TestOwnerRepository_Ownership(t *testing.T) {
	// Create a fresh database for this test with a unique identifier
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Drop table if exists
	db.Exec("DROP TABLE IF EXISTS owner_test_entities")

	// Create tables
	require.NoError(t, db.AutoMigrate(&OwnerTestEntity{}))

	// Verify we start with an empty database
	var initialCount int64
	db.Model(&OwnerTestEntity{}).Count(&initialCount)
	require.Equal(t, int64(0), initialCount, "Should start with an empty database")

	// Create resource
	res := resource.NewResource(resource.ResourceConfig{
		Name:  "owner-test-entity",
		Model: &OwnerTestEntity{},
	})

	// Create owner resource with default config
	ownerRes := resource.NewOwnerResource(res, resource.DefaultOwnerConfig())

	// Create repository
	repo, err := NewOwnerRepository(db, ownerRes)
	require.NoError(t, err)

	// Test Create with owner from context
	t.Run("Create sets owner ID from context", func(t *testing.T) {
		ctx := ownerContext("test-owner")
		item := &OwnerTestEntity{Name: "Test Item"}

		result, err := repo.Create(ctx, item)
		require.NoError(t, err)
		assert.Equal(t, "test-owner", result.(*OwnerTestEntity).OwnerID)

		// Verify in database
		var dbItem OwnerTestEntity
		err = db.First(&dbItem, result.(*OwnerTestEntity).ID).Error
		require.NoError(t, err)
		assert.Equal(t, "test-owner", dbItem.OwnerID)
	})

	// Create test data and test basic filtering
	t.Run("Basic ownership filtering", func(t *testing.T) {
		// Create test data for owner A and B
		itemsA := []OwnerTestEntity{
			{Name: "A1", OwnerID: "owner-a"},
			{Name: "A2", OwnerID: "owner-a"},
		}
		itemsB := []OwnerTestEntity{
			{Name: "B1", OwnerID: "owner-b"},
			{Name: "B2", OwnerID: "owner-b"},
		}

		db.Create(&itemsA)
		db.Create(&itemsB)

		// List with owner A context
		ctxA := ownerContext("owner-a")
		options := query.QueryOptions{
			Page:              1,
			PerPage:           100,
			DisablePagination: false,
		}
		resultsA, count, err := repo.List(ctxA, options)
		require.NoError(t, err)

		// Should return only owner A's items
		itemList := resultsA.(*[]OwnerTestEntity)
		assert.Equal(t, int64(2), count)
		for _, item := range *itemList {
			assert.Equal(t, "owner-a", item.OwnerID)
		}

		// Try with ownership disabled
		disabledRepo, err := NewOwnerRepository(db, resource.NewOwnerResource(res, resource.OwnerConfig{
			OwnerField:       "OwnerID",
			EnforceOwnership: false,
		}))
		require.NoError(t, err)

		// Should return all items
		resultsAll, count, err := disabledRepo.List(ctxA, options)
		require.NoError(t, err)
		assert.Greater(t, count, int64(3)) // Should be at least 4 (2 for A + 2 for B)
		assert.GreaterOrEqual(t, len(*resultsAll.(*[]OwnerTestEntity)), 4)
	})
}

// Add a new simple test
func TestOwnerRepository_DisabledEnforcement(t *testing.T) {
	// Create a fresh database for this test
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Drop table if exists
	db.Exec("DROP TABLE IF EXISTS owner_test_entities")

	// Create tables
	require.NoError(t, db.AutoMigrate(&OwnerTestEntity{}))

	// Double check we have no stale data
	var initialCount int64
	db.Model(&OwnerTestEntity{}).Count(&initialCount)
	require.Equal(t, int64(0), initialCount, "Should start with empty database")

	// Enable SQL logging
	db = db.Debug()

	// Create test data
	items := []OwnerTestEntity{
		{Name: "Item 1", OwnerID: "owner-a"},
		{Name: "Item 2", OwnerID: "owner-a"},
		{Name: "Item 3", OwnerID: "owner-b"},
		{Name: "Item 4", OwnerID: "owner-b"},
		{Name: "Item 5", OwnerID: "owner-c"},
	}
	db.Create(&items)

	// Verify data was created
	var count int64
	db.Model(&OwnerTestEntity{}).Count(&count)
	require.Equal(t, int64(5), count, "Should have created 5 items")

	// Test direct query to verify data
	var allItems []OwnerTestEntity
	err = db.Find(&allItems).Error
	require.NoError(t, err)
	t.Logf("Direct DB query found %d items", len(allItems))
	for i, item := range allItems {
		t.Logf("Item %d: ID=%d, Name=%s, OwnerID=%s", i, item.ID, item.Name, item.OwnerID)
	}

	// Create resource with ownership disabled
	res := resource.NewResource(resource.ResourceConfig{
		Name:  "owner-test-entity",
		Model: &OwnerTestEntity{},
	})

	ownerRes := resource.NewOwnerResource(res, resource.OwnerConfig{
		OwnerField:       "OwnerID",
		EnforceOwnership: false, // Disabled!
	})

	// Create repository with ownership disabled
	repo, err := NewOwnerRepository(db, ownerRes)
	require.NoError(t, err)

	// Test if ownership is actually disabled
	t.Logf("Is ownership enforced: %v", ownerRes.IsOwnershipEnforced())

	// List with any context - should return all items
	ctx := ownerContext("owner-a") // Even with an owner ID, should ignore it
	options := query.QueryOptions{
		Page:              1,
		PerPage:           100,
		DisablePagination: false,
	}
	results, total, err := repo.List(ctx, options)
	require.NoError(t, err)

	// Verify we got all items
	resultItems := results.(*[]OwnerTestEntity)
	t.Logf("Got %d items with total count %d", len(*resultItems), total)
	for i, item := range *resultItems {
		t.Logf("Result item %d: ID=%d, Name=%s, OwnerID=%s", i, item.ID, item.Name, item.OwnerID)
	}

	assert.Equal(t, int64(5), total, "Should return all 5 items when ownership is disabled")
	assert.Equal(t, 5, len(*resultItems), "Should return all 5 items when ownership is disabled")
}
