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
//
// Summary: Represents a SessionRegistry.
type SessionRegistry struct {
	mu       sync.RWMutex
	sessions map[mcp.Session]tool.Session
}

// NewSessionRegistry creates a new session registry.
//
// Summary: Creates a new session registry.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - *SessionRegistry: The result.
func NewSessionRegistry() *SessionRegistry {
	return &SessionRegistry{
		sessions: make(map[mcp.Session]tool.Session),
	}
}

// Register register register.
//
// Summary: Register register.
//
// Parameters: - None.
//   - upstreamSession (mcp.Session): The upstream session.
//   - downstreamSession (tool.Session): The downstream session.
//
// Returns: - None.
//   - None.
func (r *SessionRegistry) Register(upstreamSession mcp.Session, downstreamSession tool.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[upstreamSession] = downstreamSession
}

// Unregister unregister unregister.
//
// Summary: Unregister unregister.
//
// Parameters: - None.
//   - upstreamSession (mcp.Session): The upstream session.
//
// Returns: - None.
//   - None.
func (r *SessionRegistry) Unregister(upstreamSession mcp.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, upstreamSession)
}

// Get retrieves the .
//
// Summary: Retrieves the .
//
// Parameters: - None.
//   - upstreamSession (mcp.Session): The upstream session.
//
// Returns: - None.
//   - tool.Session: The result.
//   - bool: The result.
func (r *SessionRegistry) Get(upstreamSession mcp.Session) (tool.Session, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[upstreamSession]
	return s, ok
}
