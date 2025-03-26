package resource

import (
	"sync"
)

// ResourceRegistry provides a registry for managing resources
type ResourceRegistry struct {
	resources map[string]Resource
	mutex     sync.RWMutex
}

// NewResourceRegistry creates a new resource registry
func NewResourceRegistry() *ResourceRegistry {
	return &ResourceRegistry{
		resources: make(map[string]Resource),
	}
}

// Register adds a resource to the registry
func (r *ResourceRegistry) Register(res Resource) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.resources[res.GetName()] = res
}

// GetByName retrieves a resource by name
func (r *ResourceRegistry) GetByName(name string) (Resource, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	res, ok := r.resources[name]
	return res, ok
}

// GetAll returns all registered resources
func (r *ResourceRegistry) GetAll() []Resource {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	result := make([]Resource, 0, len(r.resources))
	for _, res := range r.resources {
		result = append(result, res)
	}

	return result
}

// GlobalResourceRegistry is a singleton instance of the resource registry
var GlobalResourceRegistry = NewResourceRegistry()

// RegisterToRegistry registers a resource to the global registry
func RegisterToRegistry(res Resource) {
	GlobalResourceRegistry.Register(res)
}
