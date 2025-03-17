# Custom ID Example

This example demonstrates how to use custom ID fields with refine-gin. It shows how to:

1. Define models with custom ID field names (other than the default "ID")
2. Configure resources to use custom ID field names
3. Register resources with custom URL parameter names for IDs

## Models

The example defines two models with custom ID fields:

- `User` with an ID field named `UID`
- `Product` with an ID field named `GUID`

## Usage

### 1. Define models with custom ID fields

```go
// User is a model with a custom ID field named UID
type User struct {
    UID   string `json:"uid" gorm:"primaryKey"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Product is a model with a custom ID field named GUID
type Product struct {
    GUID        string  `json:"guid" gorm:"primaryKey"`
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Price       float64 `json:"price"`
}
```

### 2. Create resources with custom ID field names

```go
// Create resource with custom ID field
userResource := resource.NewResource(resource.ResourceConfig{
    Name:        "users",
    Model:       User{},
    IDFieldName: "UID", // Specify custom ID field name
    Operations:  []resource.Operation{...},
})
```

### 3. Create repository using factory

```go
// Create repository factory
repoFactory := repository.NewGormRepositoryFactory(db)

// Create repository for the resource
userRepo := repoFactory.CreateRepository(userResource)
```

### 4. Register resource with custom URL parameter name

```go
// Register resource with custom ID parameter
handler.RegisterResourceWithOptions(api, userResource, userRepo, handler.RegisterOptions{
    IDParamName: "uid", // Specify custom URL parameter name
})
```

## Endpoints

With the above configuration, the following endpoints will be available:

- `GET /api/users` - List all users
- `GET /api/users/:uid` - Get a user by UID
- `POST /api/users` - Create a new user
- `PUT /api/users/:uid` - Update a user by UID
- `DELETE /api/users/:uid` - Delete a user by UID

- `GET /api/products` - List all products
- `GET /api/products/:guid` - Get a product by GUID
- `POST /api/products` - Create a new product
- `PUT /api/products/:guid` - Update a product by GUID
- `DELETE /api/products/:guid` - Delete a product by GUID

## Running the Example

```bash
go run main.go
```

The server will start on port 8080.

## Testing

Run the integration test:

```bash
go test -v
``` 