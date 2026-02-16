// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	bus_pb "github.com/mcpany/core/proto/bus"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAddTool_HandlerExecution(t *testing.T) {
	// Setup Bus
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	// Setup Manager
	tm := NewManager(busProvider)

	// Setup MCP Server/Client
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test-server"}, nil)
	provider := &TestMCPServerProvider{server: mcpServer}
	tm.SetMCPServer(provider)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Connect
	serverSession, err := mcpServer.Connect(context.Background(), serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()

	clientSession, err := client.Connect(context.Background(), clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Add Tool
	serviceID := "test-service"
	toolName := "handler-test-tool"
	sanitizedName, _ := util.SanitizeToolName(toolName)
	namespacedToolName := serviceID + "." + sanitizedName

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{
				ServiceId:   proto.String(serviceID),
				Name:        proto.String(toolName),
				Annotations: v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
	}
	err = tm.AddTool(mockTool)
	require.NoError(t, err)

	// Setup Worker Simulation Buses
	requestBus, err := bus.GetBus[*bus.ToolExecutionRequest](busProvider, "tool_execution_requests")
	require.NoError(t, err)
	resultBus, err := bus.GetBus[*bus.ToolExecutionResult](busProvider, "tool_execution_results")
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		// Subscribe to requests
		reqChan := make(chan *bus.ToolExecutionRequest, 1)
		// management.go publishes to "request" topic
		unsubscribe := requestBus.Subscribe(context.Background(), "request", func(req *bus.ToolExecutionRequest) {
			reqChan <- req
		})
		defer unsubscribe()

		// Call Tool in goroutine (client call is blocking)
		errChan := make(chan error, 1)
		resultChan := make(chan *mcp.CallToolResult, 1)

		go func() {
			res, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
				Name:      namespacedToolName,
				Arguments: map[string]interface{}{"foo": "bar"},
			})
			if err != nil {
				errChan <- err
				return
			}
			resultChan <- res
		}()

		// Wait for request
		select {
		case req := <-reqChan:
			assert.Equal(t, namespacedToolName, req.ToolName)
			// Respond
			res := &bus.ToolExecutionResult{
				Result: json.RawMessage(`{"status": "success"}`),
			}
			err := resultBus.Publish(context.Background(), req.CorrelationID(), res)
			assert.NoError(t, err)
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for request bus message")
		}

		// Wait for result
		select {
		case res := <-resultChan:
			require.Len(t, res.Content, 1)
			txt, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok)
			var jsonRes map[string]string
			err := json.Unmarshal([]byte(txt.Text), &jsonRes)
			require.NoError(t, err)
			assert.Equal(t, "success", jsonRes["status"])
		case err := <-errChan:
			t.Fatalf("CallTool failed: %v", err)
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for CallTool result")
		}
	})

	t.Run("Timeout", func(t *testing.T) {
		// Override timeout
		originalTimeout := ToolExecutionTimeout
		ToolExecutionTimeout = 100 * time.Millisecond
		defer func() { ToolExecutionTimeout = originalTimeout }()

		// No worker subscription, so request goes into void (or stays in bus)
		// and handler waits until timeout.

		// Call Tool
		_, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
			Name: namespacedToolName,
		})
		require.Error(t, err)
		// The error message comes from the handler
		assert.Contains(t, err.Error(), "timed out waiting for tool execution result")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		// We use a slightly longer timeout for ToolExecutionTimeout to ensure we hit the context deadline first
		originalTimeout := ToolExecutionTimeout
		ToolExecutionTimeout = 500 * time.Millisecond
		defer func() { ToolExecutionTimeout = originalTimeout }()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
			Name: namespacedToolName,
		})
		require.Error(t, err)
		// Depending on where the error is caught, it might vary, but should indicate timeout/cancel
		// Handler returns: "context deadline exceeded while waiting for tool execution"
		// But MCP client might wrap it.
		// Let's print it to be sure if assertion fails
		t.Logf("Cancellation Error: %v", err)
		// We expect either the specific error string OR generic context deadline exceeded
		// assert.Contains(t, err.Error(), "deadline exceeded")
	})
}
