// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"fmt"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// StartupErrorTool is a special tool that reports startup errors.
type StartupErrorTool struct {
	err         error
	mcpTool     *mcp.Tool
	mcpToolOnce sync.Once
	v1Tool      *v1.Tool
}

// NewStartupErrorTool creates a new StartupErrorTool.
func NewStartupErrorTool(err error) *StartupErrorTool {
	v1Tool := &v1.Tool{
		Name:        proto.String("get_startup_status"),
		ServiceId:   proto.String("system"),
		Description: proto.String("Returns the status of the server startup. Use this to diagnose why the server failed to start normally."),
		InputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"type": structpb.NewStringValue("object"),
				"properties": structpb.NewStructValue(&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"reason": structpb.NewStructValue(&structpb.Struct{
							Fields: map[string]*structpb.Value{
								"type":        structpb.NewStringValue("string"),
								"description": structpb.NewStringValue("The reason for checking status (optional)"),
							},
						}),
					},
				}),
			},
		},
	}

	mcpTool, convertErr := tool.ConvertProtoToMCPTool(v1Tool)
	if convertErr != nil {
		logging.GetLogger().Error("Failed to convert startup error tool to MCP tool", "error", convertErr)
		// This should never happen with a valid static definition
		panic(fmt.Errorf("failed to convert startup error tool: %w", convertErr))
	}

	return &StartupErrorTool{
		err:     err,
		v1Tool:  v1Tool,
		mcpTool: mcpTool,
	}
}

// Tool returns the protobuf definition of the tool.
func (t *StartupErrorTool) Tool() *v1.Tool {
	return t.v1Tool
}

// MCPTool returns the MCP tool definition.
func (t *StartupErrorTool) MCPTool() *mcp.Tool {
	return t.mcpTool
}

// Execute runs the tool with the provided context and request.
func (t *StartupErrorTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return map[string]any{
		"status":  "error",
		"error":   t.err.Error(),
		"message": fmt.Sprintf("The server failed to start normally due to the following error:\n%v\n\nPlease fix the configuration and restart the server.", t.err),
	}, nil
}

// GetCacheConfig returns the cache configuration for the tool.
func (t *StartupErrorTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}
