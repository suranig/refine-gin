# Refine-Gin Integration

Biblioteka integracyjna między Refine.js a Gin, ułatwiająca tworzenie aplikacji z wykorzystaniem obu technologii.

## Funkcje

- Automatyczne generowanie endpointów REST na podstawie definicji zasobów
- Zgodność z konwencjami Refine.js (filtry, sortowanie, paginacja)
- Typowa bezpieczność dzięki generowaniu interfejsów TypeScript
- Automatyczna walidacja i sanityzacja danych
- Generowanie dokumentacji API (Swagger)

## Instalacja

### Backend (Go)

```bash
go get github.com/yourusername/refine-gin-integration
```

### Frontend (TypeScript)

```bash
npm install refine-gin-integration
```

## Przykład użycia

### Backend (Go)

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/yourusername/refine-gin-integration/pkg/resource"
    "github.com/yourusername/refine-gin-integration/pkg/handler"
)

func main() {
    r := gin.Default()
    
    // Definicja zasobu
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
    
    // Rejestracja zasobu
    api := r.Group("/api/v1")
    handler.RegisterResource(api, userResource, userRepository)
    
    r.Run(":8080")
}
```

### Frontend (TypeScript)

```typescript
import { Refine } from "@refinedev/core";
import { dataProvider } from "refine-gin-integration";

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

## Dokumentacja

Pełna dokumentacja dostępna jest [tutaj](link-do-dokumentacji).

## Licencja

MIT 