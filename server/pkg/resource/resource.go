// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resource

import (
// Resource defines the interface for a resource that can be managed by the Manager.
//
// Summary: Interface for a managed resource.
//
// A resource represents a data source (e.g., a file, a database record) that can be
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// read by an MCP client.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
// ManagerInterface defines the interface for managing a collection of resources.
//
// Summary: Interface for resource management.
//
// It provides methods for adding, removing, listing, and retrieving resources, as well
// as managing callbacks for list changes.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
	//
	// Returns:
	//   None.
	AddResource(resource Resource)

	// RemoveResource removes a resource from the manager by its URI.
	//
	// Parameters:
	//   - uri: string. The URI of the resource to remove.
	//
	// Returns:
	//   None.
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
	//
	// Returns:
	//   None.
	OnListChanged(f func())

	// ClearResourcesForService removes all resources associated with a given service ID.
	//
	// Parameters:
// Manager is a thread-safe implementation of the ManagerInterface.
//
// Summary: Thread-safe resource manager implementation.
//
// It manages the lifecycle and retrieval of resources, providing thread-safe access
// and efficient listing via caching.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// NewManager creates and returns a new, empty Manager.
// GetResource retrieves a resource from the manager by its URI.
// AddResource adds a new resource to the manager.
//
// Summary: Adds a resource to the manager.
//
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Parameters:
//   - resource: Resource. The resource to add.
//
// Returns:
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//	None.
//
// Side Effects:
// RemoveResource removes a resource from the manager by its URI.
//
// Summary: Removes a resource from the manager.
// Errors:
//   - triggers relevant error states on failure.
//
// Parameters:
//   - uri: string. The URI of the resource.
//
// Returns:
//
//	None.
//
// Side Effects:
//   - Updates the internal resource storage.
//   - Invalidates the list cache.
//   - Triggers the on-change callback if registered.
// Errors:
//   - triggers relevant error states on failure.
func (rm *Manager) RemoveResource(uri string) {
	var callback func()
	rm.mu.Lock()
	if _, ok := rm.resources[uri]; ok {
		delete(rm.resources, uri)
// ListResources returns a slice containing all the resources currently registered in the manager.
//
// Summary: Lists all managed resources.
//
// It uses a read-through cache (double-checked locking) to minimize allocation overhead
// for frequent calls.
//
// Returns:
//   - []Resource: A slice of currently registered resources.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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

// OnListChanged sets a callback function that will be invoked whenever the list
// of resources is modified.
//
// Summary: Registers a callback for list changes.
//
// Parameters:
// Subscribe finds a resource by its URI and calls its Subscribe method.
//
// ClearResourcesForService removes all resources associated with a given service ID.
//
// Summary: Clears resources for a specific service.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Parameters:
//   - serviceID: string. The service ID.
//
// Returns:
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
//	None.
//
// Side Effects:
//   - Removes matching resources from storage.
//   - Invalidates the list cache.
//   - Triggers the on-change callback.
// Errors:
//   - triggers relevant error states on failure.
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
