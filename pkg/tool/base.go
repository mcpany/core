// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package tool defines the interface for tools that can be executed by the upstream service.
package tool

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

type baseTool struct {
	tool          *v1.Tool
	serviceConfig *configv1.UpstreamServiceConfig
	callable      Callable
}

func newBaseTool(toolDef *configv1.ToolDefinition, serviceConfig *configv1.UpstreamServiceConfig, callable Callable, inputSchema, outputSchema *structpb.Struct) (*baseTool, error) {
	pbTool, err := ConvertToolDefinitionToProto(toolDef, inputSchema, outputSchema)
	if err != nil {
		return nil, err
	}
	return &baseTool{
		tool:          pbTool,
		serviceConfig: serviceConfig,
		callable:      callable,
	}, nil
}

func (t *baseTool) Tool() *v1.Tool {
	return t.tool
}

func (t *baseTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}
