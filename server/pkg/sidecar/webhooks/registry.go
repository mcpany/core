// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"net/http"
	"sync"
)

// Handler defines the interface for handling webhook requests.
//
// Summary: Defines the interface for handling webhook requests.
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

// Registry manages the registration and retrieval of system webhooks.
//
// Summary: Manages the registration and retrieval of system webhooks.
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

// NewRegistry creates and initializes a new Registry instance.
//
// Summary: Creates and initializes a new Registry instance.
//
// Parameters:
//   - None.
//
// Returns:
//   - *Registry: Return value.
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

// Register registers a handler with a specific name.
//
// Summary: Registers a handler with a specific name.
//
// Parameters:
//   - name (string): Parameter.
//   - handler (Handler): Parameter.
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

// Get retrieves a handler by its name.
//
// Summary: Retrieves a handler by its name.
//
// Parameters:
//   - name (string): Parameter.
//
// Returns:
//   - Handler: Return value.
//   - bool: Return value.
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
