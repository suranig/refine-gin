# Refine-Gin Quick Start Guide

## Getting Started

This guide will help you get up and running with Refine-Gin in minutes. You'll create a simple API with user management functionality.

## Prerequisites

- Go 1.23 or later
- PostgreSQL, MySQL, or SQLite database
- Basic knowledge of Go and REST APIs

## Installation

1. **Create a new Go project:**

```bash
mkdir my-refine-api
cd my-refine-api
go mod init my-refine-api
```

2. **Add Refine-Gin dependency:**

```bash
go get github.com/suranig/refine-gin
```

3. **Add required dependencies:**

```bash
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres  # or mysql, sqlite
```

## Step 1: Define Your Models

Create a `models.go` file:

```go
package main

import (
    "time"
)

type User struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" refine:"filterable;searchable;required"`
    Email     string    `json:"email" refine:"filterable;required;validation=email"`
    Role      string    `json:"role" refine:"filterable"`
    CreatedAt time.Time `json:"created_at" refine:"filterable;sortable;readonly"`
}

type Post struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Title     string    `json:"title" refine:"filterable;searchable;required"`
    Content   string    `json:"content" refine:"required"`
    UserID    string    `json:"user_id" refine:"filterable"`
    CreatedAt time.Time `json:"created_at" refine:"filterable;sortable;readonly"`
}
```

## Step 2: Set Up Database Connection

Create a `database.go` file:

```go
package main

import (
    "log"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func setupDatabase() *gorm.DB {
    // Replace with your database connection string
    dsn := "host=localhost user=postgres password=password dbname=refine_demo port=5432 sslmode=disable"
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    // Auto migrate your models
    err = db.AutoMigrate(&User{}, &Post{})
    if err != nil {
        log.Fatalf("Failed to migrate database: %v", err)
    }

    return db
}
```

## Step 3: Create Resources

Create a `resources.go` file:

```go
package main

import (
    "github.com/suranig/refine-gin/pkg/resource"
)

func createResources() (resource.Resource, resource.Resource) {
    // User resource
    userResource := resource.NewResource(resource.ResourceConfig{
        Name:  "users",
        Label: "Users",
        Icon:  "user",
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
            Order: "desc",
        },
        Permissions: map[string][]string{
            "list":   {"admin", "user"},
            "read":   {"admin", "user"},
            "create": {"admin"},
            "update": {"admin"},
            "delete": {"admin"},
        },
    })

    // Post resource
    postResource := resource.NewResource(resource.ResourceConfig{
        Name:  "posts",
        Label: "Posts",
        Icon:  "file-text",
        Model: Post{},
        Operations: []resource.Operation{
            resource.OperationList,
            resource.OperationRead,
            resource.OperationCreate,
            resource.OperationUpdate,
            resource.OperationDelete,
        },
    })

    return userResource, postResource
}
```

## Step 4: Create Repositories

Create a `repositories.go` file:

```go
package main

import (
    "github.com/suranig/refine-gin/pkg/repository"
    "github.com/suranig/refine-gin/pkg/resource"
    "gorm.io/gorm"
)

func createRepositories(db *gorm.DB, userResource, postResource resource.Resource) (repository.Repository, repository.Repository) {
    // Create repositories
    userRepo := repository.NewGenericRepositoryWithResource(db, userResource)
    postRepo := repository.NewOwnerRepository(db, &Post{}, "user_id", "users")

    return userRepo, postRepo
}
```

## Step 5: Set Up Authentication (Optional)

Create an `auth.go` file:

```go
package main

import (
    "time"
    "github.com/suranig/refine-gin/pkg/auth"
)

func setupAuth() auth.JWTConfig {
    return auth.JWTConfig{
        Secret:         "your-secret-key-change-in-production",
        ExpirationTime: time.Hour * 24,
        Issuer:         "my-refine-api",
        Audience:       "my-refine-api-users",
    }
}
```

## Step 6: Create Main Application

Create `main.go`:

```go
package main

import (
    "log"
    "github.com/gin-gonic/gin"
    "github.com/suranig/refine-gin/pkg/auth"
    "github.com/suranig/refine-gin/pkg/handler"
    "github.com/suranig/refine-gin/pkg/swagger"
)

func main() {
    // Setup database
    db := setupDatabase()

    // Create resources
    userResource, postResource := createResources()

    // Create repositories
    userRepo, postRepo := createRepositories(db, userResource, postResource)

    // Setup router
    r := gin.Default()

    // CORS middleware
    r.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    })

    // Optional: JWT middleware
    // jwtConfig := setupAuth()
    // r.Use(auth.JWTMiddleware(jwtConfig))

    // Register resources
    api := r.Group("/api")
    handler.RegisterResourceForRefine(api, userResource, userRepo, "id")
    handler.RegisterResourceForRefine(api, postResource, postRepo, "id")

    // Swagger documentation
    swaggerInfo := swagger.SwaggerInfo{
        Title:       "My Refine API",
        Description: "API documentation for my Refine-Gin application",
        Version:     "1.0.0",
        BasePath:    "/api",
    }
    swagger.RegisterSwagger(r.Group(""), []resource.Resource{userResource, postResource}, swaggerInfo)

    // Health check
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "ok",
            "message": "Refine-Gin API is running",
        })
    })

    // Start server
    log.Printf("Server starting on http://localhost:8080")
    log.Printf("Swagger UI available at http://localhost:8080/swagger")
    log.Fatal(r.Run(":8080"))
}
```

## Step 7: Run Your Application

1. **Start your database server**

2. **Update the database connection string in `database.go`**

3. **Run the application:**

```bash
go run .
```

4. **Test your API:**

```bash
# Health check
curl http://localhost:8080/health

