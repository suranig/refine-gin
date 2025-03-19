# Generic Repository

Generic repository is an implementation of the repository pattern that provides a unified API for working with different data models. In this project, `GenericRepository` extends the capabilities of the existing `GormRepository` with additional functionalities.

## Key Features

1. **Basic CRUD Operations**
   - Create - creating new records
   - Read - retrieving single records or lists
   - Update - updating existing records
   - Delete - removing records

2. **Advanced Queries**
   - Filtering
   - Sorting
   - Pagination
   - Searching

3. **Relationship Handling**
   - Automatic relationship loading
   - Selection of specific relationships to load

4. **Transactions**
   - Performing multiple operations within a single transaction
   - Automatic rollback in case of errors

5. **Bulk Operations**
   - Bulk creation of records
   - Bulk update of multiple records simultaneously

## Repository Initialization

The repository can be initialized in two ways:

### 1. Directly, by creating a GenericRepository instance

```go
// Create a resource
productResource := resource.NewResource(resource.ResourceConfig{
    Name:  "products",
    Model: Product{},
})

// Initialize repository directly
repo := repository.NewGenericRepository(db, productResource)
```

### 2. Using the GenericRepositoryFactory

```go
// Create a repository factory
repoFactory := repository.NewGenericRepositoryFactory(db)

// Create a repository for the resource
productRepo := repoFactory.CreateRepository(productResource)

// Convert to GenericRepository to access additional methods
genericRepo := productRepo.(*repository.GenericRepository)
```

## Usage Examples

### Basic CRUD Operations

```go
// Get a single record
product, err := repo.Get(ctx, "product-id")

// List with filtering and pagination
options := query.QueryOptions{
    Resource: productResource,
    Page:     1,
    PerPage:  10,
    Sort:     "name",
    Order:    "asc",
    Filters:  map[string]interface{}{"category_id": "1"},
}
products, total, err := repo.List(ctx, options)

// Create a new record
newProduct := &Product{Name: "New product"}
createdProduct, err := repo.Create(ctx, newProduct)

// Update a record
updatedProduct, err := repo.Update(ctx, "product-id", &Product{
    ID:   "product-id",
    Name: "Updated product",
})

// Delete a record
err := repo.Delete(ctx, "product-id")
```

### Operations with Relationships

```go
// Get a product with its category relationship
product, err := genericRepo.GetWithRelations(ctx, "product-id", []string{"Category"})

// Get a list of products with relationships
products, total, err := genericRepo.ListWithRelations(ctx, options, []string{"Category", "Tags"})
```

### Transactions

```go
err := genericRepo.WithTransaction(func(txRepo repository.Repository) error {
    // Perform multiple operations within a single transaction
    _, err := txRepo.Create(ctx, product1)
    if err != nil {
        return err
    }
    
    _, err = txRepo.Create(ctx, product2)
    if err != nil {
        return err
    }
    
    return nil
})
```

### Bulk Operations

```go
// Bulk creation of multiple records
products := []Product{
    {Name: "Product 1", Price: 100},
    {Name: "Product 2", Price: 200},
}
err := genericRepo.BulkCreate(ctx, &products)

// Bulk update of multiple records
err := genericRepo.BulkUpdate(ctx, 
    map[string]interface{}{"category_id": "1"},  // Condition
    map[string]interface{}{"price": 99.99}       // Updates
)
```

### Custom Searches

```go
// Find one record matching a condition
product, err := genericRepo.FindOneBy(ctx, map[string]interface{}{
    "name": "Smartphone",
})

// Find all records matching a condition
products, err := genericRepo.FindAllBy(ctx, map[string]interface{}{
    "category_id": "electronics",
    "price":       100.0,
})
```

## Integration with Existing API

GenericRepository is fully compatible with existing API handlers. It can be used as a drop-in replacement for a standard repository:

```go
// Register a resource in API using the generic repository
api := r.Group("/api")
handler.RegisterResource(api, productResource, genericRepo)
```

## Benefits of Using Generic Repository

1. **Reduction of Repetitive Code** - no need to implement similar methods for each model
2. **Consistent API** - all models use the same interface
3. **Easy Extensibility** - new functionalities can be added in one place
4. **Transaction Support** - safely perform multiple operations as a single atomic unit
5. **Flexible Queries** - rich set of methods for filtering and searching data 