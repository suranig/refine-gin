package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stanxing/refine-gin/pkg/query"
	"github.com/stanxing/refine-gin/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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
