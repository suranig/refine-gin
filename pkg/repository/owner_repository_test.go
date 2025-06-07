package repository

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/suranig/refine-gin/pkg/middleware"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Utility function to generate a unique database name
func uniqueDBName() string {
	return fmt.Sprintf("file::memory:test-%d-%d?cache=shared", time.Now().UnixNano(), time.Now().UnixNano()%100)
}

func TestOwnerRepository_UpdateSuccess(t *testing.T) {
	db, _, ownerRes := setupOwnerTest(t)
	repoIface, err := NewOwnerRepository(db, ownerRes)
	require.NoError(t, err)

	// Insert initial records
	createOwnerTestData(t, db)

	// Grab one record owned by owner-a
	var item OwnerTestEntity
	require.NoError(t, db.First(&item, "owner_id = ?", "owner-a").Error)

	t.Run("Struct input updates record when owner matches", func(t *testing.T) {
		ctx := ownerContext("owner-a")
		updated := &OwnerTestEntity{Name: "Updated Struct"}

		res, err := repoIface.Update(ctx, item.ID, updated)
		require.NoError(t, err)

		updatedItem := res.(*OwnerTestEntity)
		assert.Equal(t, "Updated Struct", updatedItem.Name)
		assert.Equal(t, "owner-a", updatedItem.OwnerID)

		var dbItem OwnerTestEntity
		require.NoError(t, db.First(&dbItem, item.ID).Error)
		assert.Equal(t, "Updated Struct", dbItem.Name)
		assert.Equal(t, "owner-a", dbItem.OwnerID)
	})

	t.Run("Map input leaves owner unchanged when blank", func(t *testing.T) {
		ctx := ownerContext("owner-a")
		updates := map[string]interface{}{
			"name":     "Updated Map",
			"owner_id": "",
		}

		res, err := repoIface.Update(ctx, item.ID, updates)
		require.NoError(t, err)

		upd := res.(*OwnerTestEntity)
		assert.Equal(t, "Updated Map", upd.Name)
		assert.Equal(t, "owner-a", upd.OwnerID)

		var dbItem OwnerTestEntity
		require.NoError(t, db.First(&dbItem, item.ID).Error)
		assert.Equal(t, "Updated Map", dbItem.Name)
		assert.Equal(t, "owner-a", dbItem.OwnerID)
	})

	t.Run("Update fails for mismatched owner", func(t *testing.T) {
		_, err := repoIface.Update(ownerContext("owner-b"), item.ID, &OwnerTestEntity{Name: "Fail"})
		assert.Equal(t, ErrOwnerMismatch, err)
	})

	t.Run("Updating non-existent ID returns error", func(t *testing.T) {
		_, err := repoIface.Update(ownerContext("owner-a"), uint(99999), map[string]interface{}{"name": "none"})
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})
}

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

// verifyOwnership should return an error when the context owner does not match
// the record owner. It should also error when no owner value is present in the
// context.
func TestVerifyOwnershipRejectsMismatch(t *testing.T) {
	repo, db := setupOwnerRepo(t, true, nil)
	items := []OwnerTestEntity{{Name: "One", OwnerID: "a"}, {Name: "Two", OwnerID: "b"}}
	require.NoError(t, db.Create(&items).Error)

	// Mismatched owner should return ErrOwnerMismatch
	err := repo.verifyOwnership(ownerContext("a"), items[1].ID)
	assert.Equal(t, ErrOwnerMismatch, err)

	// Missing owner in context should return ErrOwnerIDNotFound
	err = repo.verifyOwnership(context.Background(), items[0].ID)
	assert.Equal(t, ErrOwnerIDNotFound, err)
}

// setOwnership should automatically populate the owner field using the value
// from context. When no owner value is available an error is returned.
func TestSetOwnershipAssignsOwnerID(t *testing.T) {
	repo, _ := setupOwnerRepo(t, true, nil)

	// Owner ID is set when present in context
	item := &OwnerTestEntity{Name: "auto"}
	require.NoError(t, repo.setOwnership(ownerContext("owner-x"), item))
	assert.Equal(t, "owner-x", item.OwnerID)

	// No owner value yields an error
	itemNoCtx := &OwnerTestEntity{Name: "none"}
	err := repo.setOwnership(context.Background(), itemNoCtx)
	assert.Equal(t, ErrOwnerIDNotFound, err)
}

// applyOwnerFilter should inject a where clause filtering by owner when a value
// exists in the context. If no owner value is provided, an error is returned.
func TestApplyOwnerFilterFiltersByOwner(t *testing.T) {
	repo, db := setupOwnerRepo(t, true, nil)
	ctx := ownerContext("abc")

	tx, err := repo.applyOwnerFilter(ctx, db.Session(&gorm.Session{DryRun: true}))
	require.NoError(t, err)
	tx.Find(&[]OwnerTestEntity{})
	assert.Contains(t, tx.Statement.SQL.String(), "owner_id")
	assert.Equal(t, "abc", tx.Statement.Vars[len(tx.Statement.Vars)-1])

	// Missing owner should produce ErrOwnerIDNotFound
	_, err = repo.applyOwnerFilter(context.Background(), db)
	assert.Equal(t, ErrOwnerIDNotFound, err)
}

