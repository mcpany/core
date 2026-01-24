// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPSession wraps an MCP session to provide client interaction capabilities like sampling and roots.
type MCPSession struct {
	session  *mcp.ServerSession
	debugger *middleware.Debugger
}

// NewMCPSession creates a new MCPSession.
//
// session is the session.
// debugger is the debugger instance for tracing.
//
// Returns the result.
func NewMCPSession(session *mcp.ServerSession, debugger *middleware.Debugger) *MCPSession {
	return &MCPSession{session: session, debugger: debugger}
}

// NewMCPSampler is a deprecated alias for NewMCPSession.
//
// session is the session.
//
// Returns the result.
func NewMCPSampler(session *mcp.ServerSession) *MCPSession {
	return NewMCPSession(session, nil)
}

// CreateMessage requests a message creation from the client (sampling).
//
// ctx is the context for the request.
// params is the params.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *MCPSession) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for sampling")
	}

	start := time.Now()
	res, err := s.session.CreateMessage(ctx, params)
	duration := time.Since(start)

	if s.debugger != nil {
		reqBytes, _ := json.Marshal(params)
		resBytes, _ := json.Marshal(res)
		status := http.StatusOK
		if err != nil {
			errBytes, _ := json.Marshal(map[string]string{"error": err.Error()})
			resBytes = errBytes
			status = http.StatusInternalServerError
		}

		s.debugger.AddEntry(middleware.DebugEntry{
			ID:        uuid.New().String(),
			Timestamp: start,
			Method:    "SAMPLING",
			Path:      "/mcp/sampling/create_message",
			Status:    status,
			Duration:  duration,
			RequestHeaders: http.Header{
				"Content-Type": []string{"application/json"},
			},
			ResponseHeaders: http.Header{
				"Content-Type": []string{"application/json"},
			},
			RequestBody:  string(reqBytes),
			ResponseBody: string(resBytes),
		})
	}

	return res, err
}

// ListRoots requests the list of roots from the client.
//
// ctx is the context for the request.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *MCPSession) ListRoots(ctx context.Context) (*mcp.ListRootsResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for roots inspection")
	}
	// The SDK exposes ListRoots on ServerSession
	return s.session.ListRoots(ctx, nil)
}

// Verify that MCPSession implements tool.Session.
var _ tool.Session = (*MCPSession)(nil)
