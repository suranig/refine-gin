# Refine-Gin Examples Guide

This guide provides comprehensive examples of various use cases and advanced features in the Refine-Gin framework.

## Table of Contents

1. [Basic CRUD Operations](#basic-crud-operations)
2. [Advanced Field Types](#advanced-field-types)
3. [Authentication & Authorization](#authentication--authorization)
4. [Relations & Joins](#relations--joins)
5. [Custom Handlers](#custom-handlers)
6. [File Upload](#file-upload)
7. [JSON Fields](#json-fields)
8. [Computed Fields](#computed-fields)
9. [Custom Validation](#custom-validation)
10. [Middleware Examples](#middleware-examples)
11. [Testing Examples](#testing-examples)

## Basic CRUD Operations

### Simple User Management

```go
package main

import (
    "time"
    "github.com/gin-gonic/gin"
    "github.com/suranig/refine-gin/pkg/handler"
    "github.com/suranig/refine-gin/pkg/repository"
    "github.com/suranig/refine-gin/pkg/resource"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

type User struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" refine:"filterable;searchable;required"`
    Email     string    `json:"email" refine:"filterable;required;validation=email"`
    Age       int       `json:"age" refine:"filterable;min=18;max=120"`
    CreatedAt time.Time `json:"created_at" refine:"filterable;sortable;readonly"`
}

func main() {
    db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    db.AutoMigrate(&User{})

    userResource := resource.NewResource(resource.ResourceConfig{
        Name:  "users",
        Label: "Users",
        Model: User{},
        Operations: []resource.Operation{
            resource.OperationList,
            resource.OperationRead,
            resource.OperationCreate,
            resource.OperationUpdate,
            resource.OperationDelete,
        },
    })

    userRepo := repository.NewGenericRepositoryWithResource(db, userResource)

    r := gin.Default()
    api := r.Group("/api")
    handler.RegisterResource(api, userResource, userRepo)

    r.Run(":8080")
}
```

## Advanced Field Types

### Rich Form Configuration

```go
type Product struct {
    ID          string    `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" refine:"filterable;searchable;required"`
    Description string    `json:"description" refine:"required"`
    Price       float64   `json:"price" refine:"filterable;min=0"`
    Category    string    `json:"category" refine:"filterable"`
    Tags        []string  `json:"tags" gorm:"serializer:json"`
    Image       string    `json:"image"`
    IsActive    bool      `json:"is_active" refine:"filterable"`
    CreatedAt   time.Time `json:"created_at" refine:"filterable;sortable;readonly"`
}

productResource := resource.NewResource(resource.ResourceConfig{
    Name:  "products",
    Label: "Products",
    Model: Product{},
    Fields: []resource.Field{
        {
            Name:  "name",
            Type:  "string",
            Label: "Product Name",
            Validation: &resource.Validation{
                Required:  true,
                MinLength: 3,
                MaxLength: 100,
            },
            Form: &resource.FormConfig{
                Placeholder: "Enter product name",
                Help:        "Product name must be between 3 and 100 characters",
                WidthPercent: 50,
            },
        },
        {
            Name:  "description",
            Type:  "textarea",
            Label: "Description",
            Validation: &resource.Validation{
                Required:  true,
                MaxLength: 1000,
            },
            Form: &resource.FormConfig{
                Placeholder: "Enter product description",
                Height:      "100px",
            },
        },
        {
            Name:  "price",
            Type:  "number",
            Label: "Price",
            Validation: &resource.Validation{
                Required: true,
                Min:      0,
            },
            Form: &resource.FormConfig{
                Placeholder: "0.00",
                Help:        "Price in USD",
            },
        },
        {
            Name:  "category",
            Type:  "select",
            Label: "Category",
            Select: &resource.SelectConfig{
                Multiple:    false,
                Searchable:  true,
                Clearable:   true,
                Placeholder: "Select category",
            },
            Options: []resource.Option{
                {Value: "electronics", Label: "Electronics"},
                {Value: "clothing", Label: "Clothing"},
                {Value: "books", Label: "Books"},
                {Value: "home", Label: "Home & Garden"},
            },
        },
        {
            Name:  "tags",
            Type:  "multiselect",
            Label: "Tags",
            Select: &resource.SelectConfig{
                Multiple:    true,
                Searchable:  true,
                Creatable:   true,
                Placeholder: "Select or create tags",
            },
            Options: []resource.Option{
                {Value: "featured", Label: "Featured"},
                {Value: "new", Label: "New"},
                {Value: "sale", Label: "Sale"},
                {Value: "popular", Label: "Popular"},
            },
        },
        {
            Name:  "image",
            Type:  "image",
            Label: "Product Image",
            File: &resource.FileConfig{
                AllowedTypes: []string{"image/jpeg", "image/png", "image/webp"},
                MaxSize:      5 * 1024 * 1024, // 5MB
                IsImage:      true,
                MaxWidth:     1200,
                MaxHeight:    800,
                GenerateThumbnails: true,
                ThumbnailSizes: []resource.ThumbnailSize{
                    {Name: "small", Width: 150, Height: 150, KeepAspectRatio: true},
                    {Name: "medium", Width: 300, Height: 300, KeepAspectRatio: true},
                },
            },
        },
        {
            Name:  "is_active",
            Type:  "boolean",
            Label: "Active",
            Form: &resource.FormConfig{
                Help: "Whether this product is available for purchase",
            },
        },
    },
    Operations: []resource.Operation{
        resource.OperationList,
        resource.OperationRead,
        resource.OperationCreate,
        resource.OperationUpdate,
        resource.OperationDelete,
    },
})
```

## Authentication & Authorization

### JWT Authentication with Role-Based Permissions

```go
package main

import (
    "time"
    "github.com/gin-gonic/gin"
    "github.com/suranig/refine-gin/pkg/auth"
    "github.com/suranig/refine-gin/pkg/handler"
    "github.com/suranig/refine-gin/pkg/repository"
    "github.com/suranig/refine-gin/pkg/resource"
)

type User struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" refine:"filterable;searchable;required"`
    Email     string    `json:"email" refine:"filterable;required;validation=email"`
    Role      string    `json:"role" refine:"filterable"`
    Password  string    `json:"password" refine:"hidden"`
    CreatedAt time.Time `json:"created_at" refine:"filterable;sortable;readonly"`
}

func main() {
    // Setup JWT configuration
    jwtConfig := auth.JWTConfig{
        Secret:         "your-secret-key-change-in-production",
        ExpirationTime: time.Hour * 24,
        Issuer:         "my-app",
        Audience:       "my-app-users",
    }

    // Create resources with permissions
    userResource := resource.NewResource(resource.ResourceConfig{
        Name:  "users",
        Label: "Users",
        Model: User{},
        Operations: []resource.Operation{
            resource.OperationList,
            resource.OperationRead,
            resource.OperationCreate,
            resource.OperationUpdate,
            resource.OperationDelete,
        },
        Permissions: map[string][]string{
            "list":   {"admin", "manager"},
            "read":   {"admin", "manager", "user"},
            "create": {"admin"},
            "update": {"admin", "manager"},
            "delete": {"admin"},
        },
    })

    // Setup router with authentication
    r := gin.Default()

    // Public routes
    public := r.Group("/api")
    public.POST("/login", loginHandler(jwtConfig))
    public.POST("/register", registerHandler())

    // Protected routes
    protected := r.Group("/api")
    protected.Use(auth.JWTMiddleware(jwtConfig))
    
    userRepo := repository.NewGenericRepositoryWithResource(db, userResource)
    handler.RegisterResource(protected, userResource, userRepo)

    r.Run(":8080")
}

func loginHandler(jwtConfig auth.JWTConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        var loginData struct {
            Email    string `json:"email" binding:"required,email"`
            Password string `json:"password" binding:"required"`
        }

        if err := c.ShouldBindJSON(&loginData); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        // Validate credentials (implement your own logic)
        user, err := validateCredentials(loginData.Email, loginData.Password)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid credentials"})
            return
        }

        // Generate JWT token
        token, err := auth.GenerateJWTWithStandardClaims(jwtConfig, user.ID, map[string]interface{}{
            "role": user.Role,
            "email": user.Email,
        })

        if err != nil {
            c.JSON(500, gin.H{"error": "Failed to generate token"})
            return
        }

        c.JSON(200, gin.H{
            "token": token,
            "user": gin.H{
                "id":    user.ID,
                "name":  user.Name,
                "email": user.Email,
                "role":  user.Role,
            },
        })
    }
}
```

## Relations & Joins

### User with Posts and Comments

```go
type User struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" refine:"filterable;searchable;required"`
    Email     string    `json:"email" refine:"filterable;required;validation=email"`
    Posts     []Post    `json:"posts" gorm:"foreignKey:UserID"`
    CreatedAt time.Time `json:"created_at" refine:"filterable;sortable;readonly"`
}

type Post struct {
    ID        string     `json:"id" gorm:"primaryKey"`
    Title     string     `json:"title" refine:"filterable;searchable;required"`
    Content   string     `json:"content" refine:"required"`
    UserID    string     `json:"user_id" refine:"filterable"`
    User      User       `json:"user" gorm:"foreignKey:UserID"`
    Comments  []Comment  `json:"comments" gorm:"foreignKey:PostID"`
    CreatedAt time.Time  `json:"created_at" refine:"filterable;sortable;readonly"`
}

type Comment struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Content   string    `json:"content" refine:"required"`
    PostID    string    `json:"post_id" refine:"filterable"`
    UserID    string    `json:"user_id" refine:"filterable"`
    Post      Post      `json:"post" gorm:"foreignKey:PostID"`
    User      User      `json:"user" gorm:"foreignKey:UserID"`
    CreatedAt time.Time `json:"created_at" refine:"filterable;sortable;readonly"`
}

// Create resources with relations
userResource := resource.NewResource(resource.ResourceConfig{
    Name:  "users",
    Label: "Users",
    Model: User{},
    Operations: []resource.Operation{
        resource.OperationList,
        resource.OperationRead,
        resource.OperationCreate,
        resource.OperationUpdate,
        resource.OperationDelete,
    },
    Relations: []resource.Relation{
        {
            Name:         "posts",
            Resource:     "posts",
            Type:         "hasMany",
            ForeignKey:   "user_id",
            LocalKey:     "id",
        },
    },
})

postResource := resource.NewResource(resource.ResourceConfig{
    Name:  "posts",
    Label: "Posts",
    Model: Post{},
    Operations: []resource.Operation{
        resource.OperationList,
        resource.OperationRead,
        resource.OperationCreate,
        resource.OperationUpdate,
        resource.OperationDelete,
    },
    Relations: []resource.Relation{
        {
            Name:         "user",
            Resource:     "users",
            Type:         "belongsTo",
            ForeignKey:   "user_id",
            LocalKey:     "id",
        },
        {
            Name:         "comments",
            Resource:     "comments",
            Type:         "hasMany",
            ForeignKey:   "post_id",
            LocalKey:     "id",
        },
    },
})
```

## Custom Handlers

### Custom Business Logic

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/suranig/refine-gin/pkg/handler"
    "github.com/suranig/refine-gin/pkg/resource"
)

// Custom handler for user registration with validation
func customUserRegistrationHandler(res resource.Resource, repo repository.Repository) gin.HandlerFunc {
    return func(c *gin.Context) {
        var userData map[string]interface{}
        if err := c.ShouldBindJSON(&userData); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        // Custom validation
        if email, exists := userData["email"].(string); exists {
            if isEmailTaken(email, repo) {
                c.JSON(409, gin.H{"error": "Email already exists"})
                return
            }
        }

        // Hash password
        if password, exists := userData["password"].(string); exists {
            hashedPassword, err := hashPassword(password)
            if err != nil {
                c.JSON(500, gin.H{"error": "Failed to hash password"})
                return
            }
            userData["password"] = hashedPassword
        }

        // Create user
        user, err := repo.Create(c.Request.Context(), userData)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        // Send welcome email
        go sendWelcomeEmail(user)

        c.JSON(201, gin.H{"data": user})
    }
}

// Custom handler for bulk operations
func bulkDeleteHandler(res resource.Resource, repo repository.Repository) gin.HandlerFunc {
    return func(c *gin.Context) {
        var request struct {
            IDs []string `json:"ids" binding:"required"`
        }

        if err := c.ShouldBindJSON(&request); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        // Convert string IDs to interface{} slice
        ids := make([]interface{}, len(request.IDs))
        for i, id := range request.IDs {
            ids[i] = id
        }

        // Perform bulk delete
        deletedCount, err := repo.DeleteMany(c.Request.Context(), ids)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, gin.H{
            "message": "Successfully deleted items",
            "deletedCount": deletedCount,
        })
    }
}

// Register custom handlers
func registerCustomHandlers(router *gin.RouterGroup, userResource resource.Resource, userRepo repository.Repository) {
    router.POST("/users/register", customUserRegistrationHandler(userResource, userRepo))
    router.POST("/users/bulk-delete", bulkDeleteHandler(userResource, userRepo))
}
```

## File Upload

### Image Upload with Processing

```go
type Product struct {
    ID          string    `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" refine:"filterable;searchable;required"`
    Description string    `json:"description" refine:"required"`
    Images      []string  `json:"images" gorm:"serializer:json"`
    CreatedAt   time.Time `json:"created_at" refine:"filterable;sortable;readonly"`
}

// Custom file upload handler
func imageUploadHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        file, err := c.FormFile("image")
        if err != nil {
            c.JSON(400, gin.H{"error": "No file uploaded"})
            return
        }

        // Validate file type
        if !isValidImageType(file.Header.Get("Content-Type")) {
            c.JSON(400, gin.H{"error": "Invalid file type"})
            return
        }

        // Validate file size (5MB limit)
        if file.Size > 5*1024*1024 {
            c.JSON(400, gin.H{"error": "File too large"})
            return
        }

        // Generate unique filename
        filename := generateUniqueFilename(file.Filename)
        filepath := "uploads/images/" + filename

        // Save file
        if err := c.SaveUploadedFile(file, filepath); err != nil {
            c.JSON(500, gin.H{"error": "Failed to save file"})
            return
        }

        // Process image (resize, create thumbnails)
        go processImage(filepath)

        c.JSON(200, gin.H{
            "filename": filename,
            "url":      "/uploads/images/" + filename,
            "size":     file.Size,
        })
    }
}

func isValidImageType(contentType string) bool {
    validTypes := []string{
        "image/jpeg",
        "image/png",
        "image/gif",
        "image/webp",
    }

    for _, validType := range validTypes {
        if contentType == validType {
            return true
        }
    }
    return false
}

func processImage(filepath string) {
    // Implement image processing logic
    // - Resize to max dimensions
    // - Create thumbnails
    // - Optimize file size
    // - Add watermarks if needed
}
```

## JSON Fields

### Complex Configuration Storage

```go
type AppConfig struct {
    ID          string                 `json:"id" gorm:"primaryKey"`
    Name        string                 `json:"name" refine:"filterable;required"`
    Environment string                 `json:"environment" refine:"filterable"`
    Settings    map[string]interface{} `json:"settings" gorm:"serializer:json"`
    CreatedAt   time.Time              `json:"created_at" refine:"filterable;sortable;readonly"`
}

appConfigResource := resource.NewResource(resource.ResourceConfig{
    Name:  "app-configs",
    Label: "App Configurations",
    Model: AppConfig{},
    Fields: []resource.Field{
        {
            Name:  "name",
            Type:  "string",
            Label: "Configuration Name",
            Validation: &resource.Validation{
                Required: true,
            },
        },
        {
            Name:  "environment",
            Type:  "select",
            Label: "Environment",
            Select: &resource.SelectConfig{
                Multiple:    false,
                Placeholder: "Select environment",
            },
            Options: []resource.Option{
                {Value: "development", Label: "Development"},
                {Value: "staging", Label: "Staging"},
                {Value: "production", Label: "Production"},
            },
        },
        {
            Name:  "settings",
            Type:  "json",
            Label: "Settings",
            Json: &resource.JsonConfig{
                DefaultExpanded: true,
                EditorType:      "form",
                Properties: []resource.JsonProperty{
                    {
                        Path:  "database.host",
                        Label: "Database Host",
                        Type:  "string",
                        Validation: &resource.JsonValidation{
                            Required: true,
                        },
                    },
                    {
                        Path:  "database.port",
                        Label: "Database Port",
                        Type:  "number",
                        Validation: &resource.JsonValidation{
                            Min: 1,
                            Max: 65535,
                        },
                    },
                    {
                        Path:  "redis.enabled",
                        Label: "Redis Enabled",
                        Type:  "boolean",
                    },
                    {
                        Path:  "redis.host",
                        Label: "Redis Host",
                        Type:  "string",
                        Validation: &resource.JsonValidation{
                            Conditional: &resource.JsonConditionalValidation{
                                Path:     "redis.enabled",
                                Operator: "eq",
                                Value:    true,
                                Message:  "Redis host is required when Redis is enabled",
                            },
                        },
                    },
                    {
                        Path:  "features",
                        Label: "Features",
                        Type:  "object",
                        Properties: []resource.JsonProperty{
                            {
                                Path:  "features.authentication.enabled",
                                Label: "Authentication Enabled",
                                Type:  "boolean",
                            },
                            {
                                Path:  "features.notifications.email",
                                Label: "Email Notifications",
                                Type:  "boolean",
                            },
                            {
                                Path:  "features.notifications.push",
                                Label: "Push Notifications",
                                Type:  "boolean",
                            },
                        },
                    },
                },
            },
        },
    },
    Operations: []resource.Operation{
        resource.OperationList,
        resource.OperationRead,
        resource.OperationCreate,
        resource.OperationUpdate,
        resource.OperationDelete,
    },
})
```

## Computed Fields

### Dynamic Calculations

```go
type Order struct {
    ID          string    `json:"id" gorm:"primaryKey"`
    UserID      string    `json:"user_id" refine:"filterable"`
    Items       []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
    Subtotal    float64   `json:"subtotal" refine:"readonly"`
    Tax         float64   `json:"tax" refine:"readonly"`
    Total       float64   `json:"total" refine:"readonly"`
    Status      string    `json:"status" refine:"filterable"`
    CreatedAt   time.Time `json:"created_at" refine:"filterable;sortable;readonly"`
}

type OrderItem struct {
    ID       string  `json:"id" gorm:"primaryKey"`
    OrderID  string  `json:"order_id"`
    ProductID string `json:"product_id"`
    Quantity int     `json:"quantity" refine:"min=1"`
    Price    float64 `json:"price" refine:"min=0"`
    Total    float64 `json:"total" refine:"readonly"`
}

orderResource := resource.NewResource(resource.ResourceConfig{
    Name:  "orders",
    Label: "Orders",
    Model: Order{},
    Fields: []resource.Field{
        {
            Name:  "subtotal",
            Type:  "computed",
            Label: "Subtotal",
            Computed: &resource.ComputedFieldConfig{
                DependsOn: []string{"items"},
                Expression: "items.reduce((sum, item) => sum + (item.quantity * item.price), 0)",
                ClientSide: true,
                Persist:    true,
            },
        },
        {
            Name:  "tax",
            Type:  "computed",
            Label: "Tax",
            Computed: &resource.ComputedFieldConfig{
                DependsOn: []string{"subtotal"},
                Expression: "subtotal * 0.1", // 10% tax
                ClientSide: true,
                Persist:    true,
            },
        },
        {
            Name:  "total",
            Type:  "computed",
            Label: "Total",
            Computed: &resource.ComputedFieldConfig{
                DependsOn: []string{"subtotal", "tax"},
                Expression: "subtotal + tax",
                ClientSide: true,
                Persist:    true,
            },
        },
    },
    Operations: []resource.Operation{
        resource.OperationList,
        resource.OperationRead,
        resource.OperationCreate,
        resource.OperationUpdate,
        resource.OperationDelete,
    },
})
```

## Custom Validation

### Complex Validation Rules

```go
type User struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" refine:"filterable;searchable;required"`
    Email     string    `json:"email" refine:"filterable;required;validation=email"`
    Age       int       `json:"age" refine:"filterable;min=18;max=120"`
    Password  string    `json:"password" refine:"hidden;required;minLength=8"`
    ConfirmPassword string `json:"confirm_password" refine:"hidden"`
    CreatedAt time.Time `json:"created_at" refine:"filterable;sortable;readonly"`
}

// Custom validation middleware
func customValidationMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method == "POST" || c.Request.Method == "PUT" {
            var userData map[string]interface{}
            if err := c.ShouldBindJSON(&userData); err != nil {
                c.JSON(400, gin.H{"error": err.Error()})
                return
            }

            // Custom password validation
            if password, exists := userData["password"].(string); exists {
                if !isValidPassword(password) {
                    c.JSON(400, gin.H{"error": "Password must contain at least 8 characters, one uppercase, one lowercase, one number, and one special character"})
                    return
                }
            }

            // Password confirmation validation
            if password, exists := userData["password"].(string); exists {
                if confirmPassword, exists := userData["confirm_password"].(string); exists {
                    if password != confirmPassword {
                        c.JSON(400, gin.H{"error": "Passwords do not match"})
                        return
                    }
                }
            }

            // Age validation based on registration date
            if age, exists := userData["age"].(float64); exists {
                if age < 18 {
                    c.JSON(400, gin.H{"error": "User must be at least 18 years old"})
                    return
                }
            }
        }

        c.Next()
    }
}

