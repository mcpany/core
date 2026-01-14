// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	routerv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestWebsocketTool_Getters_Coverage(t *testing.T) {
	toolProto := &routerv1.Tool{Name: proto.String("test")}
	toolProto.SetServiceId("s1")
	callDef := &configv1.WebsocketCallDefinition{
		Cache: &configv1.CacheConfig{IsEnabled: proto.Bool(true)},
	}

	wsTool := NewWebsocketTool(toolProto, nil, "s1", nil, callDef)

	assert.NotNil(t, wsTool.MCPTool())
	assert.NotNil(t, wsTool.GetCacheConfig())
	assert.True(t, wsTool.GetCacheConfig().GetIsEnabled())
}

func TestGRPCTool_Getters_Coverage(t *testing.T) {
	toolProto := &routerv1.Tool{Name: proto.String("test")}
	toolProto.SetServiceId("s1")

    grpcTool := &GRPCTool{
        tool: toolProto,
        cache: &configv1.CacheConfig{IsEnabled: proto.Bool(true)},
    }

	assert.NotNil(t, grpcTool.MCPTool())
	assert.NotNil(t, grpcTool.GetCacheConfig())
}

func TestLocalCommandTool_Getters_Coverage(t *testing.T) {
	toolProto := &routerv1.Tool{Name: proto.String("test")}
	toolProto.SetServiceId("s1")

    service := &configv1.CommandLineUpstreamService{
        Command: proto.String("echo"),
    }
    callDef := &configv1.CommandLineCallDefinition{
        Cache: &configv1.CacheConfig{IsEnabled: proto.Bool(true)},
        Args: []string{"hello"},
    }

	lct := NewLocalCommandTool(toolProto, service, callDef, nil, "call1")

	assert.NotNil(t, lct.MCPTool())
	assert.NotNil(t, lct.GetCacheConfig())
}
