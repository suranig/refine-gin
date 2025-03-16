# Refine-Gin

A seamless integration library between Refine.js and Gin, making it easy to build full-stack applications with both technologies.

## Features

- Automatic REST endpoint generation based on resource definitions
- Full compatibility with Refine.js conventions (filters, sorting, pagination)
- Type safety through TypeScript interface generation
- Automatic data validation and sanitization
- API documentation generation (Swagger)

## Installation

### Backend (Go)

```bash
go get github.com/suranig/refine-gin
```

## Quick Start

### Backend (Go)

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/suranig/refine-gin/pkg/resource"
    "github.com/suranig/refine-gin/pkg/handler"
)

// Define your model
type User struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" refine:"filterable;searchable"`
    Email     string    `json:"email" refine:"filterable"`
    CreatedAt time.Time `json:"created_at" refine:"filterable;sortable"`
}

func main() {
    r := gin.Default()
    
    // Define your resource
    userResource := resource.NewResource(
        resource.ResourceConfig{
            Name: "users",
            Model: User{},
            Operations: []resource.Operation{
                resource.OperationList, 
                resource.OperationCreate, 
                resource.OperationRead, 
                resource.OperationUpdate, 
                resource.OperationDelete,
            },
        },
    )
    
    // Register the resource
    api := r.Group("/api/v1")
    handler.RegisterResource(api, userResource, userRepository)
    
    r.Run(":8080")
}
```

### Frontend (TypeScript)

```typescript
import { Refine } from "@refinedev/core";
import { dataProvider } from "@suranig/refine-gin";

const App = () => {
    return (
        <Refine
            dataProvider={dataProvider("http://localhost:8080/api/v1")}
            resources={[
                {
                    name: "users",
                    list: "/users",
                    create: "/users/create",
                    edit: "/users/edit/:id",
                    show: "/users/show/:id",
                },
            ]}
        />
    );
};
```

## Documentation

For full documentation, visit [refine-gin.suranig.com](https://refine-gin.suranig.com).

## Features

### Resource Definition

Define your API resources with a simple, declarative syntax:

```go
userResource := resource.NewResource(
    resource.ResourceConfig{
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
    },
)
```

### Automatic CRUD Endpoints

The library automatically generates all necessary CRUD endpoints:

- `GET /api/v1/users` - List users with filtering, sorting, and pagination
- `GET /api/v1/users/:id` - Get a single user
- `POST /api/v1/users` - Create a new user
- `PUT /api/v1/users/:id` - Update a user
- `DELETE /api/v1/users/:id` - Delete a user

### Query Parameters

The library supports all Refine.js query parameters:

- Filtering: `?name=John&email_operator=contains&email=example.com`
- Sorting: `?sort=created_at&order=desc`
- Pagination: `?page=1&per_page=10`
- Search: `?q=searchterm`

## License
MIT