func isValidPassword(password string) bool {
    if len(password) < 8 {
        return false
    }

    hasUpper := false
    hasLower := false
    hasNumber := false
    hasSpecial := false

    for _, char := range password {
        switch {
        case char >= 'A' && char <= 'Z':
            hasUpper = true
        case char >= 'a' && char <= 'z':
            hasLower = true
        case char >= '0' && char <= '9':
            hasNumber = true
        case char >= 33 && char <= 47 || char >= 58 && char <= 64 || char >= 91 && char <= 96 || char >= 123 && char <= 126:
            hasSpecial = true
        }
    }

    return hasUpper && hasLower && hasNumber && hasSpecial
}
```

## Middleware Examples

### Caching Middleware

```go
package main

import (
    "time"
    "github.com/gin-gonic/gin"
    "github.com/suranig/refine-gin/pkg/middleware"
)

func setupCaching() {
    // Cache configuration for different resources
    userCacheConfig := middleware.CacheConfig{
        Duration: time.Minute * 5,
        Methods:  []string{"GET", "OPTIONS"},
        Headers:  []string{"Authorization"},
    }

    productCacheConfig := middleware.CacheConfig{
        Duration: time.Minute * 10,
        Methods:  []string{"GET", "OPTIONS"},
        Headers:  []string{"Authorization"},
    }

    // Apply caching to specific resources
    router.Use(middleware.CacheByResource("users", userCacheConfig))
    router.Use(middleware.CacheByResource("products", productCacheConfig))
    
    // Disable cache for write operations
    router.Use(middleware.NoCacheMiddleware())
}

