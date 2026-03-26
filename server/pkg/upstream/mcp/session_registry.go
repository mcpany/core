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
// Summary: Registry for mapping upstream MCP sessions to downstream tool sessions.
type SessionRegistry struct {
	mu       sync.RWMutex
	sessions map[mcp.Session]tool.Session
}

// NewSessionRegistry creates a new SessionRegistry.
//
// Summary: Initializes a new SessionRegistry.
//
// Returns:
//   - *SessionRegistry: A pointer to the newly created SessionRegistry.
//
// Parameters:
//   - None.
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

// Register registers a mapping between an upstream session and a downstream session.
//
// Summary: Executes the Register operation to map an upstream session to a downstream one.
//
// Parameters:
//   - upstreamSession (mcp.Session): The upstream MCP session.
//   - downstreamSession (tool.Session): The corresponding downstream tool session.
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

// Unregister removes the mapping for an upstream session.
//
// Summary: Executes the Unregister operation to remove a session mapping.
//
// Parameters:
//   - upstreamSession (mcp.Session): The upstream session to unregister.
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

// Get retrieves the downstream session associated with an upstream session.
//
// Summary: Retrieves the downstream session for a given upstream session.
//
// Parameters:
//   - upstreamSession (mcp.Session): The upstream session to look up.
//
// Returns:
//   - tool.Session: The associated downstream session, if found.
//   - bool: True if the session was found, false otherwise.
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