func TestOwnerRepository_OwnerSpecificOperations(t *testing.T) {
	db, _, ownerRes := setupOwnerTest(t)
	repoIface, err := NewOwnerRepository(db, ownerRes)
	require.NoError(t, err)

	// Insert test records
	createOwnerTestData(t, db)

	// Fetch example records for owner-a and owner-b
	var itemA OwnerTestEntity
	require.NoError(t, db.First(&itemA, "owner_id = ?", "owner-a").Error)

	t.Run("Get returns a record only when the owner matches", func(t *testing.T) {
		// Correct owner can retrieve the record
		res, err := repoIface.Get(ownerContext("owner-a"), itemA.ID)
		require.NoError(t, err)
		assert.Equal(t, itemA.ID, res.(*OwnerTestEntity).ID)

		// Other owners should receive ErrOwnerMismatch
		_, err = repoIface.Get(ownerContext("owner-b"), itemA.ID)
		assert.Equal(t, ErrOwnerMismatch, err)
	})

	t.Run("Update fails with ErrOwnerMismatch for a mismatched owner", func(t *testing.T) {
		_, err := repoIface.Update(ownerContext("owner-b"), itemA.ID, map[string]interface{}{"name": "changed"})
		assert.Equal(t, ErrOwnerMismatch, err)
	})

	t.Run("Delete respects ownership", func(t *testing.T) {
		// Attempt delete with wrong owner
		err := repoIface.Delete(ownerContext("owner-b"), itemA.ID)
		assert.Equal(t, ErrOwnerMismatch, err)

		// Ensure record still exists
		var exists bool
		require.NoError(t, db.Model(&OwnerTestEntity{}).Where("id = ?", itemA.ID).Select("count(*) > 0").Find(&exists).Error)
		assert.True(t, exists)

		// Deleting with correct owner should succeed
		require.NoError(t, repoIface.Delete(ownerContext("owner-a"), itemA.ID))

		// Verify removal
		err = db.First(&OwnerTestEntity{}, itemA.ID).Error
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	t.Run("Count applies the owner filter correctly", func(t *testing.T) {
		countA, err := repoIface.Count(ownerContext("owner-a"), query.QueryOptions{})
		require.NoError(t, err)
		assert.Equal(t, int64(1), countA)

		countB, err := repoIface.Count(ownerContext("owner-b"), query.QueryOptions{})
		require.NoError(t, err)
		assert.Equal(t, int64(2), countB)
	})
}

// TestOwnerRepository_CreateMany tests the CreateMany method
func TestOwnerRepository_CreateMany(t *testing.T) {
	// Create a fresh database for this test
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&OwnerTestEntity{}))

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

	// Test CreateMany with owner from context
	t.Run("CreateMany sets owner ID from context", func(t *testing.T) {
		ctx := ownerContext("batch-owner")
		items := []OwnerTestEntity{
			{Name: "Batch Item 1"},
			{Name: "Batch Item 2"},
		}

		result, err := repo.CreateMany(ctx, &items)
		require.NoError(t, err)

		// Verify result
		resultItems := result.(*[]OwnerTestEntity)
		assert.Equal(t, 2, len(*resultItems))

		for _, item := range *resultItems {
			assert.Equal(t, "batch-owner", item.OwnerID)
		}

		// Verify in database
		var count int64
		db.Model(&OwnerTestEntity{}).Where("owner_id = ?", "batch-owner").Count(&count)
		assert.Equal(t, int64(2), count)
	})
}

// TestOwnerRepository_UpdateMany tests the UpdateMany method
func TestOwnerRepository_UpdateMany(t *testing.T) {
	// Create a fresh database for this test
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&OwnerTestEntity{}))

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

	// Create test data first
	testItems := []OwnerTestEntity{
		{Name: "Update Item 1", OwnerID: "update-owner"},
		{Name: "Update Item 2", OwnerID: "update-owner"},
		{Name: "Other Owner Item", OwnerID: "other-owner"},
	}
	result := db.Create(&testItems)
	require.NoError(t, result.Error)

	// Get the IDs
	var ids []interface{}
	for _, item := range testItems[:2] { // Only the first two belong to "update-owner"
		ids = append(ids, item.ID)
	}

	// Test UpdateMany with owner verification
	t.Run("UpdateMany only updates owned items", func(t *testing.T) {
		ctx := ownerContext("update-owner")
		updatedData := map[string]interface{}{
			"name": "Updated Batch Item",
		}

		count, err := repo.UpdateMany(ctx, ids, updatedData)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		// Verify in database
		var items []OwnerTestEntity
		db.Where("owner_id = ?", "update-owner").Find(&items)
		for _, item := range items {
			assert.Equal(t, "Updated Batch Item", item.Name)
		}

		// Check that other owner's data wasn't updated
		var otherItem OwnerTestEntity
		db.Where("owner_id = ?", "other-owner").First(&otherItem)
		assert.Equal(t, "Other Owner Item", otherItem.Name)
	})
}

