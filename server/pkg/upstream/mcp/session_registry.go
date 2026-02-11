// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"sync"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SessionRegistry manages the mapping between upstream MCP sessions and downstream tool sessions.
// This allows requests from upstream (like sampling) to be routed to the correct downstream client.
type SessionRegistry struct {
	mu       sync.RWMutex
	sessions map[mcp.Session]tool.Session
}

// NewSessionRegistry creates a new SessionRegistry.
//
// Summary: Creates a new SessionRegistry.
//
// Returns:
//   - *SessionRegistry: The new registry.
func NewSessionRegistry() *SessionRegistry {
	return &SessionRegistry{
		sessions: make(map[mcp.Session]tool.Session),
	}
}

// Register registers a mapping between an upstream session and a downstream session.
//
// Summary: Registers a session mapping.
//
// Parameters:
//   - upstreamSession: mcp.Session. The upstream session.
//   - downstreamSession: tool.Session. The downstream session.
func (r *SessionRegistry) Register(upstreamSession mcp.Session, downstreamSession tool.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[upstreamSession] = downstreamSession
}

// Unregister removes the mapping for an upstream session.
//
// Summary: Unregisters a session mapping.
//
// Parameters:
//   - upstreamSession: mcp.Session. The upstream session to remove.
func (r *SessionRegistry) Unregister(upstreamSession mcp.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, upstreamSession)
}

// Get retrieves the downstream session associated with an upstream session.
//
// Summary: Retrieves the downstream session.
//
// Parameters:
//   - upstreamSession: mcp.Session. The upstream session.
//
// Returns:
//   - tool.Session: The downstream session.
//   - bool: True if found, false otherwise.
func (r *SessionRegistry) Get(upstreamSession mcp.Session) (tool.Session, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[upstreamSession]
	return s, ok
}
