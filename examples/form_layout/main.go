package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/handler"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User model
type User struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	FirstName string `json:"firstName" form:"firstName" binding:"required"`
	LastName  string `json:"lastName" form:"lastName" binding:"required"`
	Email     string `json:"email" form:"email" binding:"required,email"`
	Age       int    `json:"age" form:"age" binding:"gte=0,lt=130"`
	Address   string `json:"address" form:"address"`
	City      string `json:"city" form:"city" binding:"required_with=Address"`
	Country   string `json:"country" form:"country"`
	Bio       string `json:"bio" form:"bio"`
}

func main() {
	// Initialize Gin
	r := gin.Default()

	// Connect to database
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create tables
	db.AutoMigrate(&User{})

	// Create seed data if needed
	var count int64
	db.Model(&User{}).Count(&count)
	if count == 0 {
		createSeedData(db)
	}

	// Create repository
	userRepo := repository.NewGormRepository[User](db)

	// Create FormLayout
	formLayout := &resource.FormLayout{
		Columns: 2,
		Gutter:  16,
		Sections: []*resource.FormSection{
			{
				ID:    "personalInfo",
				Title: "Personal Information",
				Icon:  "user",
			},
			{
				ID:    "contactInfo",
				Title: "Contact Information",
				Icon:  "mail",
			},
			{
				ID:          "additional",
				Title:       "Additional Information",
				Icon:        "info-circle",
				Collapsible: true,
			},
		},
		FieldLayouts: []*resource.FormFieldLayout{
			{
				Field:     "FirstName",
				SectionID: "personalInfo",
				Column:    0,
				Row:       0,
			},
			{
				Field:     "LastName",
				SectionID: "personalInfo",
				Column:    1,
				Row:       0,
			},
			{
				Field:     "Age",
				SectionID: "personalInfo",
				Column:    0,
				Row:       1,
			},
			{
				Field:     "Email",
				SectionID: "contactInfo",
				Column:    0,
				Row:       0,
				ColSpan:   2,
			},
			{
				Field:     "Address",
				SectionID: "contactInfo",
				Column:    0,
				Row:       1,
			},
			{
				Field:     "City",
				SectionID: "contactInfo",
				Column:    1,
				Row:       1,
			},
			{
				Field:     "Country",
				SectionID: "contactInfo",
				Column:    0,
				Row:       2,
				ColSpan:   2,
			},
			{
				Field:     "Bio",
				SectionID: "additional",
				Column:    0,
				Row:       0,
				ColSpan:   2,
			},
		},
	}

	// Create resource
	userResource := resource.NewDefaultResource(&User{})
	userResource.SetName("user")
	userResource.SetLabel("User")
	userResource.SetFormLayout(formLayout)
	userResource.SetOperations([]resource.Operation{
		resource.OperationList,
		resource.OperationCreate,
		resource.OperationRead,
		resource.OperationUpdate,
		resource.OperationDelete,
	})

	// Register API
	api := r.Group("/api")

	// Register resource endpoints
	handler.RegisterResourceEndpoints(api, userResource)

	// Start server
	fmt.Println("Server running at http://localhost:8080")
	fmt.Println("Form metadata available at http://localhost:8080/api/users/form")
	fmt.Println("Form metadata for editing available at http://localhost:8080/api/users/form/:id")
	http.ListenAndServe(":8080", r)
}

func createSeedData(db *gorm.DB) {
	users := []User{
		{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Age:       30,
			Address:   "123 Main St",
			City:      "New York",
			Country:   "USA",
			Bio:       "Software developer with 5 years experience",
		},
		{
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane.smith@example.com",
			Age:       28,
			Address:   "456 Park Ave",
			City:      "Boston",
			Country:   "USA",
			Bio:       "Product manager who loves to travel",
		},
	}

	for _, user := range users {
		db.Create(&user)
	}
}
