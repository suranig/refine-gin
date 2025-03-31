package main

import (
	"fmt"
	"log"
	"net/http"

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

// SetID sets the UID field - implements interface for ID-aware models
func (u *User) SetID(id interface{}) {
	if idStr, ok := id.(string); ok {
		u.UID = idStr
	}
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

// SetID sets the GUID field - implements interface for ID-aware models
func (p *Product) SetID(id interface{}) {
	if idStr, ok := id.(string); ok {
		p.GUID = idStr
	}
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
	repoFactory := repository.NewGenericRepositoryFactory(db)

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
	handler.RegisterResourceForRefine(api, userResource, userRepo, "uid")

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
	handler.RegisterResourceForRefine(api, productResource, productRepo, "guid")

	// Add direct endpoint for debugging update issues
	api.PUT("/users_direct/:uid", func(c *gin.Context) {
		uid := c.Param("uid")
		fmt.Printf("[DIRECT-UPDATE] Received update for UID: %s\n", uid)

		var requestBody map[string]interface{}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			fmt.Printf("[DIRECT-UPDATE] Error binding JSON: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fmt.Printf("[DIRECT-UPDATE] Request body: %+v\n", requestBody)

		var dataToUpdate interface{}
		if data, ok := requestBody["data"]; ok {
			dataToUpdate = data
			fmt.Printf("[DIRECT-UPDATE] Found data field: %+v\n", dataToUpdate)
		} else {
			dataToUpdate = requestBody
			fmt.Printf("[DIRECT-UPDATE] Using whole body: %+v\n", dataToUpdate)
		}

		// Create a new User with the UID
		user := &User{UID: uid, Name: "Updated via Direct", Email: "direct@example.com"}
		fmt.Printf("[DIRECT-UPDATE] Created user object: %+v\n", user)

		// Direct update with repo
		updated, err := userRepo.Update(c.Request.Context(), uid, user)
		if err != nil {
			fmt.Printf("[DIRECT-UPDATE] Update error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		fmt.Printf("[DIRECT-UPDATE] Update succeeded: %+v\n", updated)
		c.JSON(http.StatusOK, gin.H{"data": updated})
	})

	// Start server
	r.Run(":8080")
}
