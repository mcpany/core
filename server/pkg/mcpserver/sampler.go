package mcpserver

import (
	"context"
	"fmt"

	"github.com/mcpany/core/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPSampler wraps an MCP session to provide sampling capabilities.
type MCPSampler struct {
	session *mcp.ServerSession
}

// NewMCPSampler creates a new MCPSampler.
func NewMCPSampler(session *mcp.ServerSession) *MCPSampler {
	return &MCPSampler{session: session}
}

// CreateMessage requests a message creation from the client (sampling).
func (s *MCPSampler) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for sampling")
	}
	return s.session.CreateMessage(ctx, params)
}

// Verify that MCPSampler implements tool.Sampler
var _ tool.Sampler = (*MCPSampler)(nil)
