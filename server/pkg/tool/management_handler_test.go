// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	bus_pb "github.com/mcpany/core/proto/bus"
	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// setupHandlerTest initializes the environment for testing the MCP server handler.
// It returns the ToolManager, BusProvider, MCP Client Session, and a cleanup function.
func setupHandlerTest(t *testing.T) (*Manager, *bus.Provider, *mcp.ClientSession, func()) {
	// Initialize Bus
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	// Initialize Manager
	tm := NewManager(busProvider)

	// Initialize MCP Server
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "1.0.0"}, nil)
	provider := &TestMCPServerProvider{server: mcpServer}
	tm.SetMCPServer(provider)

	// Initialize MCP Client
	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)

	// Connect Client and Server using In-Memory Transport
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	ctx := context.Background()

	// Start Server Session
	serverSession, err := mcpServer.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)

	// Start Client Session
	clientSession, err := mcpClient.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)

	cleanup := func() {
		_ = serverSession.Close()
		_ = clientSession.Close()
	}

	return tm, busProvider, clientSession, cleanup
}

// mockWorker simulates the backend worker that processes tool execution requests.
// It subscribes to the request bus and publishes a result to the result bus.
//
// Parameters:
//   - t: Testing object.
//   - busProvider: The bus provider to use.
//   - delay: How long to wait before sending the response.
//   - resultData: The data to return as the result (if successful).
//   - resultError: The error string to return (if simulating an execution error).
func mockWorker(t *testing.T, busProvider *bus.Provider, delay time.Duration, resultData any, resultError error) {
	ctx := context.Background()
	reqBus, err := bus.GetBus[*bus.ToolExecutionRequest](busProvider, "tool_execution_requests")
	require.NoError(t, err)

	resBus, err := bus.GetBus[*bus.ToolExecutionResult](busProvider, "tool_execution_results")
	require.NoError(t, err)

	// Subscribe to requests
	unsubscribe := reqBus.Subscribe(ctx, "request", func(req *bus.ToolExecutionRequest) {
		// Simulate processing time
		time.Sleep(delay)

		result := &bus.ToolExecutionResult{}
		if resultError != nil {
			result.Error = resultError
		} else {
			// Marshal the result data
			bytes, err := json.Marshal(resultData)
			require.NoError(t, err)
			result.Result = bytes
		}

		// Publish result using the correlation ID from the request
		err := resBus.Publish(ctx, req.CorrelationID(), result)
		// We don't assert error here because in timeout tests, the subscription might be gone
		if err != nil {
			t.Logf("Worker publish error (expected in timeout tests): %v", err)
		}
	})

	// Run in background and cleanup when test ends?
	// No, this is a fire-and-forget worker for the duration of the test.
	// But we should unsubscribe to avoid leaking.
	t.Cleanup(func() {
		unsubscribe()
	})
}

func TestHandler_Success(t *testing.T) {
	tm, busProvider, client, cleanup := setupHandlerTest(t)
	defer cleanup()

	// 1. Add Tool
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	toolName := "success-tool"
	serviceID := "test-service"
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String(serviceID),
				Name:      proto.String(toolName),
				Annotations: mcp_router_v1.ToolAnnotations_builder{
					InputSchema: inputSchema,
				}.Build(),
			}.Build()
		},
	}
	err := tm.AddTool(mockTool)
	require.NoError(t, err)

	// 2. Start Mock Worker
	expectedData := map[string]interface{}{"status": "ok", "value": 42.0} // 42.0 because JSON unmarshals numbers as float64
	mockWorker(t, busProvider, 10*time.Millisecond, expectedData, nil)

	// 3. Call Tool via Client
	sanitizedName, _ := util.SanitizeToolName(toolName)
	fullToolName := serviceID + "." + sanitizedName

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name: fullToolName,
		Arguments: map[string]interface{}{"foo": "bar"},
	})
	require.NoError(t, err)

	// 4. Verify Result
	assert.False(t, result.IsError)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)

	var actualData map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &actualData)
	require.NoError(t, err)

	assert.Equal(t, expectedData["status"], actualData["status"])
	assert.Equal(t, expectedData["value"], actualData["value"])
}

func TestHandler_Timeout(t *testing.T) {
	// Override timeout for this test
	originalTimeout := toolExecutionTimeout
	toolExecutionTimeout = 50 * time.Millisecond
	defer func() { toolExecutionTimeout = originalTimeout }()

	tm, busProvider, client, cleanup := setupHandlerTest(t)
	defer cleanup()

	// 1. Add Tool
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	toolName := "timeout-tool"
	serviceID := "test-service"
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String(serviceID),
				Name:      proto.String(toolName),
				Annotations: mcp_router_v1.ToolAnnotations_builder{
					InputSchema: inputSchema,
				}.Build(),
			}.Build()
		},
	}
	err := tm.AddTool(mockTool)
	require.NoError(t, err)

	// 2. Start Mock Worker with DELAY > timeout
	mockWorker(t, busProvider, 200*time.Millisecond, map[string]string{"status": "ok"}, nil)

	// 3. Call Tool via Client
	sanitizedName, _ := util.SanitizeToolName(toolName)
	fullToolName := serviceID + "." + sanitizedName

	// Ensure context lives longer than timeout to test the *internal* timeout logic
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = client.CallTool(ctx, &mcp.CallToolParams{
		Name: fullToolName,
		Arguments: map[string]interface{}{"foo": "bar"},
	})

	// 4. Verify Timeout Error
	require.Error(t, err)
	// The error comes from mcp-sdk which wraps the error returned by handler.
	// The handler returns: fmt.Errorf("timed out waiting for tool execution result for tool %s", req.Params.Name)
	assert.Contains(t, err.Error(), "timed out waiting for tool execution result")
}

func TestHandler_ExecutionError(t *testing.T) {
	tm, busProvider, client, cleanup := setupHandlerTest(t)
	defer cleanup()

	// 1. Add Tool
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	toolName := "error-tool"
	serviceID := "test-service"
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String(serviceID),
				Name:      proto.String(toolName),
				Annotations: mcp_router_v1.ToolAnnotations_builder{
					InputSchema: inputSchema,
				}.Build(),
			}.Build()
		},
	}
	err := tm.AddTool(mockTool)
	require.NoError(t, err)

	// 2. Start Mock Worker with Error
	expectedErrorMsg := "simulated execution failure"
	mockWorker(t, busProvider, 10*time.Millisecond, nil, errors.New(expectedErrorMsg))

	// 3. Call Tool via Client
	sanitizedName, _ := util.SanitizeToolName(toolName)
	fullToolName := serviceID + "." + sanitizedName

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = client.CallTool(ctx, &mcp.CallToolParams{
		Name: fullToolName,
		Arguments: map[string]interface{}{"foo": "bar"},
	})

	// 4. Verify Error
	require.Error(t, err)
	assert.Contains(t, err.Error(), expectedErrorMsg)
}
