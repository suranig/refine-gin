# Refine-Gin

[![Go Tests](https://github.com/suranig/refine-gin/actions/workflows/ci.yml/badge.svg)](https://github.com/suranig/refine-gin/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/suranig/refine-gin/branch/main/graph/badge.svg)](https://codecov.io/gh/suranig/refine-gin)
[![Go Report Card](https://goreportcard.com/badge/github.com/suranig/refine-gin)](https://goreportcard.com/report/github.com/suranig/refine-gin)
[![GoDoc](https://godoc.org/github.com/suranig/refine-gin?status.svg)](https://godoc.org/github.com/suranig/refine-gin)

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
- Flexible JSON naming convention control (snake_case, camelCase, PascalCase)
- Count endpoint for resources
- Custom endpoints and actions
- Swagger for resources

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

### Relational Actions

Refine-Gin provides built-in support for relational actions that allow connecting and disconnecting related resources:

```go
// Register a resource with relational actions
relationNames := []string{"posts", "profile"}
handler.RegisterResourceForRefineWithRelations(api, userResource, userRepository, "id", relationNames)
```

This will automatically generate the following endpoints for each relation:

- **Attach**: `POST /api/users/:id/actions/attach-{relation}` - Connect related resources
  ```json
  {
    "ids": ["1", "2", "3"]
  }
  ```

- **Detach**: `POST /api/users/:id/actions/detach-{relation}` - Disconnect related resources
  ```json
  {
    "ids": ["1", "2", "3"]
  }
  ```

- **List**: `GET /api/users/:id/actions/list-{relation}` - List related resources

These actions work with all relationship types (one-to-one, one-to-many, many-to-one, many-to-many).

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

## Swagger Documentation

Refine-Gin automatically generates OpenAPI 3.0 documentation for your API, making it easy to understand and test your endpoints.

### Setting Up Swagger

```go
// Create a router group for the API
api := r.Group("/api")

// Register your resources
handler.RegisterResource(api, userResource, userRepository)
handler.RegisterResource(api, postResource, postRepository)

// Configure Swagger info
swaggerInfo := swagger.SwaggerInfo{
    Title:       "My Refine API",
    Description: "API for my application using Refine-Gin",
    Version:     "1.0.0",
    BasePath:    "/api",
}

// Register Swagger routes (after registering all resources)
swagger.RegisterSwagger(r.Group(""), []resource.Resource{userResource, postResource}, swaggerInfo)
```

This will create two endpoints:
- `/swagger` - Swagger UI interface for interactive API documentation
- `/swagger.json` - OpenAPI specification in JSON format

The Swagger documentation includes all endpoints, including bulk operations and relational actions, with proper request/response schemas.

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

### Advanced Filtering

Refine-Gin supports advanced filtering capabilities compatible with Refine.dev:

#### Filter Operators

The following filter operators are supported:

- `eq` - Equal to
- `ne` - Not equal to
- `lt` - Less than
- `gt` - Greater than
- `lte` - Less than or equal to
- `gte` - Greater than or equal to
- `contains` - Contains substring (case-sensitive)
- `containsi` - Contains substring (case-insensitive)
- `startswith` - Starts with
- `endswith` - Ends with
- `null` - Is null (when value is true) or is not null (when value is false)
- `in` - In a list of values

#### Refine.dev Filter Formats

Refine-Gin supports the following Refine.dev filter formats:

1. Format 1: `filter[field][operator]=value`
```
GET /api/users?filter[age][gt]=30&filter[name][contains]=John
```

2. Format 2: `filters[field]=value&operators[field]=operator`
```
GET /api/users?filters[age]=30&operators[age]=gt&filters[name]=John&operators[name]=contains
```

#### Multi-field Sorting

Refine-Gin supports sorting by multiple fields:

```
GET /api/users?sort=age,name&order=desc,asc
```

This sorts users by age in descending order, then by name in ascending order.

### Bulk Operations

Refine-Gin supports bulk operations compatible with Refine.dev standards:

#### Bulk Create

Create multiple resources at once:

```
POST /api/users/batch
{
  "values": [
    { "name": "User 1", "email": "user1@example.com" },
    { "name": "User 2", "email": "user2@example.com" }
  ]
}
```

#### Bulk Update

Update multiple resources with the same values:

```
PUT /api/users/batch
{
  "ids": ["1", "2", "3"],
  "values": {
    "status": "active"
  }
}
```

#### Bulk Delete

Delete multiple resources at once:

```
DELETE /api/users/batch
{
  "ids": ["1", "2", "3"]
}
```

All bulk operations are implemented as atomic transactions, ensuring data integrity.

### Naming Conventions

Refine-Gin supports different naming conventions for JSON fields in requests and responses:

```go
// Configure resource with snake_case naming (default)
opts := resource.DefaultOptions().WithNamingConvention(naming.SnakeCase)
handler.RegisterResourceWithOptions(api, userResource, userRepo, opts)

// Configure resource with camelCase naming
optsCamel := resource.DefaultOptions().WithNamingConvention(naming.CamelCase)
handler.RegisterResourceWithOptions(api, userResource, userRepo, optsCamel)

// Configure resource with PascalCase naming
optsPascal := resource.DefaultOptions().WithNamingConvention(naming.PascalCase)
handler.RegisterResourceWithOptions(api, userResource, userRepo, optsPascal)
```

You can also apply the naming convention middleware directly to any router group:

```go
api := r.Group("/api", middleware.NamingConventionMiddleware(naming.SnakeCase))
```

### Count Endpoint

Refine-Gin automatically generates a count endpoint for each resource, which returns the total number of records for the given filters:

```
GET /api/users/count?status=active
```

Response:
```json
{
  "count": 42
}
```

To enable the count endpoint, include the `OperationCount` operation in your resource definition:

```go
userResource := resource.NewResource(resource.ResourceConfig{
    Name: "users",
    Model: User{},
    Operations: []resource.Operation{
        resource.OperationList, 
        resource.OperationCreate, 
        resource.OperationRead, 
        resource.OperationUpdate, 
        resource.OperationDelete,
        resource.OperationCount, // Enable count endpoint
    },
})
```

### Caching and ETags

Refine-Gin supports HTTP caching via ETags to improve performance and reduce bandwidth usage. The implementation automatically generates ETags based on resource content and handles conditional requests:

```go
// Setting up a resource with ETag support
opts := resource.DefaultOptions().WithETagSupport(true)
handler.RegisterResourceWithOptions(api, userResource, userRepo, opts)
```

When a client makes a request, Refine-Gin will:

1. Generate an ETag based on the resource content.
2. Include the ETag in the response headers.
3. Handle conditional requests with `If-None-Match` headers.
4. Return 304 Not Modified when appropriate, reducing bandwidth.

Example response headers with ETag:
```
ETag: "a1b2c3d4e5f6"
Cache-Control: private, max-age=86400
```

Subsequent client requests can include the ETag to check for modifications:
```
GET /api/users/123
If-None-Match: "a1b2c3d4e5f6"
```

If the resource hasn't changed, the server will respond with:
```
HTTP/1.1 304 Not Modified
```

ETag generation and cache control settings can be customized through the options interface.

## License

MIT

## Contributing

Please see [CONTRIBUTORS.md](CONTRIBUTORS.md) for details on how to contribute to this project.

We welcome contributions from the community!

## Custom Swagger Endpoints

Starting with version 0.3.1, Refine-Gin supports registering custom Swagger endpoints. You can define a custom endpoint using the new `RegisterCustomEndpoint` function available in the `swagger` package. For example:

```go
swagger.RegisterCustomEndpoint(swagger.CustomEndpoint{
    Method: "post",
    Path: "/auth/custom",
    Operation: swagger.Operation{
        Tags:        []string{"auth"},
        Summary:     "Custom auth endpoint",
        Description: "Endpoint for custom auth functionality",
        // ... additional configuration such as RequestBody, Responses, etc. ...
    },
})
```

After registering, the custom endpoints will be merged into the generated OpenAPI documentation when calling `GenerateOpenAPI()`. This allows you to extend the Swagger documentation with endpoints that do not follow the standard resource pattern.