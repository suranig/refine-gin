package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/suranig/refine-gin/pkg/handler"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Product jest modelem reprezentującym produkt
type Product struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Price       float64    `json:"price"`
	CategoryID  string     `json:"category_id"`
	Category    Category   `json:"category" gorm:"foreignKey:CategoryID;references:ID" refine:"include:true"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// Category jest modelem reprezentującym kategorię produktów
type Category struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Products    []Product  `json:"products,omitempty" gorm:"foreignKey:CategoryID;references:ID" refine:"include:false"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func main() {
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Nie można otworzyć bazy danych: %v", err)
	}

	db.AutoMigrate(&Product{}, &Category{})

	createTestData(db)

	productResource := resource.NewResource(resource.ResourceConfig{
		Name:  "products",
		Model: Product{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
			resource.OperationCount,
		},
	})

	categoryResource := resource.NewResource(resource.ResourceConfig{
		Name:  "categories",
		Model: Category{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
			resource.OperationCount,
		},
	})

	repoFactory := repository.NewGenericRepositoryFactory(db)

	productRepo := repoFactory.CreateRepository(productResource)
	categoryRepo := repoFactory.CreateRepository(categoryResource)

	api := r.Group("/api")
	handler.RegisterResource(api, productResource, productRepo)
	handler.RegisterResource(api, categoryResource, categoryRepo)

	r.GET("/api/products-with-category", func(c *gin.Context) {
		genericRepo := productRepo.(*repository.GenericRepository)

		options := query.QueryOptions{
			Resource: productResource,
			Page:     1,
			PerPage:  10,
			Sort:     "name",
			Order:    "asc",
			Filters:  make(map[string]interface{}),
		}

		products, total, err := genericRepo.ListWithRelations(c, options, []string{"Category"})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  products,
			"total": total,
		})
	})

	r.GET("/api/products/by-category/:categoryId", func(c *gin.Context) {
		categoryId := c.Param("categoryId")

		genericRepo := productRepo.(*repository.GenericRepository)
		products, err := genericRepo.FindAllBy(c, map[string]interface{}{
			"category_id": categoryId,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": products,
		})
	})

	r.POST("/api/transfer-products", func(c *gin.Context) {
		var request struct {
			FromCategoryID string `json:"from_category_id"`
			ToCategoryID   string `json:"to_category_id"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		genericRepo := productRepo.(*repository.GenericRepository)

		err := genericRepo.WithTransaction(func(txRepo repository.Repository) error {
			txGenericRepo := txRepo.(*repository.GenericRepository)

			productsToMove, err := txGenericRepo.FindAllBy(c, map[string]interface{}{
				"category_id": request.FromCategoryID,
			})
			if err != nil {
				return err
			}

			productsSlice, ok := productsToMove.([]Product)
			if !ok {
				return fmt.Errorf("nie można przekonwertować rezultatu na slice produktów")
			}

			for _, product := range productsSlice {
				err := txGenericRepo.BulkUpdate(c, map[string]interface{}{
					"id": product.ID,
				}, map[string]interface{}{
					"category_id": request.ToCategoryID,
				})

				if err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	fmt.Println("Serwer uruchomiony na http://localhost:8080")
	r.Run(":8080")
}

func createTestData(db *gorm.DB) {
	electronics := Category{
		ID:          uuid.New().String(),
		Name:        "Elektronika",
		Description: "Urządzenia elektroniczne",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	clothing := Category{
		ID:          uuid.New().String(),
		Name:        "Odzież",
		Description: "Ubrania i akcesoria",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	db.Create(&electronics)
	db.Create(&clothing)

	products := []Product{
		{
			ID:          uuid.New().String(),
			Name:        "Smartfon",
			Description: "Najnowszy model smartfona",
			Price:       2499.99,
			CategoryID:  electronics.ID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "Laptop",
			Description: "Wydajny laptop do pracy",
			Price:       3999.99,
			CategoryID:  electronics.ID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "Koszulka",
			Description: "Bawełniana koszulka",
			Price:       49.99,
			CategoryID:  clothing.ID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "Spodnie",
			Description: "Dżinsy męskie",
			Price:       99.99,
			CategoryID:  clothing.ID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, product := range products {
		db.Create(&product)
	}
}
