// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/testutil"
	"github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestTracingIntegration(t *testing.T) {
	// 1. Setup InMemory Tracing
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	// 2. Setup Server Components
	busProvider, _ := bus.NewProvider(nil)
	toolManager := tool.NewManager(busProvider)

	// Add Tracing Middleware!
	toolManager.AddMiddleware(middleware.NewTracingMiddleware())

	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(nil, toolManager, promptManager, resourceManager, authManager)

	server, err := mcpserver.NewServer(
		context.Background(),
		toolManager,
		promptManager,
		resourceManager,
		authManager,
		serviceRegistry,
		busProvider,
		true,
	)
	require.NoError(t, err)

	// 3. Register a Mock Tool
	toolName := "mock-tool"
	serviceID := "test-service"
	displayName := "Mock Tool"
	desc := "A mock tool for testing"
	mockTool := &testutil.MockTool{
		ToolDef: &v1.Tool{
			Name:        &toolName,
			DisplayName: &displayName,
			Description: &desc,
			ServiceId:   &serviceID,
			InputSchema: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"type": structpb.NewStringValue("object"),
				},
			},
		},
		ExecuteFunc: func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return map[string]string{"status": "ok"}, nil
		},
	}
	err = toolManager.AddTool(mockTool)
	require.NoError(t, err)

	// 4. Create MCP Client connected to Server
	// We can use direct call to CallTool via Server instance since we want to verify middleware execution.
	// But `ExecuteTool` is called by `CallTool` in `server.go`.

	// Let's use `server.CallTool` directly to simulate a request processing flow that hits the manager.
	// Note: `server.CallTool` handles MCP wrapping, but `toolManager.ExecuteTool` is where middleware lives.
	// `server.CallTool` calls `toolManager.ExecuteTool`.

	req := &tool.ExecutionRequest{
		ToolName:   toolName,
		ToolInputs: []byte(`{}`),
	}

	// Execute
	_, err = server.CallTool(context.Background(), req)
	require.NoError(t, err)

	// 5. Verify Spans
	spans := exporter.GetSpans()
	assert.NotEmpty(t, spans, "Should have captured at least one span")

	found := false
	for _, span := range spans {
		if span.Name == "tool.execute" {
			found = true

			// Verify Attributes
			attrs := span.Attributes
			attrMap := make(map[string]any)
			for _, a := range attrs {
				attrMap[string(a.Key)] = a.Value.AsInterface()
			}

			assert.Equal(t, toolName, attrMap["tool.name"])
			assert.Equal(t, serviceID, attrMap["service.id"])
			break
		}
	}
	assert.True(t, found, "Did not find 'tool.execute' span")
}
