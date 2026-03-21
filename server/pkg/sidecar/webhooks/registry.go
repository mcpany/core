// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"net/http"
	"sync"
)

// Summary: Handler defines the interface for handling webhook requests. Implementations of this interface process incoming webhook events. Interface for webhook handlers.
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

// Summary: Registry manages the registration and retrieval of system webhooks. It provides a thread-safe mechanism to store and lookup handlers by name. Thread-safe registry for webhook handlers.
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

// Summary: NewRegistry creates and initializes a new Registry instance. Creates a new webhook registry.
//
// Parameters:
//   - None.
//
// Returns:
//   - *Registry: The resulting *Registry.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewRegistry() *Registry {
	return &Registry{
		hooks: make(map[string]Handler),
	}
}

// Summary: Register registers a handler with a specific name. If a handler with the same name already exists, it will be overwritten. Registers a webhook handler.
//
// Parameters:
//   - name (string): The name parameter.
//   - handler (Handler): The handler parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (r *Registry) Register(name string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks[name] = handler
}

// Summary: Get retrieves a handler by its name. Retrieves a webhook handler by name.
//
// Parameters:
//   - name (string): The name parameter.
//
// Returns:
//   - Handler: The resulting Handler.
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (r *Registry) Get(name string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.hooks[name]
	return h, ok
}