// TestOwnerRepository_DeleteMany tests the DeleteMany method
func TestOwnerRepository_DeleteMany(t *testing.T) {
	// Create a fresh database for this test with a unique identifier
	db, err := gorm.Open(sqlite.Open(uniqueDBName()), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&OwnerTestEntity{}))

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

	// Create test data first
	testItems := []OwnerTestEntity{
		{Name: "Delete Item 1", OwnerID: "delete-owner"},
		{Name: "Delete Item 2", OwnerID: "delete-owner"},
		{Name: "Other Owner Item", OwnerID: "other-owner"},
	}
	result := db.Create(&testItems)
	require.NoError(t, result.Error)

	// Verify initial count
	var initialCount int64
	db.Model(&OwnerTestEntity{}).Count(&initialCount)
	require.Equal(t, int64(3), initialCount, "Should start with 3 items")

	// Get the IDs
	var ids []interface{}
	for _, item := range testItems[:2] { // Only the first two belong to "delete-owner"
		ids = append(ids, item.ID)
	}

	// Test DeleteMany with owner verification
	t.Run("DeleteMany only deletes owned items", func(t *testing.T) {
		ctx := ownerContext("delete-owner")

		rowsAffected, err := repo.DeleteMany(ctx, ids)
		require.NoError(t, err)
		assert.Equal(t, int64(2), rowsAffected, "Should delete 2 rows")

		// Verify in database - check that owner's items are deleted
		var ownerCount int64
		db.Model(&OwnerTestEntity{}).Where("owner_id = ?", "delete-owner").Count(&ownerCount)
		assert.Equal(t, int64(0), ownerCount, "All owner's items should be deleted")

		// Check that other owner's data wasn't deleted
		var otherCount int64
		db.Model(&OwnerTestEntity{}).Where("owner_id = ?", "other-owner").Count(&otherCount)
		assert.Equal(t, int64(1), otherCount, "Other owner's items should remain")

		// Verify total count
		var finalCount int64
		db.Model(&OwnerTestEntity{}).Count(&finalCount)
		assert.Equal(t, int64(1), finalCount, "Should have 1 item remaining")
	})
}

// TestOwnerRepository_DeleteManyErrors verifies that DeleteMany returns errors
// when records don't belong to the caller or do not exist and that no data is
// removed in those cases.
func TestOwnerRepository_DeleteManyErrors(t *testing.T) {
	// Create a fresh database for this test with a unique identifier
	db, err := gorm.Open(sqlite.Open(uniqueDBName()), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&OwnerTestEntity{}))

	// Create resource and owner resource with enforcement enabled
	res := resource.NewResource(resource.ResourceConfig{
		Name:  "owner-test-entity",
		Model: &OwnerTestEntity{},
	})
	ownerRes := resource.NewOwnerResource(res, resource.DefaultOwnerConfig())

	// Create repository
	repo, err := NewOwnerRepository(db, ownerRes)
	require.NoError(t, err)

	// Insert one record for the owner and one for another owner
	items := []OwnerTestEntity{
		{Name: "Owner Item", OwnerID: "owner-a"},
		{Name: "Other Owner Item", OwnerID: "owner-b"},
	}
	require.NoError(t, db.Create(&items).Error)

	otherID := items[1].ID
	missingID := otherID + 1000 // guaranteed non-existent

	// Ensure initial count is two
	var initialCount int64
	db.Model(&OwnerTestEntity{}).Count(&initialCount)
	require.Equal(t, int64(2), initialCount)

	ctx := ownerContext("owner-a")

	t.Run("error for mismatched owner", func(t *testing.T) {
		affected, err := repo.DeleteMany(ctx, []interface{}{otherID})
		assert.Equal(t, ErrOwnerMismatch, err)
		assert.Equal(t, int64(0), affected)

		// No records should be removed
		var count int64
		db.Model(&OwnerTestEntity{}).Count(&count)
		assert.Equal(t, initialCount, count)
	})

	t.Run("error for missing record", func(t *testing.T) {
		affected, err := repo.DeleteMany(ctx, []interface{}{missingID})
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		assert.Equal(t, int64(0), affected)

		// Ensure still no deletions
		var count int64
		db.Model(&OwnerTestEntity{}).Count(&count)
		assert.Equal(t, initialCount, count)
	})
}

