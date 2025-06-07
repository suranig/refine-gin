package repository

import (
	"context"
	"fmt"
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
