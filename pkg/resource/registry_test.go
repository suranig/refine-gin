package resource

import (
	"testing"
)

// RegistryMockResource implements the Resource interface for testing
type RegistryMockResource struct {
	name string
}

func (m *RegistryMockResource) GetName() string {
	return m.name
}

func (m *RegistryMockResource) GetLabel() string {
	return m.name
}

func (m *RegistryMockResource) GetIcon() string {
	return "test-icon"
}

func (m *RegistryMockResource) GetModel() interface{} {
	return nil
}

func (m *RegistryMockResource) GetFields() []Field {
	return nil
}

func (m *RegistryMockResource) GetOperations() []Operation {
	return nil
}

func (m *RegistryMockResource) HasOperation(op Operation) bool {
	return false
}

func (m *RegistryMockResource) GetDefaultSort() *Sort {
	return nil
}

func (m *RegistryMockResource) GetFilters() []Filter {
	return nil
}

func (m *RegistryMockResource) GetMiddlewares() []interface{} {
	return nil
}

func (m *RegistryMockResource) GetRelations() []Relation {
	return nil
}

func (m *RegistryMockResource) HasRelation(name string) bool {
	return false
}

func (m *RegistryMockResource) GetRelation(name string) *Relation {
	return nil
}

func (m *RegistryMockResource) GetIDFieldName() string {
	return "ID"
}

func (m *RegistryMockResource) GetField(name string) *Field {
	return nil
}

func (m *RegistryMockResource) GetSearchable() []string {
	return nil
}

// TestGetRegistry tests the GetRegistry function
func TestGetRegistry(t *testing.T) {
	// Test that GetRegistry returns a singleton instance
	r1 := GetRegistry()
	r2 := GetRegistry()

	if r1 != r2 {
		t.Errorf("GetRegistry() returned different instances: %p != %p", r1, r2)
	}

	// Test that the registry is initialized with an empty map
	if r1.resources == nil {
		t.Errorf("Registry.resources is nil, expected an initialized map")
	}
}

// TestRegisterResource tests the RegisterResource method
func TestRegisterResource(t *testing.T) {
	// Create a clean registry for testing
	registry = &Registry{
		resources: make(map[string]Resource),
	}

	// Test registering a new resource
	resource := &RegistryMockResource{name: "test-resource"}
	GetRegistry().RegisterResource(resource)

	// Verify the resource was registered
	if !GetRegistry().HasResource("test-resource") {
		t.Errorf("RegisterResource() failed to register the resource")
	}

	// Test retrieving the registered resource
	res, err := GetRegistry().GetResource("test-resource")
	if err != nil {
		t.Errorf("GetResource() returned an error: %v", err)
	}
	if res.GetName() != resource.GetName() {
		t.Errorf("GetResource() returned a resource with incorrect name: got %s, want %s",
			res.GetName(), resource.GetName())
	}
}

// TestGetResource tests the GetResource method
func TestGetResource(t *testing.T) {
	// Create a clean registry for testing
	registry = &Registry{
		resources: make(map[string]Resource),
	}

	// Register a test resource
	resource := &RegistryMockResource{name: "test-resource"}
	GetRegistry().RegisterResource(resource)

	// Test retrieving an existing resource
	res, err := GetRegistry().GetResource("test-resource")
	if err != nil {
		t.Errorf("GetResource() returned an error for an existing resource: %v", err)
	}
	if res.GetName() != "test-resource" {
		t.Errorf("GetResource() returned a resource with incorrect name: got %s, want %s",
			res.GetName(), "test-resource")
	}

	// Test retrieving a non-existent resource
	_, err = GetRegistry().GetResource("non-existent")
	if err == nil {
		t.Errorf("GetResource() should return an error for a non-existent resource")
	}
}

// TestHasResource tests the HasResource method
func TestHasResource(t *testing.T) {
	// Create a clean registry for testing
	registry = &Registry{
		resources: make(map[string]Resource),
	}

	// Register a test resource
	resource := &RegistryMockResource{name: "test-resource"}
	GetRegistry().RegisterResource(resource)

	// Test HasResource with an existing resource
	if !GetRegistry().HasResource("test-resource") {
		t.Errorf("HasResource() returned false for an existing resource")
	}

	// Test HasResource with a non-existent resource
	if GetRegistry().HasResource("non-existent") {
		t.Errorf("HasResource() returned true for a non-existent resource")
	}
}

// TestResourceNames tests the ResourceNames method
func TestResourceNames(t *testing.T) {
	// Create a clean registry for testing
	registry = &Registry{
		resources: make(map[string]Resource),
	}

	// Register test resources
	resources := []string{"resource1", "resource2", "resource3"}
	for _, name := range resources {
		GetRegistry().RegisterResource(&RegistryMockResource{name: name})
	}

	// Test ResourceNames returns all registered resource names
	names := GetRegistry().ResourceNames()
	if len(names) != len(resources) {
		t.Errorf("ResourceNames() returned %d names, expected %d", len(names), len(resources))
	}

	// Check that all registered names are in the result
	for _, name := range resources {
		found := false
		for _, returnedName := range names {
			if returnedName == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ResourceNames() did not return the expected resource name: %s", name)
		}
	}
}

// TestRegisterToRegistry tests the RegisterToRegistry function
func TestRegisterToRegistry(t *testing.T) {
	// Create a clean registry for testing
	registry = &Registry{
		resources: make(map[string]Resource),
	}

	// Test using the convenience function
	resource := &RegistryMockResource{name: "test-resource"}
	RegisterToRegistry(resource)

	// Verify the resource was registered
	if !GetRegistry().HasResource("test-resource") {
		t.Errorf("RegisterToRegistry() failed to register the resource")
	}
}