// Custom rate limiting middleware
func rateLimitMiddleware() gin.HandlerFunc {
    // Implement rate limiting logic
    return func(c *gin.Context) {
        // Check rate limits
        // If exceeded, return 429 Too Many Requests
        c.Next()
    }
}

// Logging middleware
func loggingMiddleware() gin.HandlerFunc {
    return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
            param.ClientIP,
            param.TimeStamp.Format(time.RFC1123),
            param.Method,
            param.Path,
            param.Request.Proto,
            param.StatusCode,
            param.Latency,
            param.Request.UserAgent(),
            param.ErrorMessage,
        )
    })
}
```

## Testing Examples

### Unit Tests

```go
package main

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "bytes"
    "encoding/json"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/suranig/refine-gin/pkg/handler"
    "github.com/suranig/refine-gin/pkg/repository"
    "github.com/suranig/refine-gin/pkg/resource"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        panic("Failed to connect to test database")
    }
    
    db.AutoMigrate(&User{})
    return db
}

func TestUserCRUD(t *testing.T) {
    // Setup
    db := setupTestDB()
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
    userRepo := repository.NewGenericRepositoryWithResource(db, userResource)

    gin.SetMode(gin.TestMode)
    router := gin.New()
    api := router.Group("/api")
    handler.RegisterResource(api, userResource, userRepo)

    // Test Create
    t.Run("Create User", func(t *testing.T) {
        userData := map[string]interface{}{
            "name":  "John Doe",
            "email": "john@example.com",
            "age":   30,
        }

        jsonData, _ := json.Marshal(userData)
        req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
        req.Header.Set("Content-Type", "application/json")

        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusCreated, w.Code)

        var response map[string]interface{}
        json.Unmarshal(w.Body.Bytes(), &response)
        assert.NotNil(t, response["data"])
    })

    // Test List
    t.Run("List Users", func(t *testing.T) {
        req, _ := http.NewRequest("GET", "/api/users", nil)
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusOK, w.Code)

        var response map[string]interface{}
        json.Unmarshal(w.Body.Bytes(), &response)
        assert.NotNil(t, response["data"])
        assert.NotNil(t, response["total"])
    })

    // Test Get
    t.Run("Get User", func(t *testing.T) {
        req, _ := http.NewRequest("GET", "/api/users/1", nil)
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusOK, w.Code)

        var response map[string]interface{}
        json.Unmarshal(w.Body.Bytes(), &response)
        assert.NotNil(t, response["data"])
    })

    // Test Update
    t.Run("Update User", func(t *testing.T) {
        updateData := map[string]interface{}{
            "name": "Jane Doe",
        }

        jsonData, _ := json.Marshal(updateData)
        req, _ := http.NewRequest("PUT", "/api/users/1", bytes.NewBuffer(jsonData))
        req.Header.Set("Content-Type", "application/json")

        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusOK, w.Code)
    })

    // Test Delete
    t.Run("Delete User", func(t *testing.T) {
        req, _ := http.NewRequest("DELETE", "/api/users/1", nil)
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusOK, w.Code)
    })
}

