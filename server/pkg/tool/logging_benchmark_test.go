// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/pool"
	"google.golang.org/protobuf/proto"
)

func BenchmarkHTTPToolExecute_LoggingOverhead(b *testing.B) {
	// This benchmark measures the overhead of logging when debug is disabled.
	// Reset logger and set to INFO to disable DEBUG logs
	logging.ForTestsOnlyResetLogger()
	logging.Init(slog.LevelInfo, io.Discard)

	// Create a large JSON input
	largeInput := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		largeInput[fmt.Sprintf("key-%d", i)] = "some value to make this string longer and longer and longer"
	}
	inputBytes, _ := json.Marshal(largeInput)
	req := &ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: json.RawMessage(inputBytes),
	}

	poolManager := pool.NewManager()
	// We don't register anything, so it will return "no http pool found" error immediately after logging.

	toolProto := v1.Tool_builder{
		UnderlyingMethodFqn: proto.String("GET http://example.com"),
	}.Build()

	httpTool := NewHTTPTool(toolProto, poolManager, "service-id", nil, &configv1.HttpCallDefinition{}, nil, nil, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = httpTool.Execute(context.Background(), req)
	}
}
