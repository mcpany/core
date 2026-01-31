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
	//   w: The HTTP response writer to write the response to.
	//   r: The HTTP request containing the webhook payload.
	Handle(w http.ResponseWriter, r *http.Request)
}

// Registry manages the registration and retrieval of system webhooks.
// It provides a thread-safe mechanism to store and lookup handlers by name.
type Registry struct {
	mu    sync.RWMutex
	hooks map[string]Handler
}

// NewRegistry creates and initializes a new Registry instance.
//
// Returns:
//   A pointer to a new, empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		hooks: make(map[string]Handler),
	}
}

// Register registers a handler with a specific name.
// If a handler with the same name already exists, it will be overwritten.
//
// Parameters:
//   name: The name/path to register the handler under.
//   handler: The Handler instance to register.
func (r *Registry) Register(name string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks[name] = handler
}

// Get retrieves a handler by its name.
//
// Parameters:
//   name: The name of the handler to retrieve.
//
// Returns:
//   Handler: The registered handler, if found.
//   bool: True if the handler exists, false otherwise.
func (r *Registry) Get(name string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.hooks[name]
	return h, ok
}
