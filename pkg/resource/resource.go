/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
// ResourceManager. Each implementation of a resource is responsible for
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

// ResourceManagerInterface defines the interface for managing a collection of
// resources. It provides methods for adding, removing, retrieving, and listing
// resources, as well as for subscribing to changes.
type ResourceManagerInterface interface {
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
	// ClearAllResources removes all resources from the manager.
	ClearAllResources()
}

// ResourceManager is a thread-safe implementation of the
// ResourceManagerInterface. It uses a map to store resources and a mutex to
// protect concurrent access.
type ResourceManager struct {
	mu                sync.RWMutex
	resources         map[string]Resource
	onListChangedFunc func()
}

// NewResourceManager creates and returns a new, empty ResourceManager.
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources: make(map[string]Resource),
	}
}

// GetResource retrieves a resource from the manager by its URI.
//
// uri is the URI of the resource to retrieve.
// It returns the resource and a boolean indicating whether the resource was
// found.
func (rm *ResourceManager) GetResource(uri string) (Resource, bool) {
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
func (rm *ResourceManager) AddResource(resource Resource) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.resources[resource.Resource().URI] = resource
	if rm.onListChangedFunc != nil {
		rm.onListChangedFunc()
	}
}

// ClearAllResources removes all resources from the manager.
func (rm *ResourceManager) ClearAllResources() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.resources = make(map[string]Resource)
	if rm.onListChangedFunc != nil {
		rm.onListChangedFunc()
	}
}

// RemoveResource removes a resource from the manager by its URI. If the
// resource exists, it is removed, and the OnListChanged callback is triggered if
// one is registered.
//
// uri is the URI of the resource to be removed.
func (rm *ResourceManager) RemoveResource(uri string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if _, ok := rm.resources[uri]; ok {
		delete(rm.resources, uri)
		if rm.onListChangedFunc != nil {
			rm.onListChangedFunc()
		}
	}
}

// ListResources returns a slice containing all the resources currently
// registered in the manager.
func (rm *ResourceManager) ListResources() []Resource {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	resources := make([]Resource, 0, len(rm.resources))
	for _, resource := range rm.resources {
		resources = append(resources, resource)
	}
	return resources
}

// OnListChanged sets a callback function that will be invoked whenever the list
// of resources is modified by adding or removing a resource.
//
// f is the callback function to be set.
func (rm *ResourceManager) OnListChanged(f func()) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.onListChangedFunc = f
}

// Subscribe finds a resource by its URI and calls its Subscribe method.
//
// ctx is the context for the subscription.
// uri is the URI of the resource to subscribe to.
// It returns an error if the resource is not found or if the subscription fails.
func (rm *ResourceManager) Subscribe(ctx context.Context, uri string) error {
	resource, ok := rm.GetResource(uri)
	if !ok {
		return ErrResourceNotFound
	}
	return resource.Subscribe(ctx)
}

// ClearResourcesForService removes all resources associated with a given service ID.
func (rm *ResourceManager) ClearResourcesForService(serviceID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	for uri, resource := range rm.resources {
		if resource.Service() == serviceID {
			delete(rm.resources, uri)
		}
	}
	if rm.onListChangedFunc != nil {
		rm.onListChangedFunc()
	}
}
