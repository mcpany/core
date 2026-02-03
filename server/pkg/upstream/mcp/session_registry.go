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
// Returns the result.
//
// Returns:
//   - *SessionRegistry: The result.
func NewSessionRegistry() *SessionRegistry {
	return &SessionRegistry{
		sessions: make(map[mcp.Session]tool.Session),
	}
}

// Register registers a mapping between an upstream session and a downstream session.
//
// upstreamSession is the upstreamSession.
// downstreamSession is the downstreamSession.
//
// Parameters:
//   - upstreamSession: mcp.Session. The upstreamSession parameter.
//   - downstreamSession: tool.Session. The downstreamSession parameter.
func (r *SessionRegistry) Register(upstreamSession mcp.Session, downstreamSession tool.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[upstreamSession] = downstreamSession
}

// Unregister removes the mapping for an upstream session.
//
// upstreamSession is the upstreamSession.
//
// Parameters:
//   - upstreamSession: mcp.Session. The upstreamSession parameter.
func (r *SessionRegistry) Unregister(upstreamSession mcp.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, upstreamSession)
}

// Get retrieves the downstream session associated with an upstream session.
//
// upstreamSession is the upstreamSession.
//
// Returns the result.
// Returns true if successful.
//
// Parameters:
//   - upstreamSession: mcp.Session. The upstreamSession parameter.
//
// Returns:
//   - tool.Session: The result.
//   - bool: The result.
func (r *SessionRegistry) Get(upstreamSession mcp.Session) (tool.Session, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[upstreamSession]
	return s, ok
}
