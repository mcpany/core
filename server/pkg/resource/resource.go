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

// Resource defines the interface for a resource that can be managed by the
// Manager. Each implementation of a resource is responsible for
// providing its metadata and handling read and subscribe operations.
type Resource interface {
	// Resource returns the MCP representation of the resource, which includes its
	// metadata.
	Resource() *mcp.Resource
	// Service returns the ID of the service that provides this resource.
	Service() string
	// Read retrieves the content of the resource.
	Read(ctx context.Context) (*mcp.ReadResourceResult, error)
	// Subscribe establishes a subscription to the resource, allowing for
	// receiving updates.
	Subscribe(ctx context.Context) error
}

// ManagerInterface defines the interface for managing a collection of
// resources. It provides methods for adding, removing, retrieving, and listing
// resources, as well as for subscribing to changes.
type ManagerInterface interface {
	// GetResource retrieves a resource by its URI.
	GetResource(uri string) (Resource, bool)
	// AddResource adds a new resource to the manager.
	AddResource(resource Resource)
	// RemoveResource removes a resource from the manager by its URI.
	RemoveResource(uri string)
	// ListResources returns a slice of all resources currently in the manager.
	ListResources() []Resource
	// OnListChanged registers a callback function to be called when the list of
	// resources changes.
	OnListChanged(func())
	// ClearResourcesForService removes all resources associated with a given service ID.
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
