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
// Summary. Interface for webhook handlers.
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
// Side Effects:
	//   - Writes the response to the response writer.
	Handle(w http.ResponseWriter, r *http.Request)
}

// Registry manages the registration and retrieval of system webhooks.
// It provides a thread-safe mechanism to store and lookup handlers by name.
//
// Summary. Thread-safe registry for webhook handlers.
type Registry struct {
	mu    sync.RWMutex
	hooks map[string]Handler
}

// NewRegistry provides newregistry functionality.
//
// Summary: NewRegistry.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func NewRegistry() *Registry {
	return &Registry{
		hooks: make(map[string]Handler),
	}
}

// Register provides register functionality.
//
// Summary: Register.
//
// Parameters.
//   - name: The parameter.
//   - handler: The parameter.
//
// Returns.
//   - None.
func (r *Registry) Register(name string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks[name] = handler
}

// Get provides get functionality.
//
// Summary: Get.
//
// Parameters.
//   - name: The parameter.
//   - bool: The parameter.
//
// Returns.
//   - None.
func (r *Registry) Get(name string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.hooks[name]
	return h, ok
}
