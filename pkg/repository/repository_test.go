package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
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

// RepositoryTestSuite is a test suite for the Repository interface implementations
type RepositoryTestSuite struct {
	suite.Suite
	db         *gorm.DB
	mock       sqlmock.Sqlmock
	repository Repository
	resource   resource.Resource
}

// SetupSuite initializes the test suite with a mock database
func (s *RepositoryTestSuite) SetupSuite() {
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
	s.repository = NewGormRepository(s.db, &TestModel{})
}

// TestRepositorySuite runs the test suite
func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

func TestGormRepository(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate model
	err = db.AutoMigrate(&TestModel{})
	assert.NoError(t, err)

	// Create repository
	repo := NewGormRepository(db, &TestModel{})

	// Create a test resource
	res := resource.NewResource(resource.ResourceConfig{
		Name:  "tests",
		Model: TestModel{},
	})

	// Create a test context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	ctx := context.WithValue(c, "resource", res)

	// Test Create
	model := &TestModel{
		ID:    "1",
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	createdModel, err := repo.Create(ctx, model)
	assert.NoError(t, err)
	assert.Equal(t, model, createdModel)

	// Test Get
	retrievedModel, err := repo.Get(ctx, "1")
	assert.NoError(t, err)
	assert.Equal(t, model, retrievedModel)

	// Test List
	options := query.QueryOptions{
		Resource: res,
		Page:     1,
		PerPage:  10,
		Sort:     "id",
		Order:    "asc",
		Filters:  make(map[string]interface{}),
	}

	models, total, err := repo.List(ctx, options)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)

	// The List method returns a slice of the model type, not a slice of pointers to the model type
	modelsList, ok := models.(*[]TestModel)
	assert.True(t, ok)
	assert.Len(t, *modelsList, 1)
	assert.Equal(t, model.ID, (*modelsList)[0].ID)

	// Test Update
	model.Name = "Jane Doe"
	updatedModel, err := repo.Update(ctx, "1", model)
	assert.NoError(t, err)
	assert.Equal(t, model, updatedModel)

	// Verify update
	retrievedModel, err = repo.Get(ctx, "1")
	assert.NoError(t, err)
	assert.Equal(t, "Jane Doe", retrievedModel.(*TestModel).Name)

	// Test Delete
	err = repo.Delete(ctx, "1")
	assert.NoError(t, err)

	// Verify delete
	_, err = repo.Get(ctx, "1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
}

func TestGormRepositoryWithCustomID(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate model
	err = db.AutoMigrate(&TestModelWithCustomID{})
	assert.NoError(t, err)

	// Create a test resource with custom ID field
	res := resource.NewResource(resource.ResourceConfig{
		Name:        "tests",
		Model:       TestModelWithCustomID{},
		IDFieldName: "UID",
	})

	// Create repository with resource
	repo := NewGormRepositoryWithResource(db, res)

	// Create a test context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	ctx := context.WithValue(c, "resource", res)

	// Test Create
	model := &TestModelWithCustomID{
		UID:   "custom-1",
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	createdModel, err := repo.Create(ctx, model)
	assert.NoError(t, err)
	assert.Equal(t, model, createdModel)

	// Test Get
	retrievedModel, err := repo.Get(ctx, "custom-1")
	assert.NoError(t, err)
	assert.Equal(t, model, retrievedModel)

	// Test List
	options := query.QueryOptions{
		Resource: res,
		Page:     1,
		PerPage:  10,
		Sort:     "uid",
		Order:    "asc",
		Filters:  make(map[string]interface{}),
	}

	models, total, err := repo.List(ctx, options)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)

	// The List method returns a slice of the model type, not a slice of pointers to the model type
	modelsList, ok := models.(*[]TestModelWithCustomID)
	assert.True(t, ok)
	assert.Len(t, *modelsList, 1)
	assert.Equal(t, model.UID, (*modelsList)[0].UID)

	// Test Update
	model.Name = "Jane Doe"
	updatedModel, err := repo.Update(ctx, "custom-1", model)
	assert.NoError(t, err)
	assert.Equal(t, model, updatedModel)

	// Verify update
	retrievedModel, err = repo.Get(ctx, "custom-1")
	assert.NoError(t, err)
	assert.Equal(t, "Jane Doe", retrievedModel.(*TestModelWithCustomID).Name)

	// Test Delete
	err = repo.Delete(ctx, "custom-1")
	assert.NoError(t, err)

	// Verify delete
	_, err = repo.Get(ctx, "custom-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
}

func TestGormRepositoryFactory(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Create factory
	factory := NewGormRepositoryFactory(db)

	// Create a test resource with custom ID field
	res := resource.NewResource(resource.ResourceConfig{
		Name:        "tests",
		Model:       TestModelWithCustomID{},
		IDFieldName: "UID",
	})

	// Create repository using factory
	repo := factory.CreateRepository(res)
	assert.NotNil(t, repo)

	// Sprawdź, czy repozytorium używa niestandardowego pola identyfikatora
	gormRepo, ok := repo.(*GormRepository)
	assert.True(t, ok)
	assert.Equal(t, "UID", gormRepo.IDFieldName)
}

// TestCreateMany tests the CreateMany method of the repository
func (s *RepositoryTestSuite) TestCreateMany() {
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
func (s *RepositoryTestSuite) TestCreateMany_Error() {
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
func (s *RepositoryTestSuite) TestCreateMany_InvalidType() {
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
func (s *RepositoryTestSuite) TestUpdateMany() {
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
func (s *RepositoryTestSuite) TestUpdateMany_Error() {
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
func (s *RepositoryTestSuite) TestDeleteMany() {
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
func (s *RepositoryTestSuite) TestDeleteMany_Error() {
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
