// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resource

import (
	"context"
	"errors"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ErrResourceNotFound is returned when a requested resource cannot be found.
var ErrResourceNotFound = errors.New("resource not found")

// Resource defines the interface for a resource that can be managed by the Manager.
//
// Each implementation of a resource is responsible for providing its metadata
// and handling read and subscribe operations.
type Resource interface {
	// Resource returns the MCP representation of the resource, which includes its metadata.
	//
	// Returns:
	//   - *mcp.Resource: The MCP resource definition.
	Resource() *mcp.Resource

	// Service returns the ID of the service that provides this resource.
	//
	// Returns:
	//   - string: The service ID.
	Service() string

	// Read retrieves the content of the resource.
	//
	// Parameters:
	//   - ctx: The context for the request.
	//
	// Returns:
	//   - *mcp.ReadResourceResult: The content of the resource.
	//   - error: An error if reading the resource fails.
	Read(ctx context.Context) (*mcp.ReadResourceResult, error)

	// Subscribe establishes a subscription to the resource, allowing for receiving updates.
	//
	// Parameters:
	//   - ctx: The context for the subscription.
	//
	// Returns:
	//   - error: An error if the subscription fails.
	Subscribe(ctx context.Context) error
}

// ManagerInterface defines the contract for managing resource lifecycle.
type ManagerInterface interface {
	// GetResource retrieves a resource by URI.
	//
	// Parameters:
	//   - uri: string. The URI of the resource.
	//
	// Returns:
	//   - Resource: The resource instance.
	//   - bool: True if found, false otherwise.
	GetResource(uri string) (Resource, bool)

	// AddResource registers a new resource.
	//
	// Parameters:
	//   - resource: Resource. The resource instance to register.
	AddResource(resource Resource)

	// RemoveResource unregisters a resource by URI.
	//
	// Parameters:
	//   - uri: string. The URI of the resource.
	RemoveResource(uri string)

	// ListResources lists all registered resources.
	//
	// Returns:
	//   - []Resource: A list of all registered resources.
	ListResources() []Resource

	// OnListChanged registers a callback for resource list modifications.
	//
	// Parameters:
	//   - f: func(). The callback function.
	OnListChanged(f func())

	// ClearResourcesForService removes all resources for a service.
	//
	// Parameters:
	//   - serviceID: string. The identifier of the service.
	ClearResourcesForService(serviceID string)
}

// Manager manages resources in a thread-safe manner.
//
// It uses a map to store resources and protects concurrent access with a mutex.
type Manager struct {
	mu                sync.RWMutex
	resources         map[string]Resource
	onListChangedFunc func()
	cachedResources   []Resource
}

// NewManager initializes a new Resource Manager.
//
// Parameters:
//   None.
//
// Returns:
//   - *Manager: The initialized manager instance.
func NewManager() *Manager {
	return &Manager{
		resources: make(map[string]Resource),
	}
}

// GetResource retrieves a resource from the manager by its URI.
//
// Parameters:
//   - uri: The URI of the resource to retrieve.
//
// Returns:
//   - Resource: The resource instance if found.
//   - bool: True if the resource exists, false otherwise.
func (rm *Manager) GetResource(uri string) (Resource, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	resource, ok := rm.resources[uri]
	return resource, ok
}

// AddResource adds a new resource to the manager.
//
// If a resource with the same URI already exists, it will be overwritten.
// After adding the resource, it triggers the OnListChanged callback if one is registered.
//
// Parameters:
//   - resource: The resource to be added.
func (rm *Manager) AddResource(resource Resource) {
	var callback func()
	rm.mu.Lock()
	rm.resources[resource.Resource().URI] = resource
	rm.cachedResources = nil
	callback = rm.onListChangedFunc
	rm.mu.Unlock()

	if callback != nil {
		callback()
	}
}

// RemoveResource removes a resource from the manager by its URI.
//
// If the resource exists, it is removed, and the OnListChanged callback is triggered if one is registered.
//
// Parameters:
//   - uri: The URI of the resource to be removed.
func (rm *Manager) RemoveResource(uri string) {
	var callback func()
	rm.mu.Lock()
	if _, ok := rm.resources[uri]; ok {
		delete(rm.resources, uri)
		rm.cachedResources = nil
		callback = rm.onListChangedFunc
	}
	rm.mu.Unlock()

	if callback != nil {
		callback()
	}
}

// ListResources returns a slice containing all the resources currently registered in the manager.
//
// Returns:
//   - []Resource: A slice of all registered resources.
func (rm *Manager) ListResources() []Resource {
	// ⚡ Bolt: Use a read-through cache to avoid repeated map iteration and slice allocation.
	// The cache is invalidated on any write operation (Add/Remove).
	// We use double-checked locking to safely upgrade from RLock to Lock.
	rm.mu.RLock()
	if rm.cachedResources != nil {
		// Return a copy to ensure thread safety
		result := make([]Resource, len(rm.cachedResources))
		copy(result, rm.cachedResources)
		rm.mu.RUnlock()
		return result
	}
	rm.mu.RUnlock()

	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Double-check after acquiring the write lock
	if rm.cachedResources != nil {
		// Return a copy to ensure thread safety
		result := make([]Resource, len(rm.cachedResources))
		copy(result, rm.cachedResources)
		return result
	}

	resources := make([]Resource, 0, len(rm.resources))
	for _, resource := range rm.resources {
		resources = append(resources, resource)
	}
	rm.cachedResources = resources

	// Return a copy to ensure thread safety
	result := make([]Resource, len(resources))
	copy(result, resources)
	return result
}

// OnListChanged sets a callback function that will be invoked whenever the list
// of resources is modified by adding or removing a resource.
//
// Parameters:
//   - f: The callback function to be set.
func (rm *Manager) OnListChanged(f func()) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.onListChangedFunc = f
}

// Subscribe finds a resource by its URI and calls its Subscribe method.
//
// Parameters:
//   - ctx: The context for the subscription.
//   - uri: The URI of the resource to subscribe to.
//
// Returns:
//   - error: An error if the resource is not found or if the subscription fails.
func (rm *Manager) Subscribe(ctx context.Context, uri string) error {
	resource, ok := rm.GetResource(uri)
	if !ok {
		return ErrResourceNotFound
	}
	return resource.Subscribe(ctx)
}

// ClearResourcesForService removes all resources associated with a given service ID.
//
// Parameters:
//   - serviceID: The ID of the service whose resources should be cleared.
func (rm *Manager) ClearResourcesForService(serviceID string) {
	var callback func()
	rm.mu.Lock()
	changed := false
	for uri, resource := range rm.resources {
		if resource.Service() == serviceID {
			delete(rm.resources, uri)
			changed = true
		}
	}
	if changed {
		rm.cachedResources = nil
		callback = rm.onListChangedFunc
	}
	rm.mu.Unlock()

	if callback != nil {
		callback()
	}
}
