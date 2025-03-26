package resource

import (
	"fmt"
	"time"
)

// This example demonstrates how to create a resource with field lists
func Example_fieldLists() {
	// Define a model
	type User struct {
		ID        uint      `refine:"label=User ID;width=80;fixed=left"`
		Name      string    `refine:"label=Full Name;required;min=3;max=50;searchable"`
		Email     string    `refine:"label=Email Address;required;pattern=^[\\w-\\.]+@([\\w-]+\\.)+[\\w-]{2,4}$;searchable"`
		Role      string    `refine:"label=User Role;required"`
		Active    bool      `refine:"label=Is Active"`
		CreatedAt time.Time `refine:"label=Created At;width=150"`
	}

	// Create resource configuration with field lists
	config := ResourceConfig{
		Name:  "users",
		Label: "Users",
		Icon:  "user",
		Model: &User{},
		Operations: []Operation{
			OperationList,
			OperationCreate,
			OperationUpdate,
			OperationDelete,
		},
		// Define which fields can be filtered
		FilterableFields: []string{
			"name",
			"email",
			"role",
			"active",
		},
		// Define which fields can be searched
		SearchableFields: []string{
			"name",
			"email",
		},
		// Define which fields can be sorted
		SortableFields: []string{
			"id",
			"name",
			"email",
			"created_at",
		},
		// Define which fields appear in the table view
		TableFields: []string{
			"id",
			"name",
			"email",
			"role",
			"active",
			"created_at",
		},
		// Define which fields appear in forms
		FormFields: []string{
			"name",
			"email",
			"role",
			"active",
		},
		// Define which fields are required
		RequiredFields: []string{
			"name",
			"email",
			"role",
		},
	}

	// Create the resource
	res := NewResource(config)

	// Use the resource
	fmt.Printf("Resource: %s\n", res.GetName())
	fmt.Printf("Label: %s\n", res.GetLabel())
	fmt.Printf("Icon: %s\n", res.GetIcon())

	// Get searchable fields
	searchable := res.GetSearchable()
	fmt.Printf("Searchable fields: %v\n", searchable)

	// Get a specific field configuration
	if field := res.GetField("Email"); field != nil {
		fmt.Printf("Email field label: %s\n", field.Label)
		if field.Validation != nil {
			fmt.Printf("Email validation pattern: %s\n", field.Validation.Pattern)
		}
	}

	// Output:
	// Resource: users
	// Label: Users
	// Icon: user
	// Searchable fields: [name email]
	// Email field label: Email Address
	// Email validation pattern: ^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$
}

// This example demonstrates how to create a resource with default field lists
func Example_defaultFieldLists() {
	// Define a model
	type User struct {
		ID        uint      `refine:"label=User ID"`
		Name      string    `refine:"label=Full Name;required"`
		Email     string    `refine:"label=Email Address;required"`
		CreatedAt time.Time `refine:"label=Created At"`
	}

	// Create minimal resource configuration
	config := ResourceConfig{
		Name:  "users",
		Model: &User{},
	}

	// Create the resource - field lists will be generated automatically
	res := NewResource(config)

	// Print the default field lists
	defaultRes := res.(*DefaultResource)
	fmt.Printf("Filterable fields: %v\n", defaultRes.FilterableFields)
	fmt.Printf("Sortable fields: %v\n", defaultRes.SortableFields)
	fmt.Printf("Form fields: %v\n", defaultRes.FormFields)
	fmt.Printf("Required fields: %v\n", defaultRes.RequiredFields)

	// Output:
	// Filterable fields: [ID Name Email CreatedAt]
	// Sortable fields: [ID Name Email CreatedAt]
	// Form fields: [Name Email CreatedAt]
	// Required fields: [Name Email]
}
