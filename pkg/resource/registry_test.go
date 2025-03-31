package resource

import (
	"strings"
	"testing"
)

// RegistryMockResource is a mock resource for testing the registry
type RegistryMockResource struct {
	ResourceName string
	ResourceIcon string
	Model        interface{}
}

func (r *RegistryMockResource) GetName() string {
	return r.ResourceName
}

func (r *RegistryMockResource) GetLabel() string {
	return strings.Title(r.ResourceName)
}

func (r *RegistryMockResource) GetIcon() string {
	return r.ResourceIcon
}

func (r *RegistryMockResource) GetModel() interface{} {
	return r.Model
}

func (r *RegistryMockResource) GetFields() []Field {
	return []Field{}
}

func (r *RegistryMockResource) GetOperations() []Operation {
	return []Operation{}
}

func (r *RegistryMockResource) HasOperation(op Operation) bool {
	return false
}

func (r *RegistryMockResource) GetDefaultSort() *Sort {
	return nil
}

func (r *RegistryMockResource) GetFilters() []Filter {
	return []Filter{}
}

func (r *RegistryMockResource) GetMiddlewares() []interface{} {
	return []interface{}{}
}

func (r *RegistryMockResource) GetRelations() []Relation {
	return []Relation{}
}

func (r *RegistryMockResource) HasRelation(name string) bool {
	return false
}

func (r *RegistryMockResource) GetRelation(name string) *Relation {
	return nil
}

func (r *RegistryMockResource) GetIDFieldName() string {
	return "ID"
}

func (r *RegistryMockResource) GetField(name string) *Field {
	return nil
}

func (r *RegistryMockResource) GetSearchable() []string {
	return []string{}
}

func (r *RegistryMockResource) GetFilterableFields() []string {
	return []string{}
}

func (r *RegistryMockResource) GetSortableFields() []string {
	return []string{}
}

func (r *RegistryMockResource) GetTableFields() []string {
	return []string{}
}

func (r *RegistryMockResource) GetFormFields() []string {
	return []string{}
}

func (r *RegistryMockResource) GetRequiredFields() []string {
	return []string{}
}

func (r *RegistryMockResource) GetEditableFields() []string {
	return []string{}
}

// TestResourceRegistry tests the ResourceRegistry and GlobalResourceRegistry
func TestResourceRegistry(t *testing.T) {
	// Reset GlobalResourceRegistry for testing
	GlobalResourceRegistry = NewResourceRegistry()

	// Test registering a new resource
	resource := &RegistryMockResource{ResourceName: "test-resource", ResourceIcon: "test-icon"}
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
		GlobalResourceRegistry.Register(&RegistryMockResource{ResourceName: name, ResourceIcon: "test-icon"})
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
	resource = &RegistryMockResource{ResourceName: "test-resource", ResourceIcon: "test-icon"}
	RegisterToRegistry(resource)

	// Verify the resource was registered
	_, ok = GlobalResourceRegistry.GetByName("test-resource")
	if !ok {
		t.Errorf("RegisterToRegistry() failed to register the resource")
	}
}
