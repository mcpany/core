// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"net/http"
	"sync"
)

// Handler defines the interface for handling webhook requests.
// Implementations of this interface process incoming webhook events.
//
// Summary: Interface for handling webhook requests.
type Handler interface {
	// Handle processes the webhook request.
	//
	// Parameters:
	//   w: The HTTP response writer to write the response to.
	//   r: The HTTP request containing the webhook payload.
	Handle(w http.ResponseWriter, r *http.Request)
}

// Registry manages the registration and retrieval of system webhooks.
// It provides a thread-safe mechanism to store and lookup handlers by name.
//
// Summary: Thread-safe registry for webhook handlers.
type Registry struct {
	mu    sync.RWMutex
	hooks map[string]Handler
}

// NewRegistry creates and initializes a new Registry instance.
//
// Summary: Creates a new webhook handler registry.
//
// Returns:
//   - *Registry: A pointer to the initialized Registry.
func NewRegistry() *Registry {
	return &Registry{
		hooks: make(map[string]Handler),
	}
}

// Register registers a handler with a specific name.
// If a handler with the same name already exists, it will be overwritten.
//
// Summary: Registers a webhook handler.
//
// Parameters:
//   - name: string. The name/path to register the handler under.
//   - handler: Handler. The handler instance to register.
func (r *Registry) Register(name string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks[name] = handler
}

// Get retrieves a handler by its name.
//
// Summary: Retrieves a registered webhook handler.
//
// Parameters:
//   - name: string. The name of the handler to retrieve.
//
// Returns:
//   - Handler: The registered handler instance.
//   - bool: True if the handler was found, false otherwise.
func (r *Registry) Get(name string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.hooks[name]
	return h, ok
}
