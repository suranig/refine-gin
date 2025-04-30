package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stanxing/refine-gin/pkg/resource"
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
	userResource := &resource.DefaultResource{
		Name:       "user",
		Label:      "User",
		FormLayout: formLayout,
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
		},
		Model: &User{},
	}

	// Register API
	api := r.Group("/api")

	// Register resource endpoints
	resourceRouter := api.Group(fmt.Sprintf("/%s", userResource.GetName()))

	// Register form metadata endpoint
	resourceRouter.GET("/form", func(c *gin.Context) {
		// Generate form metadata
		fields := resource.GenerateFieldsMetadata(userResource.GetFields())
		layout := resource.GenerateFormLayoutMetadata(userResource.GetFormLayout())

		c.JSON(http.StatusOK, gin.H{
			"fields": fields,
			"layout": layout,
		})
	})

	// Start server
	fmt.Println("Server running at http://localhost:8080")
	fmt.Println("Form metadata available at http://localhost:8080/api/user/form")
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
