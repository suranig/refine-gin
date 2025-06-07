package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestModel for repository tests
type TestModel struct {
	ID    string `json:"id" gorm:"primaryKey"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// TestModelWithCustomID for custom ID field tests
type TestModelWithCustomID struct {
	UID   string `json:"uid" gorm:"primaryKey"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// RepositoryMockTestSuite is a test suite for the Repository interface implementations with mock DB
type RepositoryMockTestSuite struct {
	suite.Suite
	db         *gorm.DB
	mock       sqlmock.Sqlmock
	repository Repository
	resource   resource.Resource
}

// SetupSuite initializes the test suite with a mock database
func (s *RepositoryMockTestSuite) SetupSuite() {
	// Setup mock database
	var err error
	var mockDB *sql.DB

	mockDB, s.mock, err = sqlmock.New()
	s.Require().NoError(err)

	// Connect GORM to the mock database
	dialector := postgres.New(postgres.Config{
		DSN:                  "sqlmock_db",
		DriverName:           "postgres",
		Conn:                 mockDB,
		PreferSimpleProtocol: true,
	})

	s.db, err = gorm.Open(dialector, &gorm.Config{})
	s.Require().NoError(err)

	// Create resource
	s.resource = resource.NewResource(resource.ResourceConfig{
		Name:  "tests",
		Model: TestModel{},
	})

	// Create repository
	s.repository = NewGenericRepositoryWithResource(s.db, s.resource)
}

// TestMockRepositorySuite runs the mock test suite
func TestMockRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositoryMockTestSuite))
}

// RepositorySuite is a test suite for the Repository interface implementations with SQLite
type RepositorySuite struct {
	suite.Suite
	db         *gorm.DB
	repository Repository
}

// SetupTest initializes a new test database for each test
func (s *RepositorySuite) SetupTest() {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s.Require().NoError(err)

	// Migrate model
	err = db.AutoMigrate(&TestModel{})
	s.Require().NoError(err)

	s.db = db
	s.repository = NewGenericRepositoryWithResource(s.db, resource.NewResource(resource.ResourceConfig{
		Model: &TestModel{},
	}))
}

func (s *RepositorySuite) TestCRUD() {
	ctx := context.Background()

	// Create
	model := &TestModel{
		ID:    "test-1",
		Name:  "Test Model",
		Email: "test@example.com",
		Age:   25,
	}

	result, err := s.repository.Create(ctx, model)
	s.NoError(err)
	s.NotNil(result)

	// Get
	retrieved, err := s.repository.Get(ctx, "test-1")
	s.NoError(err)
	s.NotNil(retrieved)

	retrievedModel := retrieved.(*TestModel)
	s.Equal("test-1", retrievedModel.ID)
	s.Equal("Test Model", retrievedModel.Name)
	s.Equal("test@example.com", retrievedModel.Email)
	s.Equal(25, retrievedModel.Age)

	// Update
	retrievedModel.Name = "Updated Model"
	updated, err := s.repository.Update(ctx, "test-1", retrievedModel)
	s.NoError(err)
	s.NotNil(updated)

	updatedModel := updated.(*TestModel)
	s.Equal("Updated Model", updatedModel.Name)

	// Delete
	err = s.repository.Delete(ctx, map[string]interface{}{"id": retrievedModel.ID})
	s.NoError(err)

	// Verify deletion
	_, err = s.repository.Get(ctx, retrievedModel.ID)
	s.Error(err)
}

// Helper to setup a SQLite test database
func setupSQLiteTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&TestModel{}, &TestModelWithCustomID{})
	assert.NoError(t, err)

	return db
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositorySuite))
}

