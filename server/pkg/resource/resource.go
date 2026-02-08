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
// A resource represents a data source (e.g., a file, a database record) that can be
// read by an MCP client.
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
	//   - ctx: context.Context. The context for the request.
	//
	// Returns:
	//   - *mcp.ReadResourceResult: The content of the resource.
	//   - error: An error if reading fails.
	Read(ctx context.Context) (*mcp.ReadResourceResult, error)

	// Subscribe establishes a subscription to the resource, allowing for receiving updates.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the subscription.
	//
	// Returns:
	//   - error: An error if subscription fails.
	Subscribe(ctx context.Context) error
}

// ManagerInterface defines the interface for managing a collection of resources.
//
// It provides methods for adding, removing, listing, and retrieving resources, as well
// as managing callbacks for list changes.
type ManagerInterface interface {
	// GetResource retrieves a resource by its URI.
	//
	// Parameters:
	//   - uri: string. The URI of the resource.
	//
	// Returns:
	//   - Resource: The resource instance.
	//   - bool: True if found, false otherwise.
	GetResource(uri string) (Resource, bool)

	// AddResource adds a new resource to the manager.
	//
	// Parameters:
	//   - resource: Resource. The resource to add.
	AddResource(resource Resource)

	// RemoveResource removes a resource from the manager by its URI.
	//
	// Parameters:
	//   - uri: string. The URI of the resource to remove.
	RemoveResource(uri string)

	// ListResources returns a slice of all resources currently in the manager.
	//
	// Returns:
	//   - []Resource: A slice of resources.
	ListResources() []Resource

	// OnListChanged registers a callback function to be called when the list of resources changes.
	//
	// Parameters:
	//   - f: func(). The callback function to execute on change.
	OnListChanged(f func())

	// ClearResourcesForService removes all resources associated with a given service ID.
	//
	// Parameters:
	//   - serviceID: string. The service ID.
	ClearResourcesForService(serviceID string)
}

// Manager is a thread-safe implementation of the ManagerInterface.
//
// It manages the lifecycle and retrieval of resources, providing thread-safe access
// and efficient listing via caching.
type Manager struct {
	mu                sync.RWMutex
	resources         map[string]Resource
	onListChangedFunc func()
	cachedResources   []Resource
}

// NewManager creates and returns a new, empty Manager.
//
// Returns:
//   - *Manager: A new Manager instance.
func NewManager() *Manager {
	return &Manager{
		resources: make(map[string]Resource),
	}
}

// GetResource retrieves a resource from the manager by its URI.
//
// Parameters:
//   - uri: string. The URI of the resource.
//
// Returns:
//   - Resource: The resource instance.
//   - bool: True if found, false otherwise.
func (rm *Manager) GetResource(uri string) (Resource, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	resource, ok := rm.resources[uri]
	return resource, ok
}

// AddResource adds a new resource to the manager.
//
// It updates the internal storage, invalidates the list cache, and triggers any registered
// change callbacks.
//
// Parameters:
//   - resource: Resource. The resource to add.
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
// It updates the internal storage, invalidates the list cache, and triggers any registered
// change callbacks.
//
// Parameters:
//   - uri: string. The URI of the resource.
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
// It uses a read-through cache (double-checked locking) to minimize allocation overhead
// for frequent calls.
//
// Returns:
//   - []Resource: A slice of currently registered resources.
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
// of resources is modified.
//
// Parameters:
//   - f: func(). The callback function.
func (rm *Manager) OnListChanged(f func()) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.onListChangedFunc = f
}

// Subscribe finds a resource by its URI and calls its Subscribe method.
//
// Parameters:
//   - ctx: context.Context. The context for the subscription.
//   - uri: string. The URI of the resource.
//
// Returns:
//   - error: An error if resource not found or subscription fails.
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
//   - serviceID: string. The service ID.
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
