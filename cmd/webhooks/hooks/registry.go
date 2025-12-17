// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package hooks

import (
	"net/http"
	"sync"
)

// Handler handles a webhook request.
type Handler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

// Registry manages system webhooks.
type Registry struct {
	mu    sync.RWMutex
	hooks map[string]Handler
}

// NewRegistry creates a new Registry.
func NewRegistry() *Registry {
	return &Registry{
		hooks: make(map[string]Handler),
	}
}

// Register registers a handler with a name.
func (r *Registry) Register(name string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks[name] = handler
}

// Get returns a handler by name.
func (r *Registry) Get(name string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.hooks[name]
	return h, ok
}