// TestOwnerRepository_WithTransaction tests the WithTransaction method
func TestOwnerRepository_WithTransaction(t *testing.T) {
	// Create a fresh database for this test with a unique identifier
	db, err := gorm.Open(sqlite.Open(uniqueDBName()), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&OwnerTestEntity{}))

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

	// Test WithTransaction success case
	t.Run("WithTransaction commit", func(t *testing.T) {
		err := repo.WithTransaction(func(r Repository) error {
			// Create an entity in the transaction
			ctx := ownerContext("tx-owner")
			entity := &OwnerTestEntity{Name: "Transaction Test"}

			_, err := r.Create(ctx, entity)
			return err
		})

		require.NoError(t, err)

		// Verify entity was created by checking the count
		var count int64
		db.Model(&OwnerTestEntity{}).Where("name = ?", "Transaction Test").Count(&count)
		assert.Greater(t, count, int64(0), "Transaction should have created at least one entity")
	})

	// Test WithTransaction rollback case
	t.Run("WithTransaction rollback", func(t *testing.T) {
		// Count before
		var countBefore int64
		db.Model(&OwnerTestEntity{}).Count(&countBefore)

		err := repo.WithTransaction(func(r Repository) error {
			// Create an entity in the transaction
			ctx := ownerContext("tx-owner-rollback")
			entity := &OwnerTestEntity{Name: "Transaction Rollback Test"}

			_, err := r.Create(ctx, entity)
			if err != nil {
				return err
			}

			// Return an error to trigger rollback
			return fmt.Errorf("test error to trigger rollback")
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "test error to trigger rollback")

		// Verify entity was NOT created (count should be the same)
		var countAfter int64
		db.Model(&OwnerTestEntity{}).Count(&countAfter)
		assert.Equal(t, countBefore, countAfter, "Count should remain the same after rollback")
	})
}

// TestOwnerRepository_WithRelations tests the WithRelations method
func TestOwnerRepository_WithRelations(t *testing.T) {
	// Create a fresh database for this test
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&OwnerTestEntity{}))

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

	// Test WithRelations
	t.Run("WithRelations returns repository with relations", func(t *testing.T) {
		relations := []string{"TestRelation"}
		repoWithRelations := repo.WithRelations(relations...)

		// Verify it's still an owner repository
		ownerRepo, ok := repoWithRelations.(*OwnerGenericRepository)
		require.True(t, ok, "Should return an OwnerGenericRepository")

		// Verify the resource is maintained
		assert.Equal(t, ownerRes, ownerRepo.Resource)
	})
}

// MockOwnerRepository is a mock implementation of OwnerRepository for testing
type MockOwnerRepository struct {
	mock.Mock
	Resource resource.OwnerResource
}

func (m *MockOwnerRepository) List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error) {
	args := m.Called(ctx, options)
	return args.Get(0), args.Get(1).(int64), args.Error(2)
}

func (m *MockOwnerRepository) Get(ctx context.Context, id interface{}) (interface{}, error) {
	args := m.Called(ctx, id)
	return args.Get(0), args.Error(1)
}

func (m *MockOwnerRepository) Create(ctx context.Context, data interface{}) (interface{}, error) {
	args := m.Called(ctx, data)
	return args.Get(0), args.Error(1)
}

func (m *MockOwnerRepository) Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error) {
	args := m.Called(ctx, id, data)
	return args.Get(0), args.Error(1)
}

func (m *MockOwnerRepository) Delete(ctx context.Context, id interface{}) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOwnerRepository) Count(ctx context.Context, options query.QueryOptions) (int64, error) {
	args := m.Called(ctx, options)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOwnerRepository) CreateMany(ctx context.Context, data interface{}) (interface{}, error) {
	args := m.Called(ctx, data)
	return args.Get(0), args.Error(1)
}

func (m *MockOwnerRepository) UpdateMany(ctx context.Context, ids []interface{}, data interface{}) (int64, error) {
	args := m.Called(ctx, ids, data)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOwnerRepository) DeleteMany(ctx context.Context, ids []interface{}) (int64, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOwnerRepository) WithTransaction(fn func(Repository) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockOwnerRepository) WithRelations(relations ...string) Repository {
	args := m.Called(relations)
	return args.Get(0).(Repository)
}

func (m *MockOwnerRepository) FindOneBy(ctx context.Context, condition map[string]interface{}) (interface{}, error) {
	args := m.Called(ctx, condition)
	return args.Get(0), args.Error(1)
}

func (m *MockOwnerRepository) FindAllBy(ctx context.Context, condition map[string]interface{}) (interface{}, error) {
	args := m.Called(ctx, condition)
	return args.Get(0), args.Error(1)
}

func (m *MockOwnerRepository) BulkCreate(ctx context.Context, data interface{}) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockOwnerRepository) BulkUpdate(ctx context.Context, condition map[string]interface{}, updates map[string]interface{}) error {
	args := m.Called(ctx, condition, updates)
	return args.Error(0)
}

func (m *MockOwnerRepository) Query(ctx context.Context) *gorm.DB {
	args := m.Called(ctx)
	return args.Get(0).(*gorm.DB)
}

func (m *MockOwnerRepository) GetWithRelations(ctx context.Context, id interface{}, relations []string) (interface{}, error) {
	args := m.Called(ctx, id, relations)
	return args.Get(0), args.Error(1)
}

