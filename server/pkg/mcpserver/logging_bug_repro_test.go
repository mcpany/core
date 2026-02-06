package mcpserver_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/worker"
	bus_pb "github.com/mcpany/core/proto/bus"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestServer_CallTool_LogsArgsRedacted(t *testing.T) {
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
	factory := factory.NewUpstreamServiceFactory(poolManager, nil)
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

	upstreamWorker := worker.NewUpstreamWorker(busProvider, toolManager)
	upstreamWorker.Start(ctx)

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	tm := server.ToolManager().(*tool.Manager)

	// Add test tool
	successTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String("sensitive-tool"),
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
	}
	_ = tm.AddTool(successTool)

	// Create client-server connection
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Connect server and client
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// Call the tool with sensitive data
	sanitizedToolName, _ := util.SanitizeToolName("sensitive-tool")
	toolID := "test-service" + "." + sanitizedToolName

	sensitiveValue := "super_secret_key_12345"
	_, err = clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: toolID,
		Arguments: map[string]interface{}{
			"api_key": sensitiveValue,
		},
	})
	require.NoError(t, err)

	// Check logs
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "Calling tool...")

	// We expect the argument key to be logged (proving we are logging arguments)
	assert.Contains(t, logOutput, "api_key", "Argument key should be logged")

	// We expect the sensitive value to be REDACTED
	assert.NotContains(t, logOutput, sensitiveValue, "Sensitive value should be redacted")
}