func TestRepository(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Create test table
	err = db.AutoMigrate(&TestModel{})
	assert.NoError(t, err)

	// Create repository
	res := resource.NewResource(resource.ResourceConfig{
		Model: &TestModel{},
	})
	repo := NewGenericRepositoryWithResource(db, res)

	// Test Create
	model := &TestModel{
		ID:    "test-1",
		Name:  "John Doe",
		Email: "john@example.com",
	}

	result, err := repo.Create(context.Background(), model)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Test Get
	retrieved, err := repo.Get(context.Background(), "test-1")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)

	retrievedModel := retrieved.(*TestModel)
	assert.Equal(t, "test-1", retrievedModel.ID)
	assert.Equal(t, "John Doe", retrievedModel.Name)
	assert.Equal(t, "john@example.com", retrievedModel.Email)

	// Test Update
	retrievedModel.Name = "Jane Doe"
	updated, err := repo.Update(context.Background(), "test-1", retrievedModel)
	assert.NoError(t, err)
	assert.NotNil(t, updated)

	updatedModel := updated.(*TestModel)
	assert.Equal(t, "Jane Doe", updatedModel.Name)

	// Test Delete
	err = repo.Delete(context.Background(), map[string]interface{}{"id": retrievedModel.ID})
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.Get(context.Background(), retrievedModel.ID)
	assert.Error(t, err)
}

func TestRepositoryWithCustomID(t *testing.T) {
	db := setupSQLiteTestDB(t)
	repo := NewGenericRepositoryWithResource(db, resource.NewResource(resource.ResourceConfig{
		Model:       &TestModelWithCustomID{},
		IDFieldName: "UID",
	}))

	ctx := context.Background()

	// Test Create
	model := &TestModelWithCustomID{
		UID:   "custom-1",
		Name:  "John Doe",
		Email: "john@example.com",
	}

	createdModel, err := repo.Create(ctx, model)
	assert.NoError(t, err)
	assert.NotNil(t, createdModel)

	// Test Get
	retrievedModel, err := repo.Get(ctx, model.UID)
	assert.NoError(t, err)
	assert.Equal(t, model.Name, retrievedModel.(*TestModelWithCustomID).Name)

	// Test List
	models, total, err := repo.List(ctx, query.QueryOptions{
		Page:    1,
		PerPage: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)

	modelsList, ok := models.(*[]TestModelWithCustomID)
	assert.True(t, ok)
	assert.Len(t, *modelsList, 1)
	assert.Equal(t, model.UID, (*modelsList)[0].UID)

	// Test Update
	model.Name = "Jane Doe"
	updatedModel, err := repo.Update(ctx, model.UID, model)
	assert.NoError(t, err)
	assert.Equal(t, "Jane Doe", updatedModel.(*TestModelWithCustomID).Name)

	// Test Delete with WHERE condition
	err = repo.Delete(ctx, map[string]interface{}{"uid": model.UID})
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.Get(ctx, model.UID)
	assert.Error(t, err)
}

func TestGenericRepositoryFactory(t *testing.T) {
	db := setupSQLiteTestDB(t)

	factory := NewGenericRepositoryFactory(db)

	res := resource.NewResource(resource.ResourceConfig{
		Name:        "tests",
		Model:       &TestModelWithCustomID{},
		IDFieldName: "UID",
	})

	repo := factory.CreateRepository(res)
	assert.NotNil(t, repo)

	genericRepo, ok := repo.(*GenericRepository)
	assert.True(t, ok)
	assert.Equal(t, res, genericRepo.Resource)
}

// TestCreateMany tests the CreateMany method of the repository
func (s *RepositoryMockTestSuite) TestCreateMany() {
	ctx := context.Background()
	models := []TestModel{
		{ID: "1", Name: "Test 1", Email: "test1@example.com", Age: 20},
		{ID: "2", Name: "Test 2", Email: "test2@example.com", Age: 25},
	}

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`INSERT INTO "test_models"`).
		WithArgs("1", "Test 1", "test1@example.com", 20, "2", "Test 2", "test2@example.com", 25).
		WillReturnResult(sqlmock.NewResult(0, 2))
	s.mock.ExpectCommit()

	result, err := s.repository.CreateMany(ctx, models)
	s.NoError(err)
	s.NotNil(result)

	createdModels, ok := result.([]TestModel)
	s.True(ok)
	s.Len(createdModels, 2)
}

