package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/handler"
	"github.com/suranig/refine-gin/pkg/middleware"
	"github.com/suranig/refine-gin/pkg/naming"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Product model with custom ID field
type Product struct {
	Code        string    `json:"code" gorm:"primaryKey" refine:"filterable;sortable;searchable"`
	Name        string    `json:"name" refine:"filterable;sortable;searchable"`
	Description string    `json:"description" refine:"searchable"`
	Price       float64   `json:"price" refine:"filterable;sortable"`
	Category    string    `json:"category" refine:"filterable;sortable"`
	InStock     bool      `json:"in_stock" refine:"filterable"`
	CreatedAt   time.Time `json:"created_at" refine:"filterable;sortable"`
	UpdatedAt   time.Time `json:"updated_at" refine:"sortable"`
}

// Implement custom repository for Product
type ProductRepository struct {
	db *gorm.DB
}

func (r *ProductRepository) List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error) {
	var products []Product
	var total int64

	q := r.db.Model(&Product{})
	total, err := options.ApplyWithPagination(q, &products)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductRepository) Get(ctx context.Context, id interface{}) (interface{}, error) {
	var product Product
	if err := r.db.First(&product, "code = ?", id).Error; err != nil {
		return nil, err
	}
	return product, nil
}

func (r *ProductRepository) Create(ctx context.Context, data interface{}) (interface{}, error) {
	product := data.(*Product)
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	if err := r.db.Create(product).Error; err != nil {
		return nil, err
	}
	return product, nil
}

func (r *ProductRepository) Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error) {
	product := data.(*Product)
	product.Code = id.(string)
	product.UpdatedAt = time.Now()

	if err := r.db.Save(product).Error; err != nil {
		return nil, err
	}
	return product, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id interface{}) error {
	return r.db.Delete(&Product{}, "code = ?", id).Error
}

func (r *ProductRepository) Count(ctx context.Context, options query.QueryOptions) (int64, error) {
	// For this example we return a mock count
	return int64(len(products)), nil
}

// CreateMany creates multiple products at once
func (r *ProductRepository) CreateMany(ctx context.Context, data interface{}) (interface{}, error) {
	productsData, ok := data.([]Product)
	if !ok {
		return nil, fmt.Errorf("invalid data type, expected []Product")
	}

	// Generate IDs for new products
	for i := range productsData {
		productsData[i].ID = fmt.Sprintf("%d", len(products)+i+1)
		products = append(products, productsData[i])
	}

	return productsData, nil
}

// UpdateMany updates multiple products at once
func (r *ProductRepository) UpdateMany(ctx context.Context, ids []interface{}, data interface{}) (int64, error) {
	productData, ok := data.(Product)
	if !ok {
		return 0, fmt.Errorf("invalid data type, expected Product")
	}

	var count int64
	for _, id := range ids {
		idStr := fmt.Sprintf("%v", id)
		for i, p := range products {
			if p.ID == idStr {
				// Preserve ID and update other fields
				products[i].Name = productData.Name
				products[i].Price = productData.Price
				products[i].Category = productData.Category
				products[i].Description = productData.Description
				products[i].Material = productData.Material
				products[i].Condition = productData.Condition
				count++
				break
			}
		}
	}

	return count, nil
}

// DeleteMany deletes multiple products at once
func (r *ProductRepository) DeleteMany(ctx context.Context, ids []interface{}) (int64, error) {
	var newProducts []Product
	var count int64

	for _, p := range products {
		keep := true
		for _, id := range ids {
			idStr := fmt.Sprintf("%v", id)
			if p.ID == idStr {
				keep = false
				count++
				break
			}
		}
		if keep {
			newProducts = append(newProducts, p)
		}
	}

	products = newProducts
	return count, nil
}

