package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/suranig/refine-gin/pkg/handler"
	"github.com/suranig/refine-gin/pkg/middleware"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
	"github.com/suranig/refine-gin/pkg/swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Define models
type User struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" refine:"filterable;searchable"`
	Email     string    `json:"email" refine:"filterable"`
	Role      string    `json:"role" refine:"filterable"`
	CreatedAt time.Time `json:"created_at" refine:"filterable;sortable"`
}

type Task struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" refine:"filterable;searchable"`
	Description string    `json:"description"`
	Status      string    `json:"status" refine:"filterable"`
	OwnerID     string    `json:"ownerId" gorm:"column:owner_id" refine:"filterable"`
	CreatedAt   time.Time `json:"created_at" refine:"filterable;sortable"`
}

type Note struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" refine:"filterable;searchable"`
	Content   string    `json:"content"`
	OwnerID   string    `json:"ownerId" gorm:"column:owner_id" refine:"filterable"`
	CreatedAt time.Time `json:"created_at" refine:"filterable;sortable"`
}

// JWT middleware to extract and validate tokens
func jwtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In a real application, you would validate the token here
		// For the sake of this example, we'll simulate a valid token

		// Create a mock JWT claims with user ID
		claims := jwt.MapClaims{
			"sub": "user-1", // A user ID for the owner
			"exp": time.Now().Add(time.Hour * 24).Unix(),
		}

		// Store claims in context for later extraction
		c.Set("claims", claims)
		c.Next()
	}
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

	// Setup database
	db, err := gorm.Open(sqlite.Open("owner_resources.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate schemas
	db.AutoMigrate(&User{}, &Task{}, &Note{})

	// Seed database if empty
	var userCount int64
	db.Model(&User{}).Count(&userCount)
	if userCount == 0 {
		// Create users
		users := []User{
			{ID: "user-1", Name: "John Doe", Email: "john@example.com", Role: "admin", CreatedAt: time.Now()},
			{ID: "user-2", Name: "Jane Smith", Email: "jane@example.com", Role: "user", CreatedAt: time.Now()},
		}

		for _, user := range users {
			db.Create(&user)
		}

		// Create tasks owned by users
		tasks := []Task{
			{ID: "task-1", Title: "Complete project", Description: "Finish the owner resources implementation", Status: "in-progress", OwnerID: "user-1", CreatedAt: time.Now()},
			{ID: "task-2", Title: "Review code", Description: "Review PR from team member", Status: "pending", OwnerID: "user-1", CreatedAt: time.Now()},
			{ID: "task-3", Title: "Create documentation", Description: "Document the new API", Status: "pending", OwnerID: "user-2", CreatedAt: time.Now()},
		}

		for _, task := range tasks {
			db.Create(&task)
		}

		// Create notes owned by users
		notes := []Note{
			{ID: "note-1", Title: "Meeting notes", Content: "Discussed project timeline and requirements", OwnerID: "user-1", CreatedAt: time.Now()},
			{ID: "note-2", Title: "Ideas", Content: "New feature ideas for next sprint", OwnerID: "user-1", CreatedAt: time.Now()},
			{ID: "note-3", Title: "Personal reminder", Content: "Don't forget to call the client", OwnerID: "user-2", CreatedAt: time.Now()},
		}

		for _, note := range notes {
			db.Create(&note)
		}
	}

	// Create resources
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
	})

	// Standard task resource (no ownership)
	taskResource := resource.NewResource(resource.ResourceConfig{
		Name:  "tasks",
		Model: Task{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationRead,
			resource.OperationCreate,
			resource.OperationUpdate,
			resource.OperationDelete,
			resource.OperationCreateMany,
			resource.OperationUpdateMany,
			resource.OperationDeleteMany,
		},
	})

	// Convert note resource to owner resource
	noteResource := resource.NewResource(resource.ResourceConfig{
		Name:  "notes",
		Model: Note{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationRead,
			resource.OperationCreate,
			resource.OperationUpdate,
			resource.OperationDelete,
			resource.OperationCreateMany,
			resource.OperationUpdateMany,
			resource.OperationDeleteMany,
		},
	})

	// Convert note to owner resource
	ownerNoteResource := resource.NewOwnerResource(noteResource, resource.OwnerConfig{
		OwnerField:       "OwnerID",
		EnforceOwnership: true,
	})

	// Create repositories
	userRepo := repository.NewGenericRepositoryWithResource(db, userResource)
	taskRepo := repository.NewGenericRepositoryWithResource(db, taskResource)

	// Create owner repository for notes
	noteRepo, err := repository.NewOwnerRepository(db, ownerNoteResource)
	if err != nil {
		log.Fatalf("Failed to create owner repository: %v", err)
	}

	// API group
	api := r.Group("/api")

	// Register standard resources
	handler.RegisterResource(api, userResource, userRepo)
	handler.RegisterResource(api, taskResource, taskRepo)

	// Register owner resource with JWT middleware
	securedApi := api.Group("")
	securedApi.Use(jwtMiddleware())
	securedApi.Use(middleware.OwnerContext(middleware.ExtractOwnerIDFromJWT("sub")))
	handler.RegisterOwnerResource(securedApi, ownerNoteResource, noteRepo)

	// Configure Swagger info
	swaggerInfo := swagger.SwaggerInfo{
		Title:       "Owner Resources Example API",
		Description: "API documentation for the owner resources example",
		Version:     "1.0.0",
		BasePath:    "/api",
	}

	// Register Swagger routes with owner resources
	swagger.RegisterSwaggerWithOwnerResources(
		r.Group(""),
		[]resource.Resource{userResource, taskResource},
		[]resource.OwnerResource{ownerNoteResource},
		swaggerInfo,
	)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Start the server
	log.Printf("Server starting on http://localhost:9004")
	if err := r.Run(":9004"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
