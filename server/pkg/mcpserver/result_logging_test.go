// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/mcpserver"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream/factory"
	"github.com/mcpany/core/pkg/util"
	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type flexibleMockTool struct {
	tool *v1.Tool
	executeFunc func(context.Context, *tool.ExecutionRequest) (any, error)
}

func (m *flexibleMockTool) Tool() *v1.Tool {
	return m.tool
}

func (m *flexibleMockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, req)
	}
	return "success", nil
}

func (m *flexibleMockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func (m *flexibleMockTool) MCPTool() *mcp.Tool {
	t, _ := tool.ConvertProtoToMCPTool(m.tool)
	return t
}

func TestServer_CallTool_ResultLogging(t *testing.T) {
	// Reset logger to capture output
	logging.ForTestsOnlyResetLogger()
	var logBuffer bytes.Buffer
	logging.Init(slog.LevelInfo, &logBuffer, "text")

	// Restore logger after test
	defer func() {
		logging.ForTestsOnlyResetLogger()
		logging.GetLogger() // Re-init default
	}()

	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	tm := server.ToolManager().(*tool.Manager)

	sensitiveValue := "sensitive_token_XYZ"
	toolName := "sensitive-result-tool"

	ft := &flexibleMockTool{
		tool: v1.Tool_builder{
			Name:      proto.String(toolName),
			ServiceId: proto.String("test-service"),
			Annotations: v1.ToolAnnotations_builder{
				InputSchema: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"type":       structpb.NewStringValue("object"),
						"properties": structpb.NewStructValue(&structpb.Struct{}),
					},
				},
			}.Build(),
		}.Build(),
		executeFunc: func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return map[string]interface{}{
				"api_key": sensitiveValue,
				"other": "value",
			}, nil
		},
	}
	_ = tm.AddTool(ft)

	// Create client-server connection
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	sanitizedToolName, _ := util.SanitizeToolName(toolName)
	toolID := "test-service" + "." + sanitizedToolName

	_, err = clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: toolID,
	})
	require.NoError(t, err)

	logOutput := logBuffer.String()
	// Verification
	assert.Contains(t, logOutput, "Tool execution completed")
	assert.NotContains(t, logOutput, sensitiveValue, "Sensitive result value should be redacted in logs")
	assert.Contains(t, logOutput, "[REDACTED]", "Logs should contain redacted placeholder")
}
