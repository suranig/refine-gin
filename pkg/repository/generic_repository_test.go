package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Test models
type TestCategory struct {
	ID        uint          `gorm:"primarykey" json:"id"`
	Name      string        `json:"name"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Products  []TestProduct `gorm:"foreignKey:CategoryID"`
}

type TestProduct struct {
	ID         uint         `gorm:"primarykey" json:"id"`
	Name       string       `json:"name"`
	Price      float64      `json:"price"`
	InStock    bool         `json:"in_stock"`
	CategoryID uint         `json:"category_id"`
	Category   TestCategory `gorm:"foreignKey:ID;references:CategoryID"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

// setupTestDB przygotowuje bazę danych testową
func setupTestDB(t *testing.T) *gorm.DB {
	// Use a unique database name for each test to avoid shared state
	dbName := fmt.Sprintf("file:memdb%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// Migracja tabel
	err = db.AutoMigrate(&TestCategory{}, &TestProduct{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return db
}

// createTestData tworzy przykładowe dane testowe
func createTestData(t *testing.T, db *gorm.DB) (TestCategory, TestProduct) {
	// Tworzenie kategorii
	category := TestCategory{
		Name: "Electronics",
	}
	err := db.Create(&category).Error
	if err != nil {
		t.Fatalf("failed to create test category: %v", err)
	}

	// Tworzenie produktu
	product := TestProduct{
		Name:       "Smartphone",
		Price:      999.99,
		InStock:    true,
		CategoryID: category.ID,
	}
	err = db.Create(&product).Error
	if err != nil {
		t.Fatalf("failed to create test product: %v", err)
	}

	return category, product
}

// Structures for JSON field update tests
type TestConfig struct {
	Enabled bool   `json:"enabled"`
	Note    string `json:"note"`
}

type TestJSONModel struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	Name      string     `json:"name"`
	Config    TestConfig `json:"config" gorm:"serializer:json"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func setupJSONTestDB(t *testing.T) *gorm.DB {
	dbName := fmt.Sprintf("file:jsondb%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	if err := db.AutoMigrate(&TestJSONModel{}); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	return db
}

func createJSONTestData(t *testing.T, db *gorm.DB) TestJSONModel {
	m := TestJSONModel{
		Name: "json model",
		Config: TestConfig{
			Enabled: true,
			Note:    "initial",
		},
	}
	if err := db.Create(&m).Error; err != nil {
		t.Fatalf("failed to create json model: %v", err)
	}
	return m
}

func TestGenericRepository_Basic(t *testing.T) {
	db := setupTestDB(t)
	_, product := createTestData(t, db)

	// Utworzenie zasobu i repozytorium
	productResource := resource.NewResource(resource.ResourceConfig{
		Name:  "products",
		Model: TestProduct{},
	})

	repo := NewGenericRepository(db, productResource)

	// Test Get
	ctx := context.Background()
	result, err := repo.Get(ctx, product.ID)
	assert.NoError(t, err)
	assert.Equal(t, product.ID, result.(*TestProduct).ID)
	assert.Equal(t, product.Name, result.(*TestProduct).Name)

	// Test List
	options := query.QueryOptions{
		Resource: productResource,
		Page:     1,
		PerPage:  10,
	}
	listResult, count, err := repo.List(ctx, options)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Cast to the correct type and check length
	products := listResult.(*[]TestProduct)
	assert.Equal(t, 1, len(*products))

	// Test Create
	newProduct := TestProduct{
		Name:       "Laptop",
		Price:      1499.99,
		InStock:    true,
		CategoryID: product.CategoryID,
	}
	createResult, err := repo.Create(ctx, &newProduct)
	assert.NoError(t, err)
	assert.NotEqual(t, uint(0), createResult.(*TestProduct).ID)

	// Test Update
	newProduct.Name = "Updated Laptop"
	updateResult, err := repo.Update(ctx, newProduct.ID, &newProduct)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Laptop", updateResult.(*TestProduct).Name)

	// Test Delete
	err = repo.Delete(ctx, newProduct.ID)
	assert.NoError(t, err)

	// Weryfikacja usunięcia
	_, err = repo.Get(ctx, newProduct.ID)
	assert.Error(t, err)
}

func TestGenericRepository_Relations(t *testing.T) {
	db := setupTestDB(t)
	category, product := createTestData(t, db)

	// Utworzenie zasobu i repozytorium
	productResource := resource.NewResource(resource.ResourceConfig{
		Name:  "products",
		Model: TestProduct{},
	})

	repo := NewGenericRepository(db, productResource)

	// Test GetWithRelations
	ctx := context.Background()
	result, err := repo.GetWithRelations(ctx, product.ID, []string{"Category"})
	assert.NoError(t, err)
	assert.Equal(t, product.ID, result.(*TestProduct).ID)
	assert.Equal(t, category.ID, result.(*TestProduct).Category.ID)
	assert.Equal(t, category.Name, result.(*TestProduct).Category.Name)

	// Test ListWithRelations
	options := query.QueryOptions{
		Resource: productResource,
		Page:     1,
		PerPage:  10,
	}
	listResult, count, err := repo.ListWithRelations(ctx, options, []string{"Category"})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	products := listResult.(*[]TestProduct)
	assert.Len(t, *products, 1)
	assert.Equal(t, category.ID, (*products)[0].Category.ID)
	assert.Equal(t, category.Name, (*products)[0].Category.Name)
}

func TestGenericRepository_FindMethods(t *testing.T) {
	db := setupTestDB(t)
	category, _ := createTestData(t, db)

	// Dodanie dodatkowych produktów
	products := []TestProduct{
		{Name: "Phone", Price: 799.99, InStock: true, CategoryID: category.ID},
		{Name: "Tablet", Price: 599.99, InStock: false, CategoryID: category.ID},
		{Name: "Smartwatch", Price: 299.99, InStock: true, CategoryID: category.ID},
	}
	for _, p := range products {
		db.Create(&p)
	}

	// Utworzenie zasobu i repozytorium
	productResource := resource.NewResource(resource.ResourceConfig{
		Name:  "products",
		Model: TestProduct{},
	})

	repo := NewGenericRepository(db, productResource)
	ctx := context.Background()

	// Test FindOneBy
	result, err := repo.FindOneBy(ctx, map[string]interface{}{"name": "Tablet"})
	assert.NoError(t, err)
	assert.Equal(t, "Tablet", result.(*TestProduct).Name)
	assert.Equal(t, 599.99, result.(*TestProduct).Price)

	// Test FindAllBy
	inStockProducts, err := repo.FindAllBy(ctx, map[string]interface{}{"in_stock": true})
	assert.NoError(t, err)
	assert.Len(t, *(inStockProducts.(*[]TestProduct)), 3) // 2 utworzone + 1 z createTestData

	// Test Query z własnymi warunkami
	var expensiveProducts []TestProduct
	err = repo.Query(ctx).Where("price > ?", 500).Find(&expensiveProducts).Error
	assert.NoError(t, err)
	assert.Len(t, expensiveProducts, 3) // Phone, Tablet, Smartphone
}

func TestGenericRepository_Transaction(t *testing.T) {
	db := setupTestDB(t)
	_, product := createTestData(t, db)

	productResource := resource.NewResource(resource.ResourceConfig{
		Name:  "products",
		Model: TestProduct{},
	})

	repo := NewGenericRepository(db, productResource)
	ctx := context.Background()

	err := repo.WithTransaction(func(txRepo Repository) error {
		product.Price = 1099.99
		_, err := txRepo.Update(ctx, product.ID, &product)
		return err
	})
	assert.NoError(t, err)

	updatedProduct, err := repo.Get(ctx, product.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1099.99, updatedProduct.(*TestProduct).Price)

	origPrice := updatedProduct.(*TestProduct).Price
	err = repo.WithTransaction(func(txRepo Repository) error {
		product.Price = 899.99
		_, err := txRepo.Update(ctx, product.ID, &product)
		if err != nil {
			return err
		}

		return gorm.ErrRecordNotFound
	})
	assert.Error(t, err)

	rollbackProduct, err := repo.Get(ctx, product.ID)
	assert.NoError(t, err)
	assert.Equal(t, origPrice, rollbackProduct.(*TestProduct).Price)
}

func TestGenericRepository_BulkOperations(t *testing.T) {
	db := setupTestDB(t)
	category, _ := createTestData(t, db)

	productResource := resource.NewResource(resource.ResourceConfig{
		Name:  "products",
		Model: TestProduct{},
	})

	repo := NewGenericRepository(db, productResource)
	ctx := context.Background()

	newProducts := []TestProduct{
		{Name: "Headphones", Price: 99.99, InStock: true, CategoryID: category.ID},
		{Name: "Speaker", Price: 199.99, InStock: true, CategoryID: category.ID},
		{Name: "Keyboard", Price: 79.99, InStock: false, CategoryID: category.ID},
	}
	err := repo.BulkCreate(ctx, &newProducts)
	assert.NoError(t, err)

	for _, p := range newProducts {
		assert.NotEqual(t, uint(0), p.ID)
	}

	// Test QueryOptions
	options := query.QueryOptions{
		Resource: productResource,
		Page:     1,
		PerPage:  10,
	}
	_, count, err := repo.List(ctx, options)
	assert.NoError(t, err)
	assert.Equal(t, int64(4), count) // 1 z createTestData + 3 z BulkCreate

	// Test BulkUpdate
	err = repo.BulkUpdate(ctx, map[string]interface{}{"in_stock": false}, map[string]interface{}{"price": 49.99})
	assert.NoError(t, err)

	discountedProducts, err := repo.FindAllBy(ctx, map[string]interface{}{"price": 49.99})
	assert.NoError(t, err)
	assert.Len(t, *(discountedProducts.(*[]TestProduct)), 1) // Tylko jeden produkt był z in_stock=false
}

func TestGenericRepository_RepositoryBulkOperations(t *testing.T) {
	db := setupTestDB(t)
	category, _ := createTestData(t, db)

	productResource := resource.NewResource(resource.ResourceConfig{
		Name:  "products",
		Model: TestProduct{},
	})

	repo := NewGenericRepository(db, productResource)
	ctx := context.Background()

	// Test CreateMany
	newProducts := []TestProduct{
		{Name: "Headphones", Price: 99.99, InStock: true, CategoryID: category.ID},
		{Name: "Speaker", Price: 199.99, InStock: true, CategoryID: category.ID},
		{Name: "Keyboard", Price: 79.99, InStock: false, CategoryID: category.ID},
	}

	createdProducts, err := repo.CreateMany(ctx, &newProducts)
	assert.NoError(t, err)
	assert.NotNil(t, createdProducts)

	products := createdProducts.(*[]TestProduct)
	for _, p := range *products {
		assert.NotEqual(t, uint(0), p.ID) // ID powinno być przypisane
	}

	var productIDs []interface{}
	for _, p := range *products {
		if p.InStock {
			productIDs = append(productIDs, p.ID)
		}
	}

	assert.Len(t, productIDs, 2)

	updateCount, err := repo.UpdateMany(ctx, productIDs, map[string]interface{}{"price": 149.99})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), updateCount)

	for _, id := range productIDs {
		product, err := repo.Get(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, 149.99, product.(*TestProduct).Price)
	}

	// Test DeleteMany
	deleteCount, err := repo.DeleteMany(ctx, productIDs)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), deleteCount)

	options := query.QueryOptions{
		Resource: productResource,
		Page:     1,
		PerPage:  10,
	}
	_, count, err := repo.List(ctx, options)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count) // 1 z createTestData + 1 z CreateMany (tylko ten z InStock=false)
}

func TestGenericRepository_WithRelationsChaining(t *testing.T) {
	db := setupTestDB(t)
	category, product := createTestData(t, db)

	productResource := resource.NewResource(resource.ResourceConfig{
		Name:  "products",
		Model: TestProduct{},
	})

	repo := NewGenericRepository(db, productResource)

	repoWith := repo.WithRelations("Category")
	ctx := context.Background()

	result, err := repoWith.Get(ctx, product.ID)
	assert.NoError(t, err)
	assert.Equal(t, category.ID, result.(*TestProduct).Category.ID)
	assert.Equal(t, category.Name, result.(*TestProduct).Category.Name)

	options := query.QueryOptions{
		Resource: productResource,
		Page:     1,
		PerPage:  10,
	}
	listResult, count, err := repoWith.List(ctx, options)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	products := listResult.(*[]TestProduct)
	assert.Len(t, *products, 1)
	assert.Equal(t, category.ID, (*products)[0].Category.ID)
	assert.Equal(t, category.Name, (*products)[0].Category.Name)
}

func TestGenericRepository_WithRelationsPreload(t *testing.T) {
	db := setupTestDB(t)
	productResource := resource.NewResource(resource.ResourceConfig{
		Name:  "products",
		Model: TestProduct{},
	})

	repo := NewGenericRepository(db, productResource)
	repoWith := repo.WithRelations("Category").(*GenericRepository)

	_, ok := repoWith.DB.Statement.Preloads["Category"]
	assert.True(t, ok)
}

func TestGenericRepository_GetWithRelations(t *testing.T) {
	db := setupTestDB(t)
	category, product := createTestData(t, db)

	productResource := resource.NewResource(resource.ResourceConfig{
		Name:  "products",
		Model: TestProduct{},
	})

	repo := NewGenericRepository(db, productResource)
	ctx := context.Background()
	result, err := repo.GetWithRelations(ctx, product.ID, []string{"Category"})
	assert.NoError(t, err)
	prod := result.(*TestProduct)
	assert.Equal(t, category.ID, prod.Category.ID)
	assert.Equal(t, category.Name, prod.Category.Name)
}

func TestGenericRepository_ListWithRelations(t *testing.T) {
	db := setupTestDB(t)
	category, _ := createTestData(t, db)

	productResource := resource.NewResource(resource.ResourceConfig{
		Name:  "products",
		Model: TestProduct{},
	})

	repo := NewGenericRepository(db, productResource)
	ctx := context.Background()
	options := query.QueryOptions{
		Resource: productResource,
		Page:     1,
		PerPage:  10,
	}
	listResult, count, err := repo.ListWithRelations(ctx, options, []string{"Category"})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
	products := listResult.(*[]TestProduct)
	assert.Len(t, *products, 1)
	assert.Equal(t, category.ID, (*products)[0].Category.ID)
	assert.Equal(t, category.Name, (*products)[0].Category.Name)
}

func TestGenericRepository_WithTransactionRollbackOnError(t *testing.T) {
	db := setupTestDB(t)
	_, product := createTestData(t, db)

	productResource := resource.NewResource(resource.ResourceConfig{
		Name:  "products",
		Model: TestProduct{},
	})

	repo := NewGenericRepository(db, productResource)
	ctx := context.Background()

	origPrice := product.Price
	err := repo.WithTransaction(func(txRepo Repository) error {
		product.Price = 1234.56
		if _, err := txRepo.Update(ctx, product.ID, &product); err != nil {
			return err
		}
		return errors.New("force rollback")
	})
	assert.Error(t, err)

	after, err := repo.Get(ctx, product.ID)
	assert.NoError(t, err)
	assert.Equal(t, origPrice, after.(*TestProduct).Price)
}

func TestGenericRepository_BulkCreate_Error(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	productResource := resource.NewResource(resource.ResourceConfig{
		Name:  "products",
		Model: TestProduct{},
	})
	repo := NewGenericRepository(db, productResource)
	ctx := context.Background()

	items := []TestProduct{{Name: "Broken"}}
	err = repo.BulkCreate(ctx, &items)
	assert.Error(t, err)
}

func TestGenericRepository_BulkUpdate_Error(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	productResource := resource.NewResource(resource.ResourceConfig{
		Name:  "products",
		Model: TestProduct{},
	})
	repo := NewGenericRepository(db, productResource)
	ctx := context.Background()

	err = repo.BulkUpdate(ctx, map[string]interface{}{"name": "foo"}, map[string]interface{}{"price": 1})
	assert.Error(t, err)
}

func TestGenericRepository_Update(t *testing.T) {
	t.Run("MapUpdateWithJSON", func(t *testing.T) {
		db := setupJSONTestDB(t)
		initial := createJSONTestData(t, db)

		jsonResource := resource.NewResource(resource.ResourceConfig{
			Name:  "jsons",
			Model: TestJSONModel{},
		})
		repo := NewGenericRepository(db, jsonResource)

		updates := map[string]interface{}{
			"name": "updated",
			"config": map[string]interface{}{
				"enabled": false,
				"note":    "changed",
			},
		}

		ctx := context.Background()
		result, err := repo.Update(ctx, initial.ID, updates)
		require.NoError(t, err)

		updated := result.(*TestJSONModel)
		assert.Equal(t, initial.ID, updated.ID)
		assert.Equal(t, "updated", updated.Name)
		assert.False(t, updated.Config.Enabled)
		assert.Equal(t, "changed", updated.Config.Note)

		// Verify persisted changes
		var check TestJSONModel
		err = db.First(&check, initial.ID).Error
		require.NoError(t, err)
		assert.Equal(t, updated.Name, check.Name)
		assert.Equal(t, updated.Config, check.Config)
	})

	t.Run("InvalidMapData", func(t *testing.T) {
		db := setupJSONTestDB(t)
		initial := createJSONTestData(t, db)

		jsonResource := resource.NewResource(resource.ResourceConfig{
			Name:  "jsons",
			Model: TestJSONModel{},
		})
		repo := NewGenericRepository(db, jsonResource)

		updates := map[string]interface{}{
			"config": "not a map",
		}

		ctx := context.Background()
		_, err := repo.Update(ctx, initial.ID, updates)
		assert.Error(t, err)
	})
}
