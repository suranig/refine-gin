package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/handler"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Domain model with JSON configuration
type Domain struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"uniqueIndex" refine:"required"`
	Config    Config    `json:"config" gorm:"type:jsonb"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Config represents a nested JSON structure
type Config struct {
	Email     EmailConfig  `json:"email,omitempty"`
	OAuth     OAuthConfig  `json:"oauth,omitempty"`
	Features  FeatureFlags `json:"features,omitempty"`
	Theme     string       `json:"theme,omitempty"`
	Active    bool         `json:"active,omitempty"`
	CreatedAt time.Time    `json:"created_at,omitempty"`
}

// EmailConfig for mail server settings
type EmailConfig struct {
	Host     string `json:"host,omitempty" validate:"required"`
	Port     int    `json:"port,omitempty" validate:"min=1,max=65535"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	From     string `json:"from,omitempty" validate:"email"`
}

// OAuthConfig for authentication settings
type OAuthConfig struct {
	GoogleClientID     string `json:"google_client_id,omitempty"`
	GoogleClientSecret string `json:"google_client_secret,omitempty"`
	GoogleRedirectURL  string `json:"google_redirect_url,omitempty"`
	GitHubClientID     string `json:"github_client_id,omitempty"`
	GitHubClientSecret string `json:"github_client_secret,omitempty"`
	GitHubRedirectURL  string `json:"github_redirect_url,omitempty"`
}

// FeatureFlags for toggling features
type FeatureFlags struct {
	EnableRegistration bool `json:"enable_registration,omitempty"`
	EnableOAuth        bool `json:"enable_oauth,omitempty"`
	EnableAPI          bool `json:"enable_api,omitempty"`
	RateLimiting       bool `json:"rate_limiting,omitempty"`
}

func main() {
	// Create a new Gin router
	r := gin.Default()

	// Connect to SQLite database
	db, err := gorm.Open(sqlite.Open("domains.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate the schema
	db.AutoMigrate(&Domain{})

	// Create a domain resource with JSON configuration
	domainResource := resource.NewResource(resource.ResourceConfig{
		Name:  "domains",
		Label: "Domains",
		Icon:  "domain",
		Model: Domain{},
		Fields: []resource.Field{
			{
				Name:  "id",
				Type:  "int",
				Label: "ID",
			},
			{
				Name:  "name",
				Type:  "string",
				Label: "Domain Name",
				Validation: &resource.Validation{
					Required: true,
				},
			},
			{
				Name:  "config",
				Type:  "json",
				Label: "Configuration",
				Json: &resource.JsonConfig{
					DefaultExpanded: true,
					EditorType:      "form",
					Properties: []resource.JsonProperty{
						{
							Path:  "email",
							Label: "Email Configuration",
							Type:  "object",
							Properties: []resource.JsonProperty{
								{
									Path:  "email.host",
									Label: "SMTP Host",
									Type:  "string",
									Validation: &resource.Validation{
										Required: true,
									},
									Form: &resource.FormConfig{
										Placeholder: "smtp.example.com",
										Help:        "Enter your SMTP server host",
									},
								},
								{
									Path:  "email.port",
									Label: "SMTP Port",
									Type:  "number",
									Validation: &resource.Validation{
										Required: true,
										Min:      1,
										Max:      65535,
									},
									Form: &resource.FormConfig{
										Placeholder: "25, 465, 587, etc.",
										Help:        "Standard ports: 25 (SMTP), 465 (SMTPS), 587 (Submission)",
									},
								},
								{
									Path:  "email.from",
									Label: "From Email",
									Type:  "string",
									Validation: &resource.Validation{
										Pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
										Message: "Must be a valid email address",
									},
								},
							},
						},
						{
							Path:  "oauth",
							Label: "OAuth Settings",
							Type:  "object",
							Properties: []resource.JsonProperty{
								{
									Path:  "oauth.google_client_id",
									Label: "Google Client ID",
									Type:  "string",
								},
								{
									Path:  "oauth.google_client_secret",
									Label: "Google Client Secret",
									Type:  "string",
									Form: &resource.FormConfig{
										Tooltip: "Keep this secure!",
									},
								},
							},
						},
						{
							Path:  "features",
							Label: "Feature Flags",
							Type:  "object",
							Properties: []resource.JsonProperty{
								{
									Path:  "features.enable_registration",
									Label: "Enable User Registration",
									Type:  "boolean",
								},
								{
									Path:  "features.enable_oauth",
									Label: "Enable OAuth Login",
									Type:  "boolean",
								},
							},
						},
						{
							Path:  "active",
							Label: "Active",
							Type:  "boolean",
						},
					},
				},
			},
			{
				Name:  "created_at",
				Type:  "time",
				Label: "Created At",
			},
			{
				Name:  "updated_at",
				Type:  "time",
				Label: "Updated At",
			},
		},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
		},
		SearchableFields: []string{"name"},
		FilterableFields: []string{"id", "name", "active"},
		SortableFields:   []string{"id", "name", "created_at", "updated_at"},
		FormFields:       []string{"name", "config"},
		TableFields:      []string{"id", "name", "created_at", "updated_at"},
	})

	// Create a repository
	repo := repository.NewGenericRepository(db, domainResource)

	// Register API routes
	api := r.Group("/api")
	handler.RegisterResource(api, domainResource, repo)

	// Add some sample data
	insertSampleData(db)

	// Start the server
	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func insertSampleData(db *gorm.DB) {
	// Check if we already have data
	var count int64
	db.Model(&Domain{}).Count(&count)
	if count > 0 {
		return
	}

	// Create a sample domain
	domain := Domain{
		Name: "example.com",
		Config: Config{
			Email: EmailConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "user@example.com",
				Password: "password123",
				From:     "no-reply@example.com",
			},
			OAuth: OAuthConfig{
				GoogleClientID:     "google-client-id",
				GoogleClientSecret: "google-client-secret",
				GoogleRedirectURL:  "https://example.com/oauth/google/callback",
			},
			Features: FeatureFlags{
				EnableRegistration: true,
				EnableOAuth:        true,
				EnableAPI:          true,
				RateLimiting:       false,
			},
			Theme:  "default",
			Active: true,
		},
	}

	db.Create(&domain)
	log.Println("Sample data inserted")
}
