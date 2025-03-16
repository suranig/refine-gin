# Refine-Gin

Refine-Gin is a library that integrates the Gin framework with Refine.js, enabling rapid development of RESTful APIs compatible with Refine.js conventions.

## Dependencies

This library integrates the following technologies:

- [Gin](https://github.com/gin-gonic/gin) - A high-performance HTTP web framework written in Go
- [Refine](https://refine.dev/) - A React-based framework for building data-intensive applications
- [JWT-Go](https://github.com/golang-jwt/jwt) - A Go implementation of JSON Web Tokens
- [GORM](https://gorm.io/) - The fantastic ORM library for Go

## Features

- Automatic generation of REST endpoints based on resource definitions
- Full compatibility with Refine.js conventions (filters, sorting, pagination)
- Input data validation
- Data transformation through DTO layer
- Support for relationships between resources
- JWT authentication and authorization
- Customizable validation rules

## Installation

```bash
go get github.com/suranig/refine-gin
```

## Quick Start

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/suranig/refine-gin/pkg/handler"
    "github.com/suranig/refine-gin/pkg/resource"
    "gorm.io/gorm"
)

// Model definition
type User struct {
    ID        string    `json:"id" gorm:"primaryKey" refine:"filterable;sortable;searchable"`
    Name      string    `json:"name" refine:"filterable;sortable"`
    Email     string    `json:"email" refine:"filterable"`
    CreatedAt time.Time `json:"created_at" refine:"filterable;sortable"`
}

// Repository implementation
type UserRepository struct {
    db *gorm.DB
}

// Implement repository methods...

func main() {
    r := gin.Default()
    
    // Resource definition
    userResource := resource.NewResource(resource.ResourceConfig{
        Name: "users",
        Model: User{},
        Operations: []resource.Operation{
            resource.OperationList, 
            resource.OperationCreate, 
            resource.OperationRead, 
            resource.OperationUpdate, 
            resource.OperationDelete,
        },
    })
    
    // Register resource
    api := r.Group("/api")
    handler.RegisterResource(api, userResource, userRepository)
    
    r.Run(":8080")
}
```

## API Documentation

### Resource Definition

Resources are defined using the `ResourceConfig` structure:

```go
userResource := resource.NewResource(resource.ResourceConfig{
    Name: "users",
    Model: User{},
    Fields: []resource.Field{
        {Name: "id", Type: "string", Filterable: true},
        {Name: "name", Type: "string", Filterable: true, Searchable: true},
        {Name: "email", Type: "string", Filterable: true},
        {Name: "created_at", Type: "time.Time", Filterable: true, Sortable: true},
    },
    Operations: []resource.Operation{
        resource.OperationList, 
        resource.OperationCreate, 
        resource.OperationRead, 
        resource.OperationUpdate, 
        resource.OperationDelete,
    },
    DefaultSort: &resource.Sort{
        Field: "created_at",
        Order: "desc",
    },
})
```

Alternatively, you can use `refine` tags in your model definition:

```go
type User struct {
    ID        string    `json:"id" gorm:"primaryKey" refine:"filterable;sortable;searchable"`
    Name      string    `json:"name" refine:"filterable;sortable"`
    Email     string    `json:"email" refine:"filterable"`
    CreatedAt time.Time `json:"created_at" refine:"filterable;sortable"`
}
```

### Relationships

You can define relationships between resources using the `relation` tag:

```go
type User struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name"`
    Posts     []Post    `json:"posts" gorm:"foreignKey:AuthorID" relation:"resource=posts;type=one-to-many;field=author_id;reference=id;include=false"`
    Profile   *Profile  `json:"profile" gorm:"foreignKey:UserID" relation:"resource=profiles;type=one-to-one;field=user_id;reference=id;include=true"`
}
```

Supported relationship types:
- `one-to-one`
- `one-to-many`
- `many-to-one`
- `many-to-many`

### Resource Registration

Resources are registered with the Gin router:

```go
api := r.Group("/api")
handler.RegisterResource(api, userResource, userRepository)
```

For advanced data transformation, you can use DTOs:

```go
dtoProvider := &dto.DefaultDTOProvider{
    Model: &User{},
}
handler.RegisterResourceWithDTO(api, userResource, userRepository, dtoProvider)
```

### Authentication and Authorization

The library provides JWT authentication and authorization:

```go
// JWT configuration
jwtConfig := auth.DefaultJWTConfig()
jwtConfig.Secret = "your-secret-key"

// JWT middleware
r.Use(auth.JWTMiddleware(jwtConfig))

// Authorization provider
authProvider := auth.NewJWTAuthorizationProvider()
authProvider.AddRule("users", resource.OperationList, auth.HasRole("admin"))
authProvider.AddRule("users", resource.OperationDelete, auth.HasAllRoles("admin", "manager"))
authProvider.AddRule("posts", resource.OperationUpdate, auth.IsOwner("sub", "AuthorID"))

// Authorization middleware
r.Use(auth.AuthorizationMiddleware(authProvider))
```

### Query Parameters

The library supports all Refine.js query parameters:

- Filtering: `?name=John&email_operator=contains&email=example.com`
- Sorting: `?sort=created_at&order=desc`
- Pagination: `?page=1&per_page=10`
- Search: `?q=searchterm`
- Including relations: `?include=posts,profile`

## License

MIT

## Contributing

Please see [CONTRIBUTORS.md](CONTRIBUTORS.md) for details on how to contribute to this project.

We welcome contributions from the community!