// TestCreateMany_Error tests error handling in CreateMany
func (s *RepositoryMockTestSuite) TestCreateMany_Error() {
	ctx := context.Background()
	models := []TestModel{
		{ID: "1", Name: "Test 1", Email: "", Age: 20},
	}

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`INSERT INTO "test_models"`).
		WithArgs("1", "Test 1", "", 20).
		WillReturnError(errors.New("database error"))
	s.mock.ExpectRollback()

	result, err := s.repository.CreateMany(ctx, models)
	s.Error(err)
	s.Equal("database error", err.Error())
	s.Equal([]TestModel(nil), result)
}

// TestCreateMany_InvalidType tests passing non-slice to CreateMany
func (s *RepositoryMockTestSuite) TestCreateMany_InvalidType() {
	ctx := context.Background()
	invalidModel := "invalid"

	result, err := s.repository.CreateMany(ctx, invalidModel)
	s.Error(err)
	s.Equal("unsupported data type: invalid: Table not set, please set it like: db.Model(&user) or db.Table(\"users\")", err.Error())
	s.Equal("invalid", result)
}

// TestUpdateMany tests the UpdateMany method of the repository
func (s *RepositoryMockTestSuite) TestUpdateMany() {
	ctx := context.Background()
	ids := []interface{}{"1", "2"}
	updates := map[string]interface{}{
		"name": "Updated Name",
	}

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "test_models"`).
		WithArgs("Updated Name", "1", "2").
		WillReturnResult(sqlmock.NewResult(0, 2))
	s.mock.ExpectCommit()

	affected, err := s.repository.UpdateMany(ctx, ids, updates)
	s.NoError(err)
	s.Equal(int64(2), affected)
}

// TestUpdateMany_Error tests error handling in UpdateMany
func (s *RepositoryMockTestSuite) TestUpdateMany_Error() {
	ctx := context.Background()
	ids := []interface{}{"1", "2"}
	updates := map[string]interface{}{
		"name": "Updated Name",
	}

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "test_models"`).
		WithArgs("Updated Name", "1", "2").
		WillReturnError(errors.New("database error"))
	s.mock.ExpectRollback()

	affected, err := s.repository.UpdateMany(ctx, ids, updates)
	s.Error(err)
	s.Equal(int64(0), affected)
}

// TestDeleteMany tests the DeleteMany method of the repository
func (s *RepositoryMockTestSuite) TestDeleteMany() {
	ctx := context.Background()
	ids := []interface{}{"1", "2"}

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`DELETE FROM "test_models"`).
		WithArgs("1", "2").
		WillReturnResult(sqlmock.NewResult(0, 2))
	s.mock.ExpectCommit()

	affected, err := s.repository.DeleteMany(ctx, ids)
	s.NoError(err)
	s.Equal(int64(2), affected)
}

// TestDeleteMany_Error tests error handling in DeleteMany
func (s *RepositoryMockTestSuite) TestDeleteMany_Error() {
	ctx := context.Background()
	ids := []interface{}{"1", "2"}

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`DELETE FROM "test_models"`).
		WithArgs("1", "2").
		WillReturnError(errors.New("database error"))
	s.mock.ExpectRollback()

	affected, err := s.repository.DeleteMany(ctx, ids)
	s.Error(err)
	s.Equal(int64(0), affected)
}

