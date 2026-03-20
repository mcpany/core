package interop

import "fmt"

// Registry holds the registered framework adapters
type Registry struct {
	adapters map[Framework]Adapter
}

// NewRegistry creates a new Interop Hub registry
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[Framework]Adapter),
	}
}

// Register adds an adapter to the registry
func (r *Registry) Register(adapter Adapter) {
	r.adapters[adapter.GetFramework()] = adapter
}

// Dispatch routes the request from one agent framework to another
func (r *Registry) Dispatch(req *InteropRequest) (*InteropResponse, error) {
	targetAdapter, exists := r.adapters[req.Target]
	if !exists {
		return nil, fmt.Errorf("target framework %s not found in registry", req.Target)
	}

	return targetAdapter.Execute(req)
}
