// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"net/http"
	"sync"
)

// Handler defines the interface for handling webhook requests.
// Implementations of this interface process incoming webhook events.
type Handler interface {
	// Handle processes the webhook request.
	//
	// Parameters:
	//   - w: http.ResponseWriter. The HTTP response writer to write the response to.
	//   - r: *http.Request. The HTTP request containing the webhook payload.
	Handle(w http.ResponseWriter, r *http.Request)
}

// Registry manages the registration and retrieval of system webhooks.
// It provides a thread-safe mechanism to store and lookup handlers by name.
type Registry struct {
	mu    sync.RWMutex
	hooks map[string]Handler
}

// NewRegistry creates a new thread-safe registry for managing webhook handlers.
//
// Returns:
//   - *Registry: A pointer to the initialized Registry.
func NewRegistry() *Registry {
	return &Registry{
		hooks: make(map[string]Handler),
	}
}

// Register adds or updates a webhook handler in the registry.
//
// Parameters:
//   - name: string. The unique name or path to register the handler under.
//   - handler: Handler. The handler instance to register.
//
// Side Effects:
//   - Acquires a write lock on the registry map.
//   - Modifies the internal map of handlers.
func (r *Registry) Register(name string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks[name] = handler
}

// Get lookups a webhook handler from the registry.
//
// Parameters:
//   - name: string. The name of the handler to retrieve.
//
// Returns:
//   - Handler: The registered handler, if found.
//   - bool: True if the handler exists, false otherwise.
//
// Side Effects:
//   - Acquires a read lock on the registry map.
func (r *Registry) Get(name string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.hooks[name]
	return h, ok
}
