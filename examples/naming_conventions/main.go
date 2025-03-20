package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/handler"
	"github.com/suranig/refine-gin/pkg/naming"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User is a sample model
type User struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func main() {
	// Create a Gin router
	r := gin.Default()

	// Connect to database
	db, err := gorm.Open(sqlite.Open("example.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate the User model
	db.AutoMigrate(&User{})

	// Create the user resource using DefaultResource
	userResource := resource.NewResource(resource.ResourceConfig{
		Name:  "users",
		Model: &User{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
			resource.OperationCount,
		},
		DefaultSort: &resource.Sort{
			Field: "id",
			Order: "asc",
		},
	})

	// Create the GORM repository
	userRepo := repository.NewGormRepositoryWithResource(db, userResource)

	// Create API group with base path
	api := r.Group("/api")

	// Example 1: Using snake_case (default)
	opts := resource.DefaultOptions().WithNamingConvention(naming.SnakeCase)
	handler.RegisterResourceWithOptions(api.Group("/snake"), userResource, userRepo, opts)

	// Example 2: Using camelCase
	optsCamel := resource.DefaultOptions().WithNamingConvention(naming.CamelCase)
	handler.RegisterResourceWithOptions(api.Group("/camel"), userResource, userRepo, optsCamel)

	// Example 3: Using PascalCase
	optsPascal := resource.DefaultOptions().WithNamingConvention(naming.PascalCase)
	handler.RegisterResourceWithOptions(api.Group("/pascal"), userResource, userRepo, optsPascal)

	// Start the server
	r.Run(":8080")
}
