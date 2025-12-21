// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package sampling provides functionality for MCP Sampling, allowing tools
// to request LLM completions from the connected client.
package sampling

import (
	"context"
	"fmt"

	"github.com/mcpany/core/pkg/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Sampler defines the interface for creating sampling messages (LLM completions).
type Sampler interface {
	CreateMessage(ctx context.Context, req *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error)
}

// MCPServerProvider defines an interface for components that can provide an
// instance of an *mcp.Server.
type MCPServerProvider interface {
	Server() *mcp.Server
}

// MCPSampler implements the Sampler interface using an MCP Server.
type MCPSampler struct {
	serverProvider MCPServerProvider
}

// NewMCPSampler creates a new MCPSampler.
func NewMCPSampler(p MCPServerProvider) *MCPSampler {
	return &MCPSampler{
		serverProvider: p,
	}
}

// CreateMessage sends a sampling request to the client.
func (s *MCPSampler) CreateMessage(ctx context.Context, req *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	server := s.serverProvider.Server()
	if server == nil {
		return nil, fmt.Errorf("mcp server not initialized")
	}

	// Strategy: Use the first available session.
	// In a future multi-tenant implementation, we should retrieve the session ID from the context.
	sessions := server.Sessions()

	// Iterate to get the first session
	var session *mcp.ServerSession
	for sess := range sessions {
		session = sess
		break
	}

	if session == nil {
		logging.GetLogger().Warn("No active MCP session found for sampling")
		return nil, fmt.Errorf("no active mcp session found")
	}

	logging.GetLogger().Debug("Sending sampling request", "sessionID", session.ID())
	return session.CreateMessage(ctx, req)
}

type contextKey string

const samplerKey = contextKey("sampler")

// NewContextWithSampler creates a new context with the given Sampler.
func NewContextWithSampler(ctx context.Context, s Sampler) context.Context {
	return context.WithValue(ctx, samplerKey, s)
}

// FromContext retrieves a Sampler from the context.
func FromContext(ctx context.Context) (Sampler, bool) {
	s, ok := ctx.Value(samplerKey).(Sampler)
	return s, ok
}