# Get users (empty initially)
curl http://localhost:8080/api/users

# Create a user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "role": "user"
  }'

# Get users again
curl http://localhost:8080/api/users
```

## Step 8: Explore Your API

### Available Endpoints

- `GET /api/users` - List users
- `POST /api/users` - Create user
- `GET /api/users/:id` - Get user by ID
- `PUT /api/users/:id` - Update user
- `DELETE /api/users/:id` - Delete user
- `GET /api/users/count` - Count users
- `OPTIONS /api/users` - Get user metadata

- `GET /api/posts` - List posts
- `POST /api/posts` - Create post
- `GET /api/posts/:id` - Get post by ID
- `PUT /api/posts/:id` - Update post
- `DELETE /api/posts/:id` - Delete post
- `GET /api/posts/count` - Count posts
- `OPTIONS /api/posts` - Get post metadata

### Query Parameters

```bash
# Pagination
curl "http://localhost:8080/api/users?current=1&pageSize=10"

# Filtering
curl "http://localhost:8080/api/users?name=John"

# Searching
curl "http://localhost:8080/api/users?q=john"

# Sorting
curl "http://localhost:8080/api/users?sort=created_at&order=desc"

# Advanced filtering
curl "http://localhost:8080/api/users?filter[name][contains]=john"
```

### Metadata Endpoint

```bash
# Get resource metadata (useful for frontend)
curl -X OPTIONS http://localhost:8080/api/users
```

## Step 9: Add Sample Data (Optional)

Create a `seed.go` file:

```go
package main

import (
    "time"
    "gorm.io/gorm"
)

func seedData(db *gorm.DB) {
    // Check if data already exists
    var count int64
    db.Model(&User{}).Count(&count)
    if count > 0 {
        return
    }

    // Create sample users
    users := []User{
        {
            ID:        "1",
            Name:      "John Doe",
            Email:     "john@example.com",
            Role:      "admin",
            CreatedAt: time.Now(),
        },
        {
            ID:        "2",
            Name:      "Jane Smith",
            Email:     "jane@example.com",
            Role:      "user",
            CreatedAt: time.Now(),
        },
    }

    for _, user := range users {
        db.Create(&user)
    }

    // Create sample posts
    posts := []Post{
        {
            ID:        "1",
            Title:     "Welcome to Refine-Gin",
            Content:   "This is our first post using Refine-Gin framework.",
            UserID:    "1",
            CreatedAt: time.Now(),
        },
        {
            ID:        "2",
            Title:     "Getting Started Guide",
            Content:   "Learn how to build APIs with Refine-Gin.",
            UserID:    "1",
            CreatedAt: time.Now(),
        },
    }

    for _, post := range posts {
        db.Create(&post)
    }
}
```

Add to your `main.go`:

```go
func main() {
    // ... existing code ...
    
    // Seed sample data
    seedData(db)
    
    // ... rest of the code ...
}
```

## Step 10: Frontend Integration

### With Refine.dev

If you're using Refine.dev frontend framework:

```typescript
import { Refine } from "@refinedev/core";
import { dataProvider } from "@refinedev/simple-rest";

const App = () => {
  return (
    <Refine
      dataProvider={dataProvider("http://localhost:8080/api")}
      resources={[
        {
          name: "users",
          list: "/users",
          create: "/users",
          edit: "/users/:id",
          show: "/users/:id",
        },
        {
          name: "posts",
          list: "/posts",
          create: "/posts",
          edit: "/posts/:id",
          show: "/posts/:id",
        },
      ]}
    >
      {/* Your components */}
    </Refine>
  );
};
```

### With React Query

```typescript
import { useQuery, useMutation } from '@tanstack/react-query';

// Fetch users
const { data: users } = useQuery({
  queryKey: ['users'],
  queryFn: () => fetch('http://localhost:8080/api/users').then(res => res.json())
});

// Create user
const createUser = useMutation({
  mutationFn: (userData) => 
    fetch('http://localhost:8080/api/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(userData)
    }).then(res => res.json())
});
```

## Next Steps

1. **Add Authentication**: Implement JWT authentication
2. **Add Validation**: Custom field validation rules
3. **Add Relations**: Connect users and posts
4. **Add Custom Handlers**: Implement business logic
5. **Add Middleware**: Caching, logging, etc.
6. **Add Tests**: Unit and integration tests

## Project Structure

```
my-refine-api/
├── main.go           # Main application
├── models.go         # Data models
├── database.go       # Database setup
├── resources.go      # Resource definitions
├── repositories.go   # Repository setup
├── auth.go          # Authentication setup
├── seed.go          # Sample data
├── go.mod           # Go modules
└── go.sum           # Dependencies
```

## Troubleshooting

### Common Issues

1. **Database Connection Error**
   - Check your database connection string
   - Ensure database server is running
   - Verify database credentials

2. **Migration Errors**
   - Check model definitions
   - Ensure all required fields are properly tagged

3. **CORS Issues**
   - Verify CORS middleware is properly configured
   - Check frontend origin settings

4. **Permission Errors**
   - Check resource permissions configuration
   - Verify JWT token and claims

### Debug Mode

Enable debug mode for more detailed logs:

```go
gin.SetMode(gin.DebugMode)
```

### Logging

Add structured logging:

```go
import "log"

log.SetFlags(log.LstdFlags | log.Lshortfile)
```

## Support

- **Documentation**: [API Documentation](API_DOCUMENTATION.md)
- **Examples**: Check the `examples/` directory
- **Issues**: Report bugs on GitHub
- **Discussions**: Join community discussions

Congratulations! You've successfully created your first Refine-Gin API. You now have a fully functional REST API with automatic CRUD operations, metadata generation, and Swagger documentation.