// Integration test with authentication
func TestAuthenticatedEndpoints(t *testing.T) {
    db := setupTestDB()
    jwtConfig := auth.DefaultJWTConfig()
    jwtConfig.Secret = "test-secret"

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
        Permissions: map[string][]string{
            "list":   {"admin"},
            "read":   {"admin", "user"},
            "create": {"admin"},
            "update": {"admin"},
            "delete": {"admin"},
        },
    })

    userRepo := repository.NewGenericRepositoryWithResource(db, userResource)

    gin.SetMode(gin.TestMode)
    router := gin.New()
    
    // Add JWT middleware
    router.Use(auth.JWTMiddleware(jwtConfig))
    
    api := router.Group("/api")
    handler.RegisterResource(api, userResource, userRepo)

    // Generate test token
    token, _ := auth.GenerateJWTWithStandardClaims(jwtConfig, "user123", map[string]interface{}{
        "role": "admin",
    })

    t.Run("Authenticated Request", func(t *testing.T) {
        req, _ := http.NewRequest("GET", "/api/users", nil)
        req.Header.Set("Authorization", "Bearer "+token)

        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusOK, w.Code)
    })

    t.Run("Unauthenticated Request", func(t *testing.T) {
        req, _ := http.NewRequest("GET", "/api/users", nil)

        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusUnauthorized, w.Code)
    })
}
```

This examples guide demonstrates various advanced features and use cases of the Refine-Gin framework. Each example can be adapted and extended based on your specific requirements.