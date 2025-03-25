package resource

import (
	"fmt"
	"sync"
)

// Registry is a singleton registry for resources
type Registry struct {
	resources map[string]Resource
	mu        sync.RWMutex
}

var (
	registry     *Registry
	registryOnce sync.Once
)

// GetRegistry returns the singleton registry instance
func GetRegistry() *Registry {
	registryOnce.Do(func() {
		registry = &Registry{
			resources: make(map[string]Resource),
		}
	})
	return registry
}

// RegisterResource adds a resource to the registry
func (r *Registry) RegisterResource(res Resource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resources[res.GetName()] = res
}

// GetResource returns a resource by name
func (r *Registry) GetResource(name string) (Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if res, ok := r.resources[name]; ok {
		return res, nil
	}
	return nil, fmt.Errorf("resource '%s' not found in registry", name)
}

// HasResource checks if a resource exists in the registry
func (r *Registry) HasResource(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.resources[name]
	return ok
}

// ResourceNames returns all resource names in the registry
func (r *Registry) ResourceNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.resources))
	for name := range r.resources {
		names = append(names, name)
	}
	return names
}

// RegisterToRegistry is a convenience function to register a resource to the global registry
func RegisterToRegistry(res Resource) {
	GetRegistry().RegisterResource(res)
}
