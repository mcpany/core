// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/util"
	mcp_sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestToolManager_MCPHandler_Success(t *testing.T) {
	messageBus, _ := bus.NewProvider(nil)
	tm := NewManager(messageBus)

	mcpServer := mcp_sdk.NewServer(&mcp_sdk.Implementation{Name: "test-server", Version: "1.0"}, nil)
	provider := &TestMCPServerProvider{server: mcpServer}
	tm.SetMCPServer(provider)

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	toolName := "success-tool"
	serviceID := "test-service"
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId:   proto.String(serviceID),
				Name:        proto.String(toolName),
				Annotations: mcp_router_v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
	}
	err := tm.AddTool(mockTool)
	require.NoError(t, err)

	// Setup Worker to listen and respond
	reqBus, err := bus.GetBus[*bus.ToolExecutionRequest](messageBus, "tool_execution_requests")
	require.NoError(t, err)
	resBus, err := bus.GetBus[*bus.ToolExecutionResult](messageBus, "tool_execution_results")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		// Simple worker loop
		unsub := reqBus.Subscribe(ctx, "request", func(req *bus.ToolExecutionRequest) {
			if req.ToolName == serviceID+"."+toolName {
				resBytes, _ := json.Marshal(map[string]string{"status": "ok"})
				res := &bus.ToolExecutionResult{
					Result: json.RawMessage(resBytes),
				}
				res.SetCorrelationID(req.CorrelationID())
				// Publish to CorrelationID topic as per management.go expectation
				_ = resBus.Publish(ctx, req.CorrelationID(), res)
			}
		})
		defer unsub()
		<-ctx.Done()
	}()

	// Setup Client
	clientTransport, serverTransport := mcp_sdk.NewInMemoryTransports()

	// Start server session
	go func() {
		serverSession, err := mcpServer.Connect(ctx, serverTransport, nil)
		if err == nil {
			defer serverSession.Close()
			<-ctx.Done()
		}
	}()

	client := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "test-client"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Call Tool
	sanitizedName, _ := util.SanitizeToolName(toolName)
	resp, err := clientSession.CallTool(ctx, &mcp_sdk.CallToolParams{
		Name: serviceID + "." + sanitizedName,
	})
	require.NoError(t, err)
	require.Len(t, resp.Content, 1)
	txt, ok := resp.Content[0].(*mcp_sdk.TextContent)
	require.True(t, ok)
	assert.Contains(t, txt.Text, "ok")

	cancel() // Stop worker
}

func TestToolManager_MCPHandler_BusError(t *testing.T) {
	// Hook to fail GetBus
	origHook := bus.GetBusHook
	defer func() { bus.GetBusHook = origHook }()

	// Re-define hook to return error for requests
	bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
		if topic == "tool_execution_requests" {
			return nil, errors.New("simulated bus error")
		}
		return nil, nil
	}

	messageBus, _ := bus.NewProvider(nil)
	tm := NewManager(messageBus)
	mcpServer := mcp_sdk.NewServer(&mcp_sdk.Implementation{Name: "test"}, nil)
	tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	toolName := "bus-error-tool"
	serviceID := "test-service"
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId:   proto.String(serviceID),
				Name:        proto.String(toolName),
				Annotations: mcp_router_v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
	}
	err := tm.AddTool(mockTool)
	require.NoError(t, err)

	clientTransport, serverTransport := mcp_sdk.NewInMemoryTransports()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		serverSession, _ := mcpServer.Connect(ctx, serverTransport, nil)
		if serverSession != nil {
			defer serverSession.Close()
			<-ctx.Done()
		}
	}()

	client := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "client"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	_, err = clientSession.CallTool(ctx, &mcp_sdk.CallToolParams{
		Name: serviceID + "." + toolName,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "simulated bus error")
}

func TestToolManager_MCPHandler_Timeout(t *testing.T) {
	// Set timeout to small value
	origTimeout := ToolExecutionTimeout
	ToolExecutionTimeout = 10 * time.Millisecond
	defer func() { ToolExecutionTimeout = origTimeout }()

	messageBus, _ := bus.NewProvider(nil)
	tm := NewManager(messageBus)
	mcpServer := mcp_sdk.NewServer(&mcp_sdk.Implementation{Name: "test"}, nil)
	tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	toolName := "timeout-tool"
	serviceID := "test-service"
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId:   proto.String(serviceID),
				Name:        proto.String(toolName),
				Annotations: mcp_router_v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
	}
	err := tm.AddTool(mockTool)
	require.NoError(t, err)

	clientTransport, serverTransport := mcp_sdk.NewInMemoryTransports()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		serverSession, _ := mcpServer.Connect(ctx, serverTransport, nil)
		if serverSession != nil {
			defer serverSession.Close()
			<-ctx.Done()
		}
	}()

	client := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "client"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	_, err = clientSession.CallTool(ctx, &mcp_sdk.CallToolParams{
		Name: serviceID + "." + toolName,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out waiting for tool execution")
}

func TestToolManager_MCPHandler_ExecutionError(t *testing.T) {
	messageBus, _ := bus.NewProvider(nil)
	tm := NewManager(messageBus)
	mcpServer := mcp_sdk.NewServer(&mcp_sdk.Implementation{Name: "test"}, nil)
	tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	toolName := "error-tool"
	serviceID := "test-service"
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId:   proto.String(serviceID),
				Name:        proto.String(toolName),
				Annotations: mcp_router_v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
	}
	err := tm.AddTool(mockTool)
	require.NoError(t, err)

	reqBus, _ := bus.GetBus[*bus.ToolExecutionRequest](messageBus, "tool_execution_requests")
	resBus, _ := bus.GetBus[*bus.ToolExecutionResult](messageBus, "tool_execution_results")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		reqBus.Subscribe(ctx, "request", func(req *bus.ToolExecutionRequest) {
			res := &bus.ToolExecutionResult{
				Error: errors.New("worker execution failed"),
			}
			res.SetCorrelationID(req.CorrelationID())
			_ = resBus.Publish(ctx, req.CorrelationID(), res)
		})
		<-ctx.Done()
	}()

	clientTransport, serverTransport := mcp_sdk.NewInMemoryTransports()
	go func() {
		serverSession, _ := mcpServer.Connect(ctx, serverTransport, nil)
		if serverSession != nil {
			defer serverSession.Close()
			<-ctx.Done()
		}
	}()

	client := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "client"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	_, err = clientSession.CallTool(ctx, &mcp_sdk.CallToolParams{
		Name: serviceID + "." + toolName,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "worker execution failed")
}
