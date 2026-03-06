// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"sync"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SessionRegistry - Auto-generated documentation.
//
// Summary: SessionRegistry manages the mapping between upstream MCP sessions and downstream tool sessions.
//
// Fields:
//   - Various fields for SessionRegistry.
type SessionRegistry struct {
	mu       sync.RWMutex
	sessions map[mcp.Session]tool.Session
}

// NewSessionRegistry - Auto-generated documentation.
//
// Summary: NewSessionRegistry creates a new SessionRegistry.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func NewSessionRegistry() *SessionRegistry {
	return &SessionRegistry{
		sessions: make(map[mcp.Session]tool.Session),
	}
}

// Register registers a mapping between an upstream session and a downstream session.
//
// Parameters:
//   - upstreamSession (mcp.Session): The parameter.
//   - downstreamSession (tool.Session): The parameter.
//
// Returns:
//   - None.
//
// Side Effects:
//   - None.
func (r *SessionRegistry) Register(upstreamSession mcp.Session, downstreamSession tool.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[upstreamSession] = downstreamSession
}

// Unregister removes the mapping for an upstream session.
//
// Parameters:
//   - upstreamSession (mcp.Session): The parameter.
//
// Returns:
//   - None.
//
// Side Effects:
//   - None.
func (r *SessionRegistry) Unregister(upstreamSession mcp.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, upstreamSession)
}

// Get retrieves the downstream session associated with an upstream session.
//
// Parameters:
//   - upstreamSession (mcp.Session): The parameter.
//
// Returns:
//   - tool.Session: The result.
//   - bool: The result.
//
// Side Effects:
//   - None.
func (r *SessionRegistry) Get(upstreamSession mcp.Session) (tool.Session, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[upstreamSession]
	return s, ok
}
