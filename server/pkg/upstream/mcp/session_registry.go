// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"sync"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Summary: SessionRegistry manages the mapping between upstream MCP sessions and downstream tool sessions. This allows requests from upstream (like sampling) to be routed to the correct downstream client. Represents a SessionRegistry.
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
type SessionRegistry struct {
	mu       sync.RWMutex
	sessions map[mcp.Session]tool.Session
}

// Summary: NewSessionRegistry creates a new SessionRegistry.
//
// Parameters:
//   - None.
//
// Returns:
//   - *SessionRegistry: The resulting *SessionRegistry.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewSessionRegistry() *SessionRegistry {
	return &SessionRegistry{
		sessions: make(map[mcp.Session]tool.Session),
	}
}

// Summary: Register registers a mapping between an upstream session and a downstream session.
//
// Parameters:
//   - upstreamSession (mcp.Session): The upstreamSession parameter.
//   - downstreamSession (tool.Session): The downstreamSession parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (r *SessionRegistry) Register(upstreamSession mcp.Session, downstreamSession tool.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[upstreamSession] = downstreamSession
}

// Summary: Unregister removes the mapping for an upstream session.
//
// Parameters:
//   - upstreamSession (mcp.Session): The upstreamSession parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (r *SessionRegistry) Unregister(upstreamSession mcp.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, upstreamSession)
}

// Summary: Get retrieves the downstream session associated with an upstream session.
//
// Parameters:
//   - upstreamSession (mcp.Session): The upstreamSession parameter.
//
// Returns:
//   - tool.Session: The resulting tool.Session.
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (r *SessionRegistry) Get(upstreamSession mcp.Session) (tool.Session, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[upstreamSession]
	return s, ok
}
