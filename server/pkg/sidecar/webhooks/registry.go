// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"net/http"
	"sync"
)

// Handler defines the interface for handling webhook requests.
//
// Summary: defines the interface for handling webhook requests.
type Handler interface {
	// Handle processes the webhook request.
	//
	// Summary: processes the webhook request.
	//
	// Parameters:
	//   - w: http.ResponseWriter. The HTTP response writer.
	//   - r: *http.Request. The HTTP request.
	//
	// Returns:
	//   None.
	Handle(w http.ResponseWriter, r *http.Request)
}

// Registry manages the registration and retrieval of system webhooks.
//
// Summary: manages the registration and retrieval of system webhooks.
type Registry struct {
	mu    sync.RWMutex
	hooks map[string]Handler
}

// NewRegistry creates and initializes a new Registry instance.
//
// Summary: creates and initializes a new Registry instance.
//
// Parameters:
//   None.
//
// Returns:
//   - *Registry: The *Registry.
func NewRegistry() *Registry {
	return &Registry{
		hooks: make(map[string]Handler),
	}
}

// Register registers a handler with a specific name.
//
// Summary: registers a handler with a specific name.
//
// Parameters:
//   - name: string. The name.
//   - handler: Handler. The handler.
//
// Returns:
//   None.
func (r *Registry) Register(name string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks[name] = handler
}

// Get retrieves a handler by its name.
//
// Summary: retrieves a handler by its name.
//
// Parameters:
//   - name: string. The name.
//
// Returns:
//   - Handler: The Handler.
//   - bool: The bool.
func (r *Registry) Get(name string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.hooks[name]
	return h, ok
}