// Order model with relations
type Order struct {
	ID         string    `json:"id" gorm:"primaryKey" refine:"filterable;sortable"`
	CustomerID string    `json:"customer_id" refine:"filterable"`
	Status     string    `json:"status" refine:"filterable;sortable"`
	Total      float64   `json:"total" refine:"filterable;sortable"`
	Items      []Item    `json:"items" gorm:"foreignKey:OrderID" relation:"resource=items;type=one-to-many;field=order_id;reference=id;include=true"`
	CreatedAt  time.Time `json:"created_at" refine:"filterable;sortable"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Item model (order line item)
type Item struct {
	ID        string  `json:"id" gorm:"primaryKey" refine:"filterable"`
	OrderID   string  `json:"order_id" gorm:"index" refine:"filterable"`
	ProductID string  `json:"product_id" refine:"filterable"`
	Quantity  int     `json:"quantity" refine:"filterable"`
	Price     float64 `json:"price" refine:"filterable"`
}

func setupSampleData(db *gorm.DB) error {
	// Create sample products
	products := []Product{
		{
			Code:        "P001",
			Name:        "Smartphone",
			Description: "Latest model smartphone with great features",
			Price:       799.99,
			Category:    "Electronics",
			InStock:     true,
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			Code:        "P002",
			Name:        "Laptop",
			Description: "High-performance laptop for professionals",
			Price:       1299.99,
			Category:    "Electronics",
			InStock:     true,
			CreatedAt:   time.Now().Add(-25 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-25 * 24 * time.Hour),
		},
		{
			Code:        "P003",
			Name:        "Headphones",
			Description: "Noise-cancelling wireless headphones",
			Price:       199.99,
			Category:    "Electronics",
			InStock:     true,
			CreatedAt:   time.Now().Add(-20 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-20 * 24 * time.Hour),
		},
		{
			Code:        "P004",
			Name:        "Coffee Maker",
			Description: "Programmable coffee maker with built-in grinder",
			Price:       129.99,
			Category:    "Home",
			InStock:     true,
			CreatedAt:   time.Now().Add(-15 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-15 * 24 * time.Hour),
		},
		{
			Code:        "P005",
			Name:        "Desk Chair",
			Description: "Ergonomic desk chair with lumbar support",
			Price:       249.99,
			Category:    "Furniture",
			InStock:     false,
			CreatedAt:   time.Now().Add(-10 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-10 * 24 * time.Hour),
		},
	}

	for _, product := range products {
		if err := db.Create(&product).Error; err != nil {
			return err
		}
	}

	// Create sample orders and items
	orders := []Order{
		{
			ID:         "O001",
			CustomerID: "C001",
			Status:     "Completed",
			Total:      999.98,
			CreatedAt:  time.Now().Add(-5 * 24 * time.Hour),
			UpdatedAt:  time.Now().Add(-5 * 24 * time.Hour),
		},
		{
			ID:         "O002",
			CustomerID: "C002",
			Status:     "Processing",
			Total:      1299.99,
			CreatedAt:  time.Now().Add(-3 * 24 * time.Hour),
			UpdatedAt:  time.Now().Add(-3 * 24 * time.Hour),
		},
		{
			ID:         "O003",
			CustomerID: "C001",
			Status:     "Shipped",
			Total:      329.98,
			CreatedAt:  time.Now().Add(-1 * 24 * time.Hour),
			UpdatedAt:  time.Now().Add(-1 * 24 * time.Hour),
		},
	}

	for _, order := range orders {
		if err := db.Create(&order).Error; err != nil {
			return err
		}
	}

	items := []Item{
		{
			ID:        "I001",
			OrderID:   "O001",
			ProductID: "P001",
			Quantity:  1,
			Price:     799.99,
		},
		{
			ID:        "I002",
			OrderID:   "O001",
			ProductID: "P003",
			Quantity:  1,
			Price:     199.99,
		},
		{
			ID:        "I003",
			OrderID:   "O002",
			ProductID: "P002",
			Quantity:  1,
			Price:     1299.99,
		},
		{
			ID:        "I004",
			OrderID:   "O003",
			ProductID: "P003",
			Quantity:  1,
			Price:     199.99,
		},
		{
			ID:        "I005",
			OrderID:   "O003",
			ProductID: "P004",
			Quantity:  1,
			Price:     129.99,
		},
	}

	for _, item := range items {
		if err := db.Create(&item).Error; err != nil {
			return err
		}
	}

	return nil
}

func main() {
	// Create a new Gin router
	r := gin.Default()

	// Enable CORS for Refine.dev frontend
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup database
	db, err := gorm.Open(sqlite.Open("refine_integration.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate models
	err = db.AutoMigrate(&Product{}, &Order{}, &Item{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Check if we need to seed the database
	var count int64
	db.Model(&Product{}).Count(&count)
	if count == 0 {
		if err := setupSampleData(db); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}
		log.Println("Database seeded with sample data")
	}

	// Create repositories
	productRepo := &ProductRepository{db: db}
	orderRepo := repository.NewGormRepositoryWithResource(db, resource.NewResource(resource.ResourceConfig{
		Name:        "orders",
		Model:       Order{},
		IDFieldName: "ID",
	}))
	itemRepo := repository.NewGormRepositoryWithResource(db, resource.NewResource(resource.ResourceConfig{
		Name:        "items",
		Model:       Item{},
		IDFieldName: "ID",
	}))

	// Create resources
	productResource := resource.NewResource(resource.ResourceConfig{
		Name:        "products",
		Model:       Product{},
		IDFieldName: "Code", // Use custom ID field
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
			resource.OperationCount,
		},
		DefaultSort: &resource.Sort{
			Field: "created_at",
			Order: "desc",
		},
	})

	orderResource := resource.NewResource(resource.ResourceConfig{
		Name:  "orders",
		Model: Order{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
			resource.OperationCount,
		},
		DefaultSort: &resource.Sort{
			Field: "created_at",
			Order: "desc",
		},
	})

	itemResource := resource.NewResource(resource.ResourceConfig{
		Name:  "items",
		Model: Item{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
			resource.OperationCount,
		},
	})

	// Create API group with camelCase naming convention for Refine.dev
	api := r.Group("/api", middleware.NamingConventionMiddleware(naming.CamelCase))

	// Register product resource with custom ID parameter
	handler.RegisterResourceForRefine(api, productResource, productRepo, "code")

	// Register order and item resources
	handler.RegisterResourceForRefine(api, orderResource, orderRepo, "")
	handler.RegisterResourceForRefine(api, itemResource, itemRepo, "")

	// Start server
	log.Println("Server starting on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
