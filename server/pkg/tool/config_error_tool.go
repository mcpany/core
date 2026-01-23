// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// ConfigErrorTool implements the Tool interface for reporting configuration errors.
type ConfigErrorTool struct {
	errorMessage string
}

// NewConfigErrorTool creates a new ConfigErrorTool with the given error message.
func NewConfigErrorTool(errorMessage string) *ConfigErrorTool {
	return &ConfigErrorTool{errorMessage: errorMessage}
}

// Tool returns the protobuf definition of the tool.
func (t *ConfigErrorTool) Tool() *v1.Tool {
	schema, _ := structpb.NewStruct(map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	})

	return &v1.Tool{
		Name:        proto.String("mcp_config_error"),
		Description: proto.String("Returns the configuration errors that prevented the server from starting correctly. Use this tool to diagnose startup issues."),
		ServiceId:   proto.String("system"),
		Annotations: &v1.ToolAnnotations{
			InputSchema: schema,
		},
	}
}

// MCPTool returns the MCP tool definition.
func (t *ConfigErrorTool) MCPTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "mcp_config_error",
		Description: "Returns the configuration errors that prevented the server from starting correctly. Use this tool to diagnose startup issues.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	}
}

// Execute runs the tool and returns the error message.
func (t *ConfigErrorTool) Execute(_ context.Context, _ *ExecutionRequest) (any, error) {
	return map[string]string{
		"error":  t.errorMessage,
		"status": "Configuration Failed",
		"action": "Please check the server logs or the configuration file syntax.",
	}, nil
}

// GetCacheConfig returns nil as this tool should not be cached.
func (t *ConfigErrorTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}
