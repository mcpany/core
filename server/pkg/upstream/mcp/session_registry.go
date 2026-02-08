// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"sync"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SessionRegistry manages the mapping between upstream MCP sessions and downstream tool sessions.
//
// Summary: manages the mapping between upstream MCP sessions and downstream tool sessions.
type SessionRegistry struct {
	mu       sync.RWMutex
	sessions map[mcp.Session]tool.Session
}

// NewSessionRegistry creates a new SessionRegistry.
//
// Summary: creates a new SessionRegistry.
//
// Parameters:
//   None.
//
// Returns:
//   - *SessionRegistry: The *SessionRegistry.
func NewSessionRegistry() *SessionRegistry {
	return &SessionRegistry{
		sessions: make(map[mcp.Session]tool.Session),
	}
}

// Register registers a mapping between an upstream session and a downstream session.
//
// Summary: registers a mapping between an upstream session and a downstream session.
//
// Parameters:
//   - upstreamSession: mcp.Session. The upstreamSession.
//   - downstreamSession: tool.Session. The downstreamSession.
//
// Returns:
//   None.
func (r *SessionRegistry) Register(upstreamSession mcp.Session, downstreamSession tool.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[upstreamSession] = downstreamSession
}

// Unregister removes the mapping for an upstream session.
//
// Summary: removes the mapping for an upstream session.
//
// Parameters:
//   - upstreamSession: mcp.Session. The upstreamSession.
//
// Returns:
//   None.
func (r *SessionRegistry) Unregister(upstreamSession mcp.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, upstreamSession)
}

// Get retrieves the downstream session associated with an upstream session.
//
// Summary: retrieves the downstream session associated with an upstream session.
//
// Parameters:
//   - upstreamSession: mcp.Session. The upstreamSession.
//
// Returns:
//   - tool.Session: The tool.Session.
//   - bool: The bool.
func (r *SessionRegistry) Get(upstreamSession mcp.Session) (tool.Session, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[upstreamSession]
	return s, ok
}