func (m *MockOwnerRepository) ListWithRelations(ctx context.Context, options query.QueryOptions, relations []string) (interface{}, int64, error) {
	args := m.Called(ctx, options, relations)
	return args.Get(0), args.Get(1).(int64), args.Error(2)
}

func (m *MockOwnerRepository) GetIDFieldName() string {
	args := m.Called()
	return args.String(0)
}

// TestOwnerRepository_FinderWithMock tests the finder methods using direct tests instead of repository integration
func TestOwnerRepository_FinderWithMock(t *testing.T) {
	// Create a fresh database for this test with a unique identifier
	db, err := gorm.Open(sqlite.Open(uniqueDBName()), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&OwnerTestEntity{}))

	// Create test data
	testItems := []OwnerTestEntity{
		{Name: "Find Item 1", OwnerID: "find-owner"},
		{Name: "Find Item 2", OwnerID: "find-owner"},
		{Name: "Other Owner Item", OwnerID: "other-owner"},
	}
	result := db.Create(&testItems)
	require.NoError(t, result.Error)

	// Verify the initial count
	var count int64
	db.Model(&OwnerTestEntity{}).Count(&count)
	require.Equal(t, int64(3), count, "Should have 3 initial items")

	// Test FindOneBy with direct DB call
	t.Run("FindOneBy with direct DB call", func(t *testing.T) {
		var entity OwnerTestEntity
		err := db.Where("name = ? AND owner_id = ?", "Find Item 1", "find-owner").First(&entity).Error
		require.NoError(t, err)

		assert.Equal(t, "Find Item 1", entity.Name)
		assert.Equal(t, "find-owner", entity.OwnerID)
	})

	// Test FindAllBy with direct DB call
	t.Run("FindAllBy with direct DB call", func(t *testing.T) {
		var items []OwnerTestEntity
		err := db.Where("name LIKE ? AND owner_id = ?", "Find Item%", "find-owner").Find(&items).Error
		require.NoError(t, err)

		assert.Equal(t, 2, len(items))
		for _, item := range items {
			assert.Equal(t, "find-owner", item.OwnerID)
		}
	})
}

// TestOwnerRepository_BulkMethodsWithDirect tests bulk methods with direct DB access
func TestOwnerRepository_BulkMethodsWithDirect(t *testing.T) {
	// Create a fresh database for this test with a unique identifier
	db, err := gorm.Open(sqlite.Open(uniqueDBName()), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&OwnerTestEntity{}))

	// Create test data
	testItems := []OwnerTestEntity{
		{Name: "Bulk Item 1", OwnerID: "bulk-owner"},
		{Name: "Bulk Item 2", OwnerID: "bulk-owner"},
		{Name: "Other Owner Item", OwnerID: "other-owner"},
	}
	result := db.Create(&testItems)
	require.NoError(t, result.Error)

	// Verify the initial count
	var count int64
	db.Model(&OwnerTestEntity{}).Count(&count)
	require.Equal(t, int64(3), count, "Should have 3 initial items")

	// Test BulkUpdate with direct DB update
	t.Run("BulkUpdate with direct DB update", func(t *testing.T) {
		// Update directly with the DB
		err := db.Model(&OwnerTestEntity{}).Where("owner_id = ?", "bulk-owner").
			Updates(map[string]interface{}{"name": "Updated Bulk Item"}).Error
		require.NoError(t, err)

		// Verify in database
		var dbItems []OwnerTestEntity
		db.Where("owner_id = ?", "bulk-owner").Find(&dbItems)

		assert.Equal(t, 2, len(dbItems))
		for _, item := range dbItems {
			assert.Equal(t, "Updated Bulk Item", item.Name)
		}
	})
}

// TestOwnerRepository_FindOneBy tests the FindOneBy method
func TestOwnerRepository_FindOneBy(t *testing.T) {
	// Create a fresh database for this test
	db, err := gorm.Open(sqlite.Open(uniqueDBName()), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&OwnerTestEntity{}))

	// Create resource
	res := resource.NewResource(resource.ResourceConfig{
		Name:  "owner-test-entity",
		Model: &OwnerTestEntity{},
	})

	// Create owner resource with default config
	ownerConfig := resource.DefaultOwnerConfig()
	// Ensure we're using the correct DB column name
	ownerConfig.OwnerField = "OwnerID" // Field name in struct
	ownerRes := resource.NewOwnerResource(res, ownerConfig)

	// Create repository
	repo, err := NewOwnerRepository(db, ownerRes)
	require.NoError(t, err)

	// Create test data
	testItems := []OwnerTestEntity{
		{Name: "FindOneBy Item 1", OwnerID: "findone-owner"},
		{Name: "FindOneBy Item 2", OwnerID: "findone-owner"},
		{Name: "FindOneBy Other Owner", OwnerID: "other-owner"},
	}
	result := db.Create(&testItems)
	require.NoError(t, result.Error)

	// Test FindOneBy with owner context
	t.Run("FindOneBy with owner context", func(t *testing.T) {
		ctx := ownerContext("findone-owner")

		// Find by name using direct GORM query to verify setup
		var directResult OwnerTestEntity
		err := db.Where("name = ? AND owner_id = ?", "FindOneBy Item 1", "findone-owner").First(&directResult).Error
		require.NoError(t, err, "Direct query should succeed")

		// Find by name using FindOneBy
		// Use DB column name "owner_id", not struct field name "OwnerID"
		condition := map[string]interface{}{
			"name": "FindOneBy Item 1",
		}

		entity, err := repo.FindOneBy(ctx, condition)
		require.NoError(t, err)

		// Verify result
		item := entity.(*OwnerTestEntity)
		assert.Equal(t, "FindOneBy Item 1", item.Name)
		assert.Equal(t, "findone-owner", item.OwnerID)
	})

	// Test FindOneBy with condition that doesn't match owner
	t.Run("FindOneBy with non-matching owner", func(t *testing.T) {
		ctx := ownerContext("findone-owner")

		// Try to find item from other owner
		condition := map[string]interface{}{
			"name": "FindOneBy Other Owner",
		}

		_, err := repo.FindOneBy(ctx, condition)
		// Should return error since this item belongs to a different owner
		assert.Error(t, err)
	})
}

