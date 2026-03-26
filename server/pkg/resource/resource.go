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
//
// Summary: Represents a ErrResourceNotFound.
var ErrResourceNotFound = errors.New("resource not found")

// Resource resource represents a resource.
//
// Summary: Resource represents a resource.
type Resource interface {
	// Resource returns the MCP representation of the resource, which includes its metadata.
	//
	// Returns: - None.
	//   - *mcp.Resource: The MCP resource definition.
	Resource() *mcp.Resource

	// Service returns the ID of the service that provides this resource.
	//
	// Returns: - None.
	//   - string: The service ID.
	Service() string

	// Read retrieves the content of the resource.
	//
	// Parameters: - None.
	//   - ctx: context.Context. The context for the request.
	//
	// Returns: - None.
	//   - *mcp.ReadResourceResult: The content of the resource.
	//   - error: An error if reading fails.
	Read(ctx context.Context) (*mcp.ReadResourceResult, error)

	// Subscribe establishes a subscription to the resource, allowing for receiving updates.
	//
	// Parameters: - None.
	//   - ctx: context.Context. The context for the subscription.
	//
	// Returns: - None.
	//   - error: An error if subscription fails.
	Subscribe(ctx context.Context) error
}

// ManagerInterface managerInterface represents a manager interface.
//
// Summary: ManagerInterface represents a manager interface.
type ManagerInterface interface {
	// GetResource retrieves a resource by its URI.
	//
	// Parameters: - None.
	//   - uri: string. The URI of the resource.
	//
	// Returns: - None.
	//   - Resource: The resource instance.
	//   - bool: True if found, false otherwise.
	GetResource(uri string) (Resource, bool)

	// AddResource adds a new resource to the manager.
	//
	// Parameters: - None.
	//   - resource: Resource. The resource to add.
	//
	// Returns: - None.
	//   None.
	AddResource(resource Resource)

	// RemoveResource removes a resource from the manager by its URI.
	//
	// Parameters: - None.
	//   - uri: string. The URI of the resource to remove.
	//
	// Returns: - None.
	//   None.
	RemoveResource(uri string)

	// ListResources returns a slice of all resources currently in the manager.
	//
	// Returns: - None.
	//   - []Resource: A slice of resources.
	ListResources() []Resource

	// OnListChanged registers a callback function to be called when the list of resources changes.
	//
	// Parameters: - None.
	//   - f: func(). The callback function to execute on change.
	//
	// Returns: - None.
	//   None.
	OnListChanged(f func())

	// ClearResourcesForService removes all resources associated with a given service ID.
	//
	// Parameters: - None.
	//   - serviceID: string. The service ID.
	//
	// Returns: - None.
	//   None.
	ClearResourcesForService(serviceID string)
}

// Manager manager represents a manager.
//
// Summary: Manager represents a manager.
type Manager struct {
	mu                sync.RWMutex
	resources         map[string]Resource
	onListChangedFunc func()
	cachedResources   []Resource
}

// NewManager creates a new manager.
//
// Summary: Creates a new manager.
//
// Parameters:
//   - None.
//
// Returns:
//   - *Manager: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewManager() *Manager {
	return &Manager{
		resources: make(map[string]Resource),
	}
}

// GetResource retrieves a resource from the manager by its URI.
//
// Summary: Retrieves a resource by URI.
//
// Parameters: - None.
//   - uri: string. The URI of the resource.
//
// Returns: - None.
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
// Summary: Adds a resource to the manager.
//
// Parameters: - None.
//   - resource: Resource. The resource to add.
//
// Returns: - None.
//
//	None.
//
// Side Effects: - None.
//   - Updates the internal resource storage.
//   - Invalidates the list cache.
//   - Triggers the on-change callback if registered.
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
// Summary: Removes a resource from the manager.
//
// Parameters: - None.
//   - uri: string. The URI of the resource.
//
// Returns: - None.
//
//	None.
//
// Side Effects: - None.
//   - Updates the internal resource storage.
//   - Invalidates the list cache.
//   - Triggers the on-change callback if registered.
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

// ListResources retrieves a list of resources.
//
// Summary: Retrieves a list of resources.
//
// Parameters:
//   - None.
//
// Returns:
//   - []Resource: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
// Summary: Registers a callback for list changes.
//
// Parameters: - None.
//   - f: func(). The callback function.
//
// Returns: - None.
//
//	None.
func (rm *Manager) OnListChanged(f func()) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.onListChangedFunc = f
}

// Subscribe finds a resource by its URI and calls its Subscribe method.
//
// Summary: Subscribes to a resource.
//
// Parameters: - None.
//   - ctx: context.Context. The context for the subscription.
//   - uri: string. The URI of the resource.
//
// Returns: - None.
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
// Summary: Clears resources for a specific service.
//
// Parameters: - None.
//   - serviceID: string. The service ID.
//
// Returns: - None.
//
//	None.
//
// Side Effects: - None.
//   - Removes matching resources from storage.
//   - Invalidates the list cache.
//   - Triggers the on-change callback.
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
