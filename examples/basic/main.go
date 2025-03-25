package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/handler"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
	"github.com/suranig/refine-gin/pkg/swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Sample data
var users []User
var posts []Post

type User struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" refine:"filterable;searchable"`
	Email     string    `json:"email" refine:"filterable"`
	CreatedAt time.Time `json:"created_at" refine:"filterable;sortable"`
}

type Post struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" refine:"filterable;searchable"`
	Content   string    `json:"content"`
	UserID    string    `json:"user_id" refine:"filterable"`
	CreatedAt time.Time `json:"created_at" refine:"filterable;sortable"`
}

func main() {
	// Create a new Gin router
	r := gin.Default()

	// Enable CORS
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

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db.AutoMigrate(&User{}, &Post{})

	var count int64
	db.Model(&User{}).Count(&count)
	if count == 0 {
		users := []User{
			{ID: "1", Name: "John Doe", Email: "john@example.com", CreatedAt: time.Now()},
			{ID: "2", Name: "Jane Smith", Email: "jane@example.com", CreatedAt: time.Now()},
		}

		for _, user := range users {
			db.Create(&user)
		}

		posts := []Post{
			{ID: "1", Title: "First Post", Content: "This is the first post", UserID: "1", CreatedAt: time.Now()},
			{ID: "2", Title: "Second Post", Content: "This is the second post", UserID: "1", CreatedAt: time.Now()},
			{ID: "3", Title: "Jane's Post", Content: "This is Jane's post", UserID: "2", CreatedAt: time.Now()},
		}

		for _, post := range posts {
			db.Create(&post)
		}
	}

	userResource := resource.NewResource(resource.ResourceConfig{
		Name:  "users",
		Model: User{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationRead,
			resource.OperationCreate,
			resource.OperationUpdate,
			resource.OperationDelete,
		},
		DefaultSort: &resource.Sort{
			Field: "created_at",
			Order: string(query.SortOrderDesc),
		},
	})

	postResource := resource.NewResource(resource.ResourceConfig{
		Name:  "posts",
		Model: Post{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationRead,
			resource.OperationCreate,
			resource.OperationUpdate,
			resource.OperationDelete,
		},
	})

	// Create repositories using GenericRepository
	userRepo := repository.NewGenericRepositoryWithResource(db, userResource)
	postRepo := repository.NewGenericRepositoryWithResource(db, postResource)

	// Register resources
	api := r.Group("/api")
	handler.RegisterResource(api, userResource, userRepo)
	handler.RegisterResource(api, postResource, postRepo)

	// Configure Swagger info
	swaggerInfo := swagger.SwaggerInfo{
		Title:       "Basic Example API",
		Description: "API documentation for the basic example",
		Version:     "1.0.0",
		BasePath:    "/api",
	}

	// Register Swagger routes
	swagger.RegisterSwagger(r.Group(""), []resource.Resource{userResource, postResource}, swaggerInfo)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Start the server
	log.Printf("Server starting on http://localhost:9003")
	if err := r.Run(":9003"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
