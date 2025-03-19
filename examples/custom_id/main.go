package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/suranig/refine-gin/pkg/handler"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User is a model with a custom ID field named UID
type User struct {
	UID   string `json:"uid" gorm:"primaryKey"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// BeforeCreate is a GORM hook that generates a UUID for the UID field if it's empty
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UID == "" {
		u.UID = uuid.New().String()
	}
	return nil
}

// Product is a model with a custom ID field named GUID
type Product struct {
	GUID        string  `json:"guid" gorm:"primaryKey"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

// BeforeCreate is a GORM hook that generates a UUID for the GUID field if it's empty
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.GUID == "" {
		p.GUID = uuid.New().String()
	}
	return nil
}

func main() {
	// Create Gin router
	r := gin.Default()

	// Setup database
	db, err := gorm.Open(sqlite.Open("custom_id.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate models
	err = db.AutoMigrate(&User{}, &Product{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Create repository factory
	repoFactory := repository.NewGormRepositoryFactory(db)

	// Create API group
	api := r.Group("/api")

	// Create and register User resource with custom ID field
	userResource := resource.NewResource(resource.ResourceConfig{
		Name:        "users",
		Model:       User{},
		IDFieldName: "UID", // Specify custom ID field name
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
		},
	})

	// Create repository for User resource
	userRepo := repoFactory.CreateRepository(userResource)

	// Register User resource with custom ID parameter
	handler.RegisterResourceWithOptions(api, userResource, userRepo,
		handler.RegisterOptionsToResourceOptions(handler.RegisterOptions{
			IDParamName: "uid", // Specify custom URL parameter name
		}))

	// Create and register Product resource with custom ID field
	productResource := resource.NewResource(resource.ResourceConfig{
		Name:        "products",
		Model:       Product{},
		IDFieldName: "GUID", // Specify custom ID field name
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
		},
	})

	// Create repository for Product resource
	productRepo := repoFactory.CreateRepository(productResource)

	// Register Product resource with custom ID parameter
	handler.RegisterResourceWithOptions(api, productResource, productRepo,
		handler.RegisterOptionsToResourceOptions(handler.RegisterOptions{
			IDParamName: "guid", // Specify custom URL parameter name
		}))

	// Start server
	r.Run(":8080")
}
