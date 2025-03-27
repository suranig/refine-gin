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

func (m *RegistryMockResource) GetFilterableFields() []string {
	return nil
}

func (m *RegistryMockResource) GetSortableFields() []string {
	return nil
}

func (m *RegistryMockResource) GetRequiredFields() []string {
	return nil
}

func (m *RegistryMockResource) GetTableFields() []string {
	return nil
}

func (m *RegistryMockResource) GetFormFields() []string {
	return nil
}

// TestResourceRegistry tests the ResourceRegistry and GlobalResourceRegistry
func TestResourceRegistry(t *testing.T) {
	// Reset GlobalResourceRegistry for testing
	GlobalResourceRegistry = NewResourceRegistry()

	// Test registering a new resource
	resource := &RegistryMockResource{name: "test-resource"}
	GlobalResourceRegistry.Register(resource)

	// Verify the resource was registered
	res, ok := GlobalResourceRegistry.GetByName("test-resource")
	if !ok {
		t.Errorf("Register() failed to register the resource")
	}
	if res.GetName() != resource.GetName() {
		t.Errorf("GetByName() returned a resource with incorrect name: got %s, want %s",
			res.GetName(), resource.GetName())
	}

	// Test retrieving a non-existent resource
	_, ok = GlobalResourceRegistry.GetByName("non-existent")
	if ok {
		t.Errorf("GetByName() should return false for a non-existent resource")
	}

	// Test GetAll
	GlobalResourceRegistry = NewResourceRegistry()

	// Register test resources
	resources := []string{"resource1", "resource2", "resource3"}
	for _, name := range resources {
		GlobalResourceRegistry.Register(&RegistryMockResource{name: name})
	}

	// Test GetAll returns all registered resources
	allResources := GlobalResourceRegistry.GetAll()
	if len(allResources) != len(resources) {
		t.Errorf("GetAll() returned %d resources, expected %d", len(allResources), len(resources))
	}

	// Check that all registered names are in the result
	for _, name := range resources {
		found := false
		for _, res := range allResources {
			if res.GetName() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetAll() did not return the expected resource: %s", name)
		}
	}

	// Test RegisterToRegistry function
	GlobalResourceRegistry = NewResourceRegistry()

	// Test using the convenience function
	resource = &RegistryMockResource{name: "test-resource"}
	RegisterToRegistry(resource)

	// Verify the resource was registered
	_, ok = GlobalResourceRegistry.GetByName("test-resource")
	if !ok {
		t.Errorf("RegisterToRegistry() failed to register the resource")
	}
}
