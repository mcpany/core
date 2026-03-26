// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"net/http"
	"sync"
)

// Handler handler represents a handler.
//
// Summary: Handler represents a handler.
type Handler interface {
	// Handle processes the webhook request.
	//
	// Parameters: - None.
	//   w: http.ResponseWriter. The HTTP response writer to write the response to.
	//   r: *http.Request. The HTTP request containing the webhook payload.
	//
	// Returns: - None.
	//
	//	None.
	//
	// Side Effects: - None.
	//   - Writes the response to the response writer.
	Handle(w http.ResponseWriter, r *http.Request)
}

// Registry registry represents a registry.
//
// Summary: Registry represents a registry.
type Registry struct {
	mu    sync.RWMutex
	hooks map[string]Handler
}

// NewRegistry creates and initializes a new Registry instance.
//
// Summary: Creates a new webhook registry.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - *Registry: A pointer to a new, empty Registry.
//
// Side Effects: - None.
//   - Allocates memory for the registry map.
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
// Parameters: - None.
//   - name: string. The name/path to register the handler under.
//   - handler: Handler. The Handler instance to register.
//
// Returns: - None.
//
//	None.
//
// Side Effects: - None.
//   - Updates the registry map.
func (r *Registry) Register(name string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks[name] = handler
}

// Get retrieves a handler by its name.
//
// Summary: Retrieves a webhook handler by name.
//
// Parameters: - None.
//   - name: string. The name of the handler to retrieve.
//
// Returns: - None.
//   - Handler: The registered handler, if found.
//   - bool: True if the handler exists, false otherwise.
//
// Side Effects: - None.
//   - None.
func (r *Registry) Get(name string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.hooks[name]
	return h, ok
}
