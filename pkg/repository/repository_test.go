package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
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

	// Test Create
	model := &TestModel{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	createdModel, err := s.repository.Create(ctx, model)
	s.NoError(err)
	s.NotNil(createdModel)

	// Test Get
	retrievedModel, err := s.repository.Get(ctx, model.ID)
	s.NoError(err)
	s.Equal(model.Name, retrievedModel.(*TestModel).Name)

	// Test List
	models, total, err := s.repository.List(ctx, query.QueryOptions{
		Page:    1,
		PerPage: 10,
	})
	s.NoError(err)
	s.Equal(int64(1), total)

	modelsList, ok := models.(*[]TestModel)
	s.True(ok)
	s.Len(*modelsList, 1)
	s.Equal(model.ID, (*modelsList)[0].ID)

	// Test Update
	model.Name = "Jane Doe"
	updatedModel, err := s.repository.Update(ctx, model.ID, model)
	s.NoError(err)
	s.Equal("Jane Doe", updatedModel.(*TestModel).Name)

	// Test Delete
	err = s.repository.Delete(ctx, model.ID)
	s.NoError(err)

	// Verify deletion
	_, err = s.repository.Get(ctx, model.ID)
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
	db := setupSQLiteTestDB(t)
	repo := NewGenericRepositoryWithResource(db, resource.NewResource(resource.ResourceConfig{
		Model: &TestModel{},
	}))

	ctx := context.Background()

	// Test Create
	model := &TestModel{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	createdModel, err := repo.Create(ctx, model)
	assert.NoError(t, err)
	assert.NotNil(t, createdModel)

	// Test Get
	retrievedModel, err := repo.Get(ctx, model.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.Name, retrievedModel.(*TestModel).Name)

	// Test List
	models, total, err := repo.List(ctx, query.QueryOptions{
		Page:    1,
		PerPage: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)

	modelsList, ok := models.(*[]TestModel)
	assert.True(t, ok)
	assert.Len(t, *modelsList, 1)
	assert.Equal(t, model.ID, (*modelsList)[0].ID)

	// Test Update
	model.Name = "Jane Doe"
	updatedModel, err := repo.Update(ctx, model.ID, model)
	assert.NoError(t, err)
	assert.Equal(t, "Jane Doe", updatedModel.(*TestModel).Name)

	// Test Delete
	err = repo.Delete(ctx, model.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.Get(ctx, model.ID)
	assert.Error(t, err)
}

func TestRepositoryWithCustomID(t *testing.T) {
	db := setupSQLiteTestDB(t)
	res := resource.NewResource(resource.ResourceConfig{
		Model:       &TestModelWithCustomID{},
		IDFieldName: "UID",
	})
	repo := NewGenericRepositoryWithResource(db, res)

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
		Sort:    "UID",
		Order:   "asc",
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

	// Test Delete
	err = repo.Delete(ctx, model.UID)
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
	// Create mock data
	data := []TestModel{
		{Name: "Test 1", Age: 20},
		{Name: "Test 2", Age: 30},
	}

	// Setup expect query to be executed
	s.mock.ExpectBegin()
	s.mock.ExpectExec(`INSERT INTO "test_models"`).
		WithArgs(sqlmock.AnyArg(), "Test 1", sqlmock.AnyArg(), 20, sqlmock.AnyArg(), "Test 2", sqlmock.AnyArg(), 30).
		WillReturnResult(sqlmock.NewResult(0, 2))
	s.mock.ExpectCommit()

	// Call repository method
	result, err := s.repository.CreateMany(context.Background(), data)

	// Assert no error
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)

	// Verify that expectations were met
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

// TestCreateMany_Error tests error handling in CreateMany
func (s *RepositoryMockTestSuite) TestCreateMany_Error() {
	// Create mock data
	data := []TestModel{
		{Name: "Test 1", Age: 20},
	}

	// Setup expect query to be executed with error
	s.mock.ExpectBegin()
	s.mock.ExpectExec(`INSERT INTO "test_models"`).
		WithArgs(sqlmock.AnyArg(), "Test 1", sqlmock.AnyArg(), 20).
		WillReturnError(errors.New("database error"))
	s.mock.ExpectRollback()

	// Call repository method
	result, err := s.repository.CreateMany(context.Background(), data)

	// Assert error
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)

	// Verify that expectations were met
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

// TestCreateMany_InvalidType tests passing non-slice to CreateMany
func (s *RepositoryMockTestSuite) TestCreateMany_InvalidType() {
	// Create invalid data (not a slice)
	data := TestModel{Name: "Test 1", Age: 20}

	// Call repository method with non-slice
	result, err := s.repository.CreateMany(context.Background(), data)

	// Assert error
	assert.Error(s.T(), err)
	assert.Equal(s.T(), resource.ErrInvalidType, err)
	assert.Nil(s.T(), result)
}

// TestUpdateMany tests the UpdateMany method of the repository
func (s *RepositoryMockTestSuite) TestUpdateMany() {
	// Create data to update
	data := TestModel{Name: "Updated Name"}
	ids := []interface{}{1, 2}

	// Setup expect query to be executed
	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "test_models" SET "name"=\$1 WHERE ID IN \(\$2,\$3\)`).
		WithArgs("Updated Name", 1, 2).
		WillReturnResult(sqlmock.NewResult(0, 2))
	s.mock.ExpectCommit()

	// Call repository method
	count, err := s.repository.UpdateMany(context.Background(), ids, data)

	// Assert no error and correct count
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(2), count)

	// Verify that expectations were met
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

// TestUpdateMany_Error tests error handling in UpdateMany
func (s *RepositoryMockTestSuite) TestUpdateMany_Error() {
	// Create data to update
	data := TestModel{Name: "Updated Name"}
	ids := []interface{}{1, 2}

	// Setup expect query to be executed with error
	s.mock.ExpectBegin()
	s.mock.ExpectExec(`UPDATE "test_models" SET "name"=\$1 WHERE ID IN \(\$2,\$3\)`).
		WithArgs("Updated Name", 1, 2).
		WillReturnError(errors.New("database error"))
	s.mock.ExpectRollback()

	// Call repository method
	count, err := s.repository.UpdateMany(context.Background(), ids, data)

	// Assert error
	assert.Error(s.T(), err)
	assert.Equal(s.T(), int64(0), count)

	// Verify that expectations were met
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

// TestDeleteMany tests the DeleteMany method of the repository
func (s *RepositoryMockTestSuite) TestDeleteMany() {
	// Create ids to delete
	ids := []interface{}{1, 2}

	// Setup expect query to be executed
	s.mock.ExpectBegin()
	s.mock.ExpectExec(`DELETE FROM "test_models" WHERE ID IN \(\$1,\$2\)`).
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(0, 2))
	s.mock.ExpectCommit()

	// Call repository method
	count, err := s.repository.DeleteMany(context.Background(), ids)

	// Assert no error and correct count
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(2), count)

	// Verify that expectations were met
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

// TestDeleteMany_Error tests error handling in DeleteMany
func (s *RepositoryMockTestSuite) TestDeleteMany_Error() {
	// Create ids to delete
	ids := []interface{}{1, 2}

	// Setup expect query to be executed with error
	s.mock.ExpectBegin()
	s.mock.ExpectExec(`DELETE FROM "test_models" WHERE ID IN \(\$1,\$2\)`).
		WithArgs(1, 2).
		WillReturnError(errors.New("database error"))
	s.mock.ExpectRollback()

	// Call repository method
	count, err := s.repository.DeleteMany(context.Background(), ids)

	// Assert error
	assert.Error(s.T(), err)
	assert.Equal(s.T(), int64(0), count)

	// Verify that expectations were met
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}
