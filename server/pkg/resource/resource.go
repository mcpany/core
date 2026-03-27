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

// Resource defines the interface for a resource that can be managed by the Manager.
//
// Summary: Interface for a managed resource.
//
// A resource represents a data source (e.g., a file, a database record) that can be
// read by an MCP client.
type Resource interface {
	// Resource returns the MCP representation of the resource, which includes its metadata.
	//
// Returns.
	//   - *mcp.Resource: The MCP resource definition.
	Resource() *mcp.Resource

	// Service returns the ID of the service that provides this resource.
	//
// Returns.
	//   - string: The service ID.
	Service() string

	// Read retrieves the content of the resource.
	//
// Parameters.
	//   - ctx: context.Context. The context for the request.
	//
// Returns.
	//   - *mcp.ReadResourceResult: The content of the resource.
	//   - error: An error if reading fails.
	Read(ctx context.Context) (*mcp.ReadResourceResult, error)

	// Subscribe establishes a subscription to the resource, allowing for receiving updates.
	//
// Parameters.
	//   - ctx: context.Context. The context for the subscription.
	//
// Returns.
	//   - error: An error if subscription fails.
	Subscribe(ctx context.Context) error
}

// ManagerInterface defines the interface for managing a collection of resources.
//
// Summary: Interface for resource management.
//
// It provides methods for adding, removing, listing, and retrieving resources, as well
// as managing callbacks for list changes.
type ManagerInterface interface {
	// GetResource retrieves a resource by its URI.
	//
// Parameters.
	//   - uri: string. The URI of the resource.
	//
// Returns.
	//   - Resource: The resource instance.
	//   - bool: True if found, false otherwise.
	GetResource(uri string) (Resource, bool)

	// AddResource adds a new resource to the manager.
	//
// Parameters.
	//   - resource: Resource. The resource to add.
	//
	// Returns: - None.
	//   None.
	AddResource(resource Resource)

	// RemoveResource removes a resource from the manager by its URI.
	//
// Parameters.
	//   - uri: string. The URI of the resource to remove.
	//
	// Returns: - None.
	//   None.
	RemoveResource(uri string)

	// ListResources returns a slice of all resources currently in the manager.
	//
// Returns.
	//   - []Resource: A slice of resources.
	ListResources() []Resource

	// OnListChanged registers a callback function to be called when the list of resources changes.
	//
// Parameters.
	//   - f: func(). The callback function to execute on change.
	//
	// Returns: - None.
	//   None.
	OnListChanged(f func())

	// ClearResourcesForService removes all resources associated with a given service ID.
	//
// Parameters.
	//   - serviceID: string. The service ID.
	//
	// Returns: - None.
	//   None.
	ClearResourcesForService(serviceID string)
}

// Manager is a thread-safe implementation of the ManagerInterface.
//
// Summary: Thread-safe resource manager implementation.
//
// It manages the lifecycle and retrieval of resources, providing thread-safe access
// and efficient listing via caching.
type Manager struct {
	mu                sync.RWMutex
	resources         map[string]Resource
	onListChangedFunc func()
	cachedResources   []Resource
}

// NewManager provides newmanager functionality.
//
// Summary: NewManager.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func NewManager() *Manager {
	return &Manager{
		resources: make(map[string]Resource),
	}
}

// GetResource provides getresource functionality.
//
// Summary: GetResource.
//
// Parameters.
//   - uri: The parameter.
//   - bool: The parameter.
//
// Returns.
//   - None.
func (rm *Manager) GetResource(uri string) (Resource, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	resource, ok := rm.resources[uri]
	return resource, ok
}

// AddResource provides addresource functionality.
//
// Summary: AddResource.
//
// Parameters.
//   - resource: The parameter.
//
// Returns.
//   - None.
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

// RemoveResource provides removeresource functionality.
//
// Summary: RemoveResource.
//
// Parameters.
//   - uri: The parameter.
//
// Returns.
//   - None.
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

// ListResources provides listresources functionality.
//
// Summary: ListResources.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
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

// OnListChanged provides onlistchanged functionality.
//
// Summary: OnListChanged.
//
// Parameters.
//   - f: The parameter.
//
// Returns.
//   - None.
func (rm *Manager) OnListChanged(f func()) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.onListChangedFunc = f
}

// Subscribe provides subscribe functionality.
//
// Summary: Subscribe.
//
// Parameters.
//   - ctx: The parameter.
//   - uri: The parameter.
//
// Returns.
//   - result: The result.
func (rm *Manager) Subscribe(ctx context.Context, uri string) error {
	resource, ok := rm.GetResource(uri)
	if !ok {
		return ErrResourceNotFound
	}
	return resource.Subscribe(ctx)
}

// ClearResourcesForService provides clearresourcesforservice functionality.
//
// Summary: ClearResourcesForService.
//
// Parameters.
//   - serviceID: The parameter.
//
// Returns.
//   - None.
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