// TestOwnerRepository_FindAllBy tests the FindAllBy method
func TestOwnerRepository_FindAllBy(t *testing.T) {
	// Create a fresh database for this test
	db, err := gorm.Open(sqlite.Open(uniqueDBName()), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&OwnerTestEntity{}))

	// Create resource
	res := resource.NewResource(resource.ResourceConfig{
		Name:  "owner-test-entity",
		Model: &OwnerTestEntity{},
	})

	// Create owner resource with default config
	ownerConfig := resource.DefaultOwnerConfig()
	// Ensure we're using the correct DB column name
	ownerConfig.OwnerField = "OwnerID" // Field name in struct
	ownerRes := resource.NewOwnerResource(res, ownerConfig)

	// Create repository
	repo, err := NewOwnerRepository(db, ownerRes)
	require.NoError(t, err)

	// Create test data
	testItems := []OwnerTestEntity{
		{Name: "FindAllBy Item 1", OwnerID: "findall-owner"},
		{Name: "FindAllBy Item 2", OwnerID: "findall-owner"},
		{Name: "FindAllBy Item 3", OwnerID: "findall-owner"},
		{Name: "FindAllBy Other Owner", OwnerID: "other-owner"},
	}
	result := db.Create(&testItems)
	require.NoError(t, result.Error)

	// Test FindAllBy with owner context
	t.Run("FindAllBy with simple condition", func(t *testing.T) {
		ctx := ownerContext("findall-owner")

		// Verify setup with direct query
		var directResults []OwnerTestEntity
		err := db.Where("owner_id = ?", "findall-owner").Find(&directResults).Error
		require.NoError(t, err, "Direct query should succeed")
		assert.Equal(t, 3, len(directResults), "Direct query should find 3 items")

		// Use simple condition with exact match
		condition := map[string]interface{}{
			"name": "FindAllBy Item 1",
		}

		results, err := repo.FindAllBy(ctx, condition)
		require.NoError(t, err)

		// Verify results
		items := results.(*[]OwnerTestEntity)
		assert.Equal(t, 1, len(*items))
		assert.Equal(t, "FindAllBy Item 1", (*items)[0].Name)
		assert.Equal(t, "findall-owner", (*items)[0].OwnerID)
	})

	// Test FindAllBy with empty condition to get all owner's items
	t.Run("FindAllBy with empty condition", func(t *testing.T) {
		ctx := ownerContext("findall-owner")

		// Use empty condition
		condition := map[string]interface{}{}

		results, err := repo.FindAllBy(ctx, condition)
		require.NoError(t, err)

		// Verify results
		items := results.(*[]OwnerTestEntity)
		assert.Equal(t, 3, len(*items))

		// All items should belong to the owner
		for _, item := range *items {
			assert.Equal(t, "findall-owner", item.OwnerID)
		}
	})

	// Test FindAllBy with disabled ownership
	t.Run("FindAllBy with disabled ownership", func(t *testing.T) {
		// Create non-enforcing repo
		nonEnforcingRes := resource.NewOwnerResource(res, resource.OwnerConfig{
			OwnerField:       "OwnerID",
			EnforceOwnership: false,
		})
		nonEnforcingRepo, err := NewOwnerRepository(db, nonEnforcingRes)
		require.NoError(t, err)

		ctx := ownerContext("findall-owner")

		// Find all items
		condition := map[string]interface{}{}

		results, err := nonEnforcingRepo.FindAllBy(ctx, condition)
		require.NoError(t, err)

		// Should return all items regardless of owner
		items := results.(*[]OwnerTestEntity)
		assert.GreaterOrEqual(t, len(*items), 4)
	})
}