// TestWithTransaction tests the WithTransaction method
func TestWithTransaction(t *testing.T) {
	db := setupSQLiteTestDB(t)
	repo := NewGenericRepositoryWithResource(db, resource.NewResource(resource.ResourceConfig{
		Model: &TestModel{},
	}))

	ctx := context.Background()

	// Test successful transaction
	t.Run("Successful transaction", func(t *testing.T) {
		// Create initial data outside the transaction
		initialModel := &TestModel{
			ID:    "tx-test-1",
			Name:  "Before Transaction",
			Email: "test@example.com",
		}
		_, err := repo.Create(ctx, initialModel)
		assert.NoError(t, err)

		// Run transaction that updates the data
		err = repo.WithTransaction(func(txRepo Repository) error {
			// Retrieve the model
			model, err := txRepo.Get(ctx, "tx-test-1")
			if err != nil {
				return err
			}

			// Update the model
			modelToUpdate := model.(*TestModel)
			modelToUpdate.Name = "After Transaction"
			_, err = txRepo.Update(ctx, "tx-test-1", modelToUpdate)
			return err
		})

		assert.NoError(t, err)

		// Verify that the transaction was committed by retrieving the updated data
		updatedModel, err := repo.Get(ctx, "tx-test-1")
		assert.NoError(t, err)
		assert.Equal(t, "After Transaction", updatedModel.(*TestModel).Name)
	})

	// Test transaction rollback
	t.Run("Transaction rollback", func(t *testing.T) {
		// Create initial data outside the transaction
		initialModel := &TestModel{
			ID:    "tx-test-2",
			Name:  "Before Rollback",
			Email: "rollback@example.com",
		}
		_, err := repo.Create(ctx, initialModel)
		assert.NoError(t, err)

		// Run transaction that updates the data but returns an error to trigger rollback
		err = repo.WithTransaction(func(txRepo Repository) error {
			// Retrieve the model
			model, err := txRepo.Get(ctx, "tx-test-2")
			if err != nil {
				return err
			}

			// Update the model
			modelToUpdate := model.(*TestModel)
			modelToUpdate.Name = "After Rollback (should not be committed)"
			_, err = txRepo.Update(ctx, "tx-test-2", modelToUpdate)
			if err != nil {
				return err
			}

			// Return an error to trigger rollback
			return errors.New("rollback error")
		})

		assert.Error(t, err)
		assert.Equal(t, "rollback error", err.Error())

		// Verify that the transaction was rolled back
		nonUpdatedModel, err := repo.Get(ctx, "tx-test-2")
		assert.NoError(t, err)
		assert.Equal(t, "Before Rollback", nonUpdatedModel.(*TestModel).Name)
	})
}

