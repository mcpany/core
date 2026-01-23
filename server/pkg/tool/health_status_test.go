// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestUnhealthyToolsAreVisibleWithWarning(t *testing.T) {
	busProvider, _ := bus.NewProvider(nil)
	tm := NewManager(busProvider)

	serviceID := "svc1"
	toolName := "my_tool"
	errorMsg := "connection refused"

	// 1. Add Tool
	tTool := &v1.Tool{
		Name:        proto.String(toolName),
		ServiceId:   proto.String(serviceID),
		Description: proto.String("A cool tool"),
	}

	tool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return tTool
		},
		MCPToolFunc: func() *mcp.Tool {
			return &mcp.Tool{
				Name:        serviceID + "." + toolName,
				Description: "A cool tool",
			}
		},
	}

	err := tm.AddTool(tool)
	assert.NoError(t, err)

	// 2. Mark Service Unhealthy via UpdateServiceHealth (simulating registry)
	// We first need to add ServiceInfo so it exists
	tm.AddServiceInfo(serviceID, &ServiceInfo{
		Name:   serviceID,
		Config: &configv1.UpstreamServiceConfig{Name: proto.String(serviceID)},
	})
	tm.UpdateServiceHealth(serviceID, false, errorMsg)

	// 3. List MCP Tools (where modification happens)
	tools := tm.ListMCPTools()

	found := false
	for _, toolItem := range tools {
		if toolItem.Name == serviceID+"."+toolName {
			found = true
			assert.Contains(t, toolItem.Description, "[⚠️ UNHEALTHY: connection refused]")
			assert.Contains(t, toolItem.Description, "A cool tool")
			break
		}
	}
	assert.True(t, found, "Tool should be visible even when service is unhealthy")

	// 4. Verify Execution Blocked
	_, err = tm.ExecuteTool(context.Background(), &ExecutionRequest{
		ToolName: serviceID + "." + toolName,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service svc1 is currently unhealthy")
}