// TestOwnerRepository_QueryMethod tests the Query method
func TestOwnerRepository_QueryMethod(t *testing.T) {
	// Create a fresh database for this test
	db, err := gorm.Open(sqlite.Open(uniqueDBName()), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&OwnerTestEntity{}))

	// Create resource
	res := resource.NewResource(resource.ResourceConfig{
		Name:  "owner-test-entity",
		Model: &OwnerTestEntity{},
	})

	// Create owner resource with default config
	ownerConfig := resource.DefaultOwnerConfig()
	// Ensure we're using the correct DB column name
	ownerConfig.OwnerField = "OwnerID" // Field name in struct
	ownerRes := resource.NewOwnerResource(res, ownerConfig)

	// Create repository
	repo, err := NewOwnerRepository(db, ownerRes)
	require.NoError(t, err)

	// Create test data
	testItems := []OwnerTestEntity{
		{Name: "Query Item 1", OwnerID: "query-owner"},
		{Name: "Query Item 2", OwnerID: "query-owner"},
		{Name: "Query Other Owner", OwnerID: "other-owner"},
	}
	result := db.Create(&testItems)
	require.NoError(t, result.Error)

	// Test Query method with owner context
	t.Run("Query with owner context", func(t *testing.T) {
		ctx := ownerContext("query-owner")

		// Verify our test data using direct queries
		var directResults []OwnerTestEntity
		err := db.Where("name LIKE ? AND owner_id = ?", "Query Item%", "query-owner").Find(&directResults).Error
		require.NoError(t, err, "Direct query should succeed")
		assert.Equal(t, 2, len(directResults), "Direct query should find 2 items")

		// Get query DB instance
		queryDB := repo.Query(ctx)
		require.NotNil(t, queryDB)

		// Execute query
		var items []OwnerTestEntity
		err = queryDB.Where("name LIKE ?", "Query Item%").Find(&items).Error
		require.NoError(t, err)

		// Verify results are filtered by owner
		assert.Equal(t, 2, len(items))
		for _, item := range items {
			assert.Equal(t, "query-owner", item.OwnerID)
		}

		// Verify that the query has already applied the owner filter
		// by checking that we only get items for the current owner
		var allItems []OwnerTestEntity
		err = queryDB.Find(&allItems).Error
		require.NoError(t, err)
		assert.Equal(t, 2, len(allItems), "Should only find items for current owner")
		for _, item := range allItems {
			assert.Equal(t, "query-owner", item.OwnerID, "All items should be for current owner")
		}
	})
}

// TestOwnerRepository_WithRelationsMethods tests the GetWithRelations and ListWithRelations methods
func TestOwnerRepository_WithRelationsMethods(t *testing.T) {
	// Create a fresh database for this test
	db, err := gorm.Open(sqlite.Open(uniqueDBName()), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&OwnerTestEntity{}))

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

	// Create test data
	testItems := []OwnerTestEntity{
		{Name: "Relations Item 1", OwnerID: "relations-owner"},
		{Name: "Relations Item 2", OwnerID: "relations-owner"},
		{Name: "Relations Other Owner", OwnerID: "other-owner"},
	}
	result := db.Create(&testItems)
	require.NoError(t, result.Error)

	// Get the first item ID for later use
	var firstItem OwnerTestEntity
	err = db.Where("owner_id = ?", "relations-owner").First(&firstItem).Error
	require.NoError(t, err)

	// Test GetWithRelations
	t.Run("GetWithRelations with owner context", func(t *testing.T) {
		ctx := ownerContext("relations-owner")

		// Get item with relations (even though there are no real relations)
		item, err := repo.GetWithRelations(ctx, firstItem.ID, []string{})
		require.NoError(t, err)

		// Verify item
		retrievedItem := item.(*OwnerTestEntity)
		assert.Equal(t, firstItem.ID, retrievedItem.ID)
		assert.Equal(t, "relations-owner", retrievedItem.OwnerID)
	})

	// Test GetWithRelations for non-owned item
	t.Run("GetWithRelations for non-owned item", func(t *testing.T) {
		ctx := ownerContext("relations-owner")

		// Try to get an item owned by someone else
		var otherItem OwnerTestEntity
		err = db.Where("owner_id = ?", "other-owner").First(&otherItem).Error
		require.NoError(t, err)

		// Should fail
		_, err := repo.GetWithRelations(ctx, otherItem.ID, []string{})
		assert.Error(t, err)
	})

	// Test ListWithRelations
	t.Run("ListWithRelations with owner context", func(t *testing.T) {
		ctx := ownerContext("relations-owner")

		// List items with relations
		options := query.QueryOptions{
			Page:              1,
			PerPage:           10,
			DisablePagination: false,
		}

		results, count, err := repo.ListWithRelations(ctx, options, []string{})
		require.NoError(t, err)

		// Verify results
		items := results.(*[]OwnerTestEntity)
		assert.Equal(t, int64(2), count)
		assert.Equal(t, 2, len(*items))

		// All items should belong to the owner
		for _, item := range *items {
			assert.Equal(t, "relations-owner", item.OwnerID)
		}
	})
}

