// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
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

func TestServer_CallTool_HugeEscapedKey_Redacted(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping huge key E2E test in short mode")
	}

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
	hugeKeyTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String("huge-key-tool"),
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
	_ = tm.AddTool(hugeKeyTool)

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

	// Prepare huge escaped key
	// 3MB (safe for test, > 2MB limit)
	targetLen := 3*1024*1024 + 100
	escapedPassword := `p\u0061ssword`
	paddingLen := targetLen - len(escapedPassword)
	padding := strings.Repeat("a", paddingLen)
	hugeKey := padding + escapedPassword
	secretValue := "EXTREMELY_SECRET_VALUE_DO_NOT_LEAK"

	sanitizedToolName, _ := util.SanitizeToolName("huge-key-tool")
	toolID := "test-service" + "." + sanitizedToolName

	// We cannot pass hugeKey as key in map[string]interface{} easily because
	// when we marshal it to JSON to send to server, it will be escaped again?
	// No, map key is string. JSON encoder escapes special chars.
	// But `hugeKey` contains `\u0061`.
	// If I put `hugeKey` string into map, Go's json encoder will escape `\` as `\\`.
	// So `p\\u0061ssword`.
	// The server receives `p\\u0061ssword`.
	// `redactJSONFast` sees `\\` (escaped backslash).
	// It's still an escape sequence.

	// Wait. The bug is about keys containing escapes.
	// If I send `{"p\u0061ssword": "val"}`.
	// Go client: `Arguments: map[string]any{"p\u0061ssword": "val"}`?
	// No, Go string "p\u0061ssword" IS "password" (interpreted at compile time).
	// I need a raw string literal or explicit escape.
	// `key := "p\\u0061ssword"` -> contains literal backslash.

	// If I use `hugeKey` as constructed above (literal backslash), then `CallToolParams` arguments map has a key with backslash.
	// When serialized to JSON by client: `"p\\u0061ssword"`.
	// The JSON text has double backslash.

	// `redactJSONFast` unescapes double backslash to single backslash.
	// So `keyToCheck` (raw) is `p\u0061ssword`.
	// `scanForSensitiveKeys` checks `p\u0061ssword` against `password`.
	// Mismatch.

	// So this setup correctly reproduces the scenario.

	_, err = clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: toolID,
		Arguments: map[string]interface{}{
			hugeKey: secretValue,
		},
	})
	require.NoError(t, err)

	// Check logs
	logOutput := logBuffer.String()

	// Verify truncation/handling didn't panic or error
	assert.Contains(t, logOutput, "Calling tool...")

	// Verify LEAK PREVENTION
	// The log should NOT contain the secret value
	if strings.Contains(logOutput, secretValue) {
		t.Fatalf("Log contains secret value! Redaction failed for huge key.")
	}

	// Verify [REDACTED] is present
	// Note: since the key is huge, the log message might be huge.
	// But we check that the value part is redacted.
	assert.Contains(t, logOutput, "[REDACTED]")
}
