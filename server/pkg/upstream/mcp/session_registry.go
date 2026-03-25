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

// NewSessionRegistry creates a new SessionRegistry.
//
// Returns: - None.
//   - *SessionRegistry: The result.
//
// Side Effects: - None.
//   - None.
//
// Summary: Initializes NewSessionRegistry operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func NewSessionRegistry() *SessionRegistry {
	return &SessionRegistry{
		sessions: make(map[mcp.Session]tool.Session),
	}
}

// Register registers a mapping between an upstream session and a downstream session.
//
// Parameters: - None.
//   - upstreamSession (mcp.Session): The parameter.
//   - downstreamSession (tool.Session): The parameter.
//
// Returns: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
//
// Summary: Executes Register operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (r *SessionRegistry) Register(upstreamSession mcp.Session, downstreamSession tool.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[upstreamSession] = downstreamSession
}

// Unregister removes the mapping for an upstream session.
//
// Parameters: - None.
//   - upstreamSession (mcp.Session): The parameter.
//
// Returns: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
//
// Summary: Executes Unregister operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (r *SessionRegistry) Unregister(upstreamSession mcp.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, upstreamSession)
}

// Get retrieves the downstream session associated with an upstream session.
//
// Parameters: - None.
//   - upstreamSession (mcp.Session): The parameter.
//
// Returns: - None.
//   - tool.Session: The result.
//   - bool: The result.
//
// Side Effects: - None.
//   - None.
//
// Summary: Retrieves Get operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (r *SessionRegistry) Get(upstreamSession mcp.Session) (tool.Session, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[upstreamSession]
	return s, ok
}