// TestOwnerRepository_GetIDFieldName tests the GetIDFieldName method
func TestOwnerRepository_GetIDFieldName(t *testing.T) {
	// Create a fresh database for this test
	db, err := gorm.Open(sqlite.Open(uniqueDBName()), &gorm.Config{})
	require.NoError(t, err)

	// Create resource
	res := resource.NewResource(resource.ResourceConfig{
		Name:        "owner-test-entity",
		Model:       &OwnerTestEntity{},
		IDFieldName: "ID", // Default ID field
	})

	// Create owner resource with default config
	ownerRes := resource.NewOwnerResource(res, resource.DefaultOwnerConfig())

	// Create repository
	repo, err := NewOwnerRepository(db, ownerRes)
	require.NoError(t, err)

	// Test GetIDFieldName
	t.Run("GetIDFieldName returns correct field name", func(t *testing.T) {
		fieldName := repo.GetIDFieldName()
		assert.Equal(t, "ID", fieldName)
	})

	// Test with custom ID field
	t.Run("GetIDFieldName with custom field", func(t *testing.T) {
		customRes := resource.NewResource(resource.ResourceConfig{
			Name:        "custom-id-entity",
			Model:       &OwnerTestEntity{},
			IDFieldName: "CustomID",
		})

		customOwnerRes := resource.NewOwnerResource(customRes, resource.DefaultOwnerConfig())
		customRepo, err := NewOwnerRepository(db, customOwnerRes)
		require.NoError(t, err)

		fieldName := customRepo.GetIDFieldName()
		assert.Equal(t, "CustomID", fieldName)
	})
}

// TestOwnerRepository_BulkCreateAndUpdate tests BulkCreate and BulkUpdate methods
func TestOwnerRepository_BulkCreateAndUpdate(t *testing.T) {
	// Create a fresh database for this test
	db, err := gorm.Open(sqlite.Open(uniqueDBName()), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&OwnerTestEntity{}))

	// Create resource
	res := resource.NewResource(resource.ResourceConfig{
		Name:  "owner-test-entity",
		Model: &OwnerTestEntity{},
	})

	// Create owner resource with default config
	ownerConfig := resource.DefaultOwnerConfig()
	// Ensure we're using the correct DB column name
	ownerConfig.OwnerField = "OwnerID" // Field name in struct
	ownerRes := resource.NewOwnerResource(res, ownerConfig)

	// Create repository
	repo, err := NewOwnerRepository(db, ownerRes)
	require.NoError(t, err)

	// Test BulkCreate with owner context
	t.Run("BulkCreate sets owner ID from context", func(t *testing.T) {
		ctx := ownerContext("bulk-create-owner")

		// Create test data
		testItems := []OwnerTestEntity{
			{Name: "Bulk Create Item 1"},
			{Name: "Bulk Create Item 2"},
			{Name: "Bulk Create Item 3"},
		}

		err := repo.BulkCreate(ctx, testItems)
		require.NoError(t, err)

		// Verify items were created with correct owner ID
		var createdItems []OwnerTestEntity
		err = db.Where("owner_id = ?", "bulk-create-owner").Find(&createdItems).Error
		require.NoError(t, err)

		assert.Equal(t, 3, len(createdItems))
		for _, item := range createdItems {
			assert.Equal(t, "bulk-create-owner", item.OwnerID)
			assert.True(t, strings.HasPrefix(item.Name, "Bulk Create Item"))
		}
	})

	// Test BulkUpdate with owner context
	t.Run("BulkUpdate only updates owned items", func(t *testing.T) {
		// Create test data with different owners
		initialItems := []OwnerTestEntity{
			{Name: "Bulk Update Item 1", OwnerID: "bulk-update-owner"},
			{Name: "Bulk Update Item 2", OwnerID: "bulk-update-owner"},
			{Name: "Bulk Update Item 3", OwnerID: "other-owner"},
		}

		result := db.Create(&initialItems)
		require.NoError(t, result.Error)

		// Get context with owner ID
		ctx := ownerContext("bulk-update-owner")

		// Update all items matching condition - use simple condition
		condition := map[string]interface{}{}

		updates := map[string]interface{}{
			"name": "Bulk Updated",
		}

		err := repo.BulkUpdate(ctx, condition, updates)
		require.NoError(t, err)

		// Verify only owned items were updated
		var updatedItems []OwnerTestEntity
		err = db.Where("name = ?", "Bulk Updated").Find(&updatedItems).Error
		require.NoError(t, err)

		// Should have updated only the 2 items owned by bulk-update-owner
		assert.Equal(t, 2, len(updatedItems))

		// All updated items should belong to the owner
		for _, item := range updatedItems {
			assert.Equal(t, "bulk-update-owner", item.OwnerID)
		}

		// Verify item with other-owner was not updated
		var otherItem OwnerTestEntity
		err = db.Where("owner_id = ?", "other-owner").First(&otherItem).Error
		require.NoError(t, err)

		assert.Equal(t, "Bulk Update Item 3", otherItem.Name)
	})
}
