package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

	// Utworzenie zasobu i repozytorium
	productResource := resource.NewResource(resource.ResourceConfig{
		Name:  "products",
		Model: TestProduct{},
	})

	repo := NewGenericRepository(db, productResource)
	ctx := context.Background()

	// Test udanej transakcji
	err := repo.WithTransaction(func(txRepo Repository) error {
		// Aktualizacja produktu w transakcji
		product.Price = 1099.99
		_, err := txRepo.Update(ctx, product.ID, &product)
		return err
	})
	assert.NoError(t, err)

	// Weryfikacja aktualizacji
	updatedProduct, err := repo.Get(ctx, product.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1099.99, updatedProduct.(*TestProduct).Price)

	// Test niepowodzącej się transakcji
	origPrice := updatedProduct.(*TestProduct).Price
	err = repo.WithTransaction(func(txRepo Repository) error {
		// Aktualizacja produktu w transakcji
		product.Price = 899.99
		_, err := txRepo.Update(ctx, product.ID, &product)
		if err != nil {
			return err
		}

		// Zwracamy błąd, który spowoduje rollback
		return gorm.ErrRecordNotFound
	})
	assert.Error(t, err)

	// Weryfikacja, że aktualizacja została wycofana
	rollbackProduct, err := repo.Get(ctx, product.ID)
	assert.NoError(t, err)
	assert.Equal(t, origPrice, rollbackProduct.(*TestProduct).Price)
}

func TestGenericRepository_BulkOperations(t *testing.T) {
	db := setupTestDB(t)
	category, _ := createTestData(t, db)

	// Utworzenie zasobu i repozytorium
	productResource := resource.NewResource(resource.ResourceConfig{
		Name:  "products",
		Model: TestProduct{},
	})

	repo := NewGenericRepository(db, productResource)
	ctx := context.Background()

	// Test BulkCreate
	newProducts := []TestProduct{
		{Name: "Headphones", Price: 99.99, InStock: true, CategoryID: category.ID},
		{Name: "Speaker", Price: 199.99, InStock: true, CategoryID: category.ID},
		{Name: "Keyboard", Price: 79.99, InStock: false, CategoryID: category.ID},
	}
	err := repo.BulkCreate(ctx, &newProducts)
	assert.NoError(t, err)

	// Weryfikacja utworzonych produktów
	for _, p := range newProducts {
		assert.NotEqual(t, uint(0), p.ID) // ID powinno być przypisane
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

	// Weryfikacja aktualizacji
	discountedProducts, err := repo.FindAllBy(ctx, map[string]interface{}{"price": 49.99})
	assert.NoError(t, err)
	assert.Len(t, *(discountedProducts.(*[]TestProduct)), 1) // Tylko jeden produkt był z in_stock=false
}