// TestGormDBPropagation tests that the GORM DB instance is correctly propagated
func TestGormDBPropagation(t *testing.T) {
	// Create a unique database name with timestamp to avoid conflicts
	dbName := fmt.Sprintf("file:transaction_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	testDB, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)

	// Migrate schema
	err = testDB.AutoMigrate(&TestModel{})
	require.NoError(t, err)

	repo := NewGenericRepositoryWithResource(testDB, resource.NewResource(resource.ResourceConfig{
		Model: &TestModel{},
	}))

	ctx := context.Background()

	// Test that Query returns a proper DB instance that can be used for operations
	t.Run("Query returns usable DB", func(t *testing.T) {
		db := repo.Query(ctx)
		assert.NotNil(t, db)

		// Try to use the returned DB
		model := &TestModel{
			ID:    "db-test-1",
			Name:  "Test DB",
			Email: "db@example.com",
		}

		err := db.Create(model).Error
		assert.NoError(t, err)

		// Verify it was saved
		var result TestModel
		err = db.Where("id = ?", "db-test-1").First(&result).Error
		assert.NoError(t, err)
		assert.Equal(t, "Test DB", result.Name)
	})
}

// TestFindOneBy tests the FindOneBy method
func TestFindOneBy(t *testing.T) {
	// Create a unique database name to prevent test interference
	dbName := fmt.Sprintf("file:find_one_by_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&TestModel{})
	require.NoError(t, err)

	// Create the repository
	repo := NewGenericRepositoryWithResource(db, resource.NewResource(resource.ResourceConfig{
		Model: &TestModel{},
	}))

	ctx := context.Background()

	// Create test data
	testModels := []TestModel{
		{ID: "find-one-1", Name: "Find One Test 1", Email: "one@example.com", Age: 25},
		{ID: "find-one-2", Name: "Find One Test 2", Email: "two@example.com", Age: 30},
		{ID: "find-one-3", Name: "Find One Test 3", Email: "three@example.com", Age: 35},
	}

	for _, model := range testModels {
		result := db.Create(&model)
		require.NoError(t, result.Error)
	}

	// Test finding by single condition
	t.Run("Find by single field", func(t *testing.T) {
		result, err := repo.FindOneBy(ctx, map[string]interface{}{"name": "Find One Test 2"})
		assert.NoError(t, err)
		assert.NotNil(t, result)

		model := result.(*TestModel)
		assert.Equal(t, "find-one-2", model.ID)
		assert.Equal(t, "Find One Test 2", model.Name)
		assert.Equal(t, "two@example.com", model.Email)
		assert.Equal(t, 30, model.Age)
	})

	// Test finding by multiple conditions
	t.Run("Find by multiple fields", func(t *testing.T) {
		result, err := repo.FindOneBy(ctx, map[string]interface{}{
			"name": "Find One Test 3",
			"age":  35,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)

		model := result.(*TestModel)
		assert.Equal(t, "find-one-3", model.ID)
		assert.Equal(t, "Find One Test 3", model.Name)
		assert.Equal(t, 35, model.Age)
	})

	// Test error when not found
	t.Run("Error when not found", func(t *testing.T) {
		result, err := repo.FindOneBy(ctx, map[string]interface{}{"name": "Nonexistent"})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})
}

// TestFindAllBy tests the FindAllBy method
func TestFindAllBy(t *testing.T) {
	// Create a unique database name to prevent test interference
	dbName := fmt.Sprintf("file:find_all_by_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&TestModel{})
	require.NoError(t, err)

	// Create the repository
	repo := NewGenericRepositoryWithResource(db, resource.NewResource(resource.ResourceConfig{
		Model: &TestModel{},
	}))

	ctx := context.Background()

	// Create test data
	testModels := []TestModel{
		{ID: "find-all-1", Name: "Find All Test", Email: "all1@example.com", Age: 25},
		{ID: "find-all-2", Name: "Find All Test", Email: "all2@example.com", Age: 30},
		{ID: "find-all-3", Name: "Different Name", Email: "diff@example.com", Age: 35},
		{ID: "find-all-4", Name: "Find All Test", Email: "all3@example.com", Age: 25},
	}

	for _, model := range testModels {
		result := db.Create(&model)
		require.NoError(t, result.Error)
	}

	// Test finding by name
	t.Run("Find by name", func(t *testing.T) {
		result, err := repo.FindAllBy(ctx, map[string]interface{}{"name": "Find All Test"})
		assert.NoError(t, err)
		assert.NotNil(t, result)

		models := *(result.(*[]TestModel))
		assert.Len(t, models, 3)

		// Verify IDs of the results
		ids := []string{models[0].ID, models[1].ID, models[2].ID}
		assert.Contains(t, ids, "find-all-1")
		assert.Contains(t, ids, "find-all-2")
		assert.Contains(t, ids, "find-all-4")
	})

	// Test finding by age
	t.Run("Find by age", func(t *testing.T) {
		result, err := repo.FindAllBy(ctx, map[string]interface{}{"age": 25})
		assert.NoError(t, err)
		assert.NotNil(t, result)

		models := *(result.(*[]TestModel))
		assert.Len(t, models, 2)

		// Verify IDs of the results
		ids := []string{models[0].ID, models[1].ID}
		assert.Contains(t, ids, "find-all-1")
		assert.Contains(t, ids, "find-all-4")
	})

	// Test finding with multiple conditions
	t.Run("Find with multiple conditions", func(t *testing.T) {
		result, err := repo.FindAllBy(ctx, map[string]interface{}{
			"name": "Find All Test",
			"age":  25,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)

		models := *(result.(*[]TestModel))
		assert.Len(t, models, 2)

		// Verify IDs of the results
		ids := []string{models[0].ID, models[1].ID}
		assert.Contains(t, ids, "find-all-1")
		assert.Contains(t, ids, "find-all-4")
	})

	// Test empty result
	t.Run("Empty result", func(t *testing.T) {
		result, err := repo.FindAllBy(ctx, map[string]interface{}{"name": "Nonexistent"})
		assert.NoError(t, err)
		assert.NotNil(t, result)

		models := *(result.(*[]TestModel))
		assert.Len(t, models, 0)
	})
}
