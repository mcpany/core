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

// Resource defines the interface for a data source or content provider within the MCP Any system.
//
// Implementations of this interface expose data (e.g., database records, files, API responses)
// as standard MCP resources, which can be read by clients or subscribed to for real-time updates.
type Resource interface {
	// Resource returns the metadata describing this resource.
	//
	// This includes the URI, name, MIME type, and description, formatted for the MCP protocol.
	//
	// Returns:
	//   - *mcp.Resource: The MCP resource definition.
	Resource() *mcp.Resource

	// Service returns the identifier of the upstream service that owns this resource.
	//
	// This is used for access control, billing, and resource lifecycle management (e.g.,
	// clearing resources when a service is reloaded).
	//
	// Returns:
	//   - string: The unique service ID.
	Service() string

	// Read fetches the current content of the resource.
	//
	// It retrieves the data from the underlying source (e.g., executing a SQL query, reading a file)
	// and returns it in a format suitable for the MCP client (text or binary blob).
	//
	// Parameters:
	//   - ctx: The context for the read operation, including timeout and cancellation.
	//
	// Returns:
	//   - *mcp.ReadResourceResult: The content of the resource.
	//   - error: An error if the read operation fails (e.g., connection error, not found).
	Read(ctx context.Context) (*mcp.ReadResourceResult, error)

	// Subscribe attempts to establish a subscription for real-time updates on this resource.
	//
	// If supported, the implementation should start monitoring the underlying source and
	// notify the system when changes occur.
	//
	// Parameters:
	//   - ctx: The context for the subscription request.
	//
	// Returns:
	//   - error: An error if subscription is not supported or fails to be established.
	Subscribe(ctx context.Context) error
}

// ManagerInterface defines the contract for a component that manages the lifecycle and registry of resources.
//
// It acts as the central repository for all active resources, handling registration, lookup,
// and notification of changes to the resource list.
type ManagerInterface interface {
	// GetResource looks up a registered resource by its unique URI.
	//
	// Parameters:
	//   - uri: The URI of the resource to find.
	//
	// Returns:
	//   - Resource: The resource instance, if found.
	//   - bool: True if the resource exists, false otherwise.
	GetResource(uri string) (Resource, bool)

	// AddResource registers a new resource with the manager.
	//
	// If a resource with the same URI already exists, it is updated.
	// This triggers any registered "list changed" callbacks.
	//
	// Parameters:
	//   - resource: The resource instance to add.
	AddResource(resource Resource)

	// RemoveResource unregisters a resource identified by its URI.
	//
	// If the resource exists, it is removed and "list changed" callbacks are triggered.
	//
	// Parameters:
	//   - uri: The URI of the resource to remove.
	RemoveResource(uri string)

	// ListResources returns a snapshot of all currently registered resources.
	//
	// Returns:
	//   - []Resource: A slice containing all resources.
	ListResources() []Resource

	// OnListChanged registers a callback function that is invoked whenever the list of resources changes.
	//
	// This is typically used to notify MCP clients that the available resources have been updated.
	//
	// Parameters:
	//   - f: The function to execute when the resource list changes.
	OnListChanged(f func())

	// ClearResourcesForService removes all resources belonging to a specific upstream service.
	//
	// This is commonly used during configuration reloads or service shutdowns to clean up
	// associated resources.
	//
	// Parameters:
	//   - serviceID: The ID of the service whose resources should be removed.
	ClearResourcesForService(serviceID string)
}

// Manager is a thread-safe implementation of the
// ManagerInterface. It uses a map to store resources and a mutex to
// protect concurrent access.
type Manager struct {
	mu                sync.RWMutex
	resources         map[string]Resource
	onListChangedFunc func()
	cachedResources   []Resource
}

// NewManager creates and returns a new, empty Manager.
//
// Returns the result.
func NewManager() *Manager {
	return &Manager{
		resources: make(map[string]Resource),
	}
}

// GetResource retrieves a resource from the manager by its URI.
//
// uri is the URI of the resource to retrieve.
// It returns the resource and a boolean indicating whether the resource was
// found.
func (rm *Manager) GetResource(uri string) (Resource, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	resource, ok := rm.resources[uri]
	return resource, ok
}

// AddResource adds a new resource to the manager. If a resource with the same
// URI already exists, it will be overwritten. After adding the resource, it
// triggers the OnListChanged callback if one is registered.
//
// resource is the resource to be added.
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

// RemoveResource removes a resource from the manager by its URI. If the
// resource exists, it is removed, and the OnListChanged callback is triggered if
// one is registered.
//
// uri is the URI of the resource to be removed.
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

// ListResources returns a slice containing all the resources currently
// registered in the manager.
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
// f is the callback function to be set.
func (rm *Manager) OnListChanged(f func()) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.onListChangedFunc = f
}

// Subscribe finds a resource by its URI and calls its Subscribe method.
//
// ctx is the context for the subscription.
// uri is the URI of the resource to subscribe to.
// It returns an error if the resource is not found or if the subscription fails.
func (rm *Manager) Subscribe(ctx context.Context, uri string) error {
	resource, ok := rm.GetResource(uri)
	if !ok {
		return ErrResourceNotFound
	}
	return resource.Subscribe(ctx)
}

// ClearResourcesForService removes all resources associated with a given service ID.
//
// serviceID is the serviceID.
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
