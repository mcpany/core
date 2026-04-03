// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"net/http"
	"sync"
)

// Handler represents the public Handler entity.
//
// Summary: Defines the structured data model representing a .
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Handler interface {
	// Handle processes the webhook request.
	//
	// Parameters:
	//   w: http.ResponseWriter. The HTTP response writer to write the response to.
	//   r: *http.Request. The HTTP request containing the webhook payload.
	//
	// Returns:
	//
	//	None.
	//
	// Side Effects:
	//   - Writes the response to the response writer.
	Handle(w http.ResponseWriter, r *http.Request)
}

// Registry represents the public Registry entity.
//
// Summary: Defines the structured data model representing a .
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Registry struct {
	mu    sync.RWMutex
	hooks map[string]Handler
}

// NewRegistry serves as a public interface for interacting with NewRegistry.
//
// Summary: Constructs and returns an initialized registry ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewRegistry() *Registry {
	return &Registry{
		hooks: make(map[string]Handler),
	}
}

// Register serves as a public interface for interacting with Register.
//
// Summary: Register the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (r *Registry) Register(name string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks[name] = handler
}

// Get serves as a public interface for interacting with Get.
//
// Summary: Fetches and returns the underlying  from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (r *Registry) Get(name string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.hooks[name]
	return h, ok
}
