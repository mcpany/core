// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"

	"github.com/mcpany/core/pkg/tool"
	v1 "github.com/mcpany/core/proto/api/v1"
	mcprouterv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToolServer implements the ToolServiceServer interface.
type ToolServer struct {
	v1.UnimplementedToolServiceServer
	toolManager tool.ManagerInterface
}

// NewToolServer creates a new ToolServer.
func NewToolServer(toolManager tool.ManagerInterface) *ToolServer {
	return &ToolServer{toolManager: toolManager}
}

// ListTools returns all registered tools.
func (s *ToolServer) ListTools(_ context.Context, _ *v1.ListToolsRequest) (*v1.ListToolsResponse, error) {
	if s.toolManager == nil {
		return nil, status.Error(codes.FailedPrecondition, "tool manager not initialized")
	}
	tools := s.toolManager.ListTools()
	responseTools := make([]*mcprouterv1.Tool, 0, len(tools))
	for _, t := range tools {
		responseTools = append(responseTools, t.Tool())
	}
	return &v1.ListToolsResponse{Tools: responseTools}, nil
}

// GetTool returns a specific tool by name.
func (s *ToolServer) GetTool(_ context.Context, req *v1.GetToolRequest) (*v1.GetToolResponse, error) {
	if s.toolManager == nil {
		return nil, status.Error(codes.FailedPrecondition, "tool manager not initialized")
	}
	t, ok := s.toolManager.GetTool(req.GetToolName())
	if !ok {
		return nil, status.Error(codes.NotFound, "tool not found")
	}
	return &v1.GetToolResponse{Tool: t.Tool()}, nil
}

// RegisterTools registers new tools (Not Implemented).
func (s *ToolServer) RegisterTools(_ context.Context, _ *v1.RegisterToolsRequest) (*v1.RegisterToolsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method RegisterTools not implemented")
}
