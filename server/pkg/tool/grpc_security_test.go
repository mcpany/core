// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	weatherpb "github.com/mcpany/core/proto/examples/weather/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGRPCTool_InformationLeakage(t *testing.T) {
	t.Parallel()

	methodDesc := findMethodDescriptor(t, "WeatherService", "GetWeather")
	toolProto := &v1.Tool{}
	toolProto.SetUnderlyingMethodFqn(string(methodDesc.FullName()))

	// Mock a sensitive error
	// We use a DSN to verify RedactDSN.
	// We use JSON to verify RedactJSON.
	// We use a long string to verify Truncation.
	longStackTrace := strings.Repeat("at /app/stack.go:10\n", 50) // 50 * ~20 chars = 1000 chars
	sensitiveErrorMsg := `upstream connect failed: postgres://user:supersecret@10.0.0.1:5432/db. Payload: {"password": "jsonSecret"}. Stack: ` + longStackTrace

	server := &mockWeatherServer{
		getWeatherFunc: func(_ context.Context, req *weatherpb.GetWeatherRequest) (*weatherpb.GetWeatherResponse, error) {
			return nil, errors.New(sensitiveErrorMsg)
		},
	}
	conn := setupGrpcTest(t, server)
	wrapper := client.NewGrpcClientWrapper(conn, nil, nil)

	pm := pool.NewManager()
	mockPool := &mockGrpcPool{
		getFunc: func(_ context.Context) (*client.GrpcClientWrapper, error) {
			return wrapper, nil
		},
	}
	pm.Register("grpc-test", mockPool)

	grpcTool := tool.NewGRPCTool(toolProto, pm, "grpc-test", methodDesc, nil, nil)
	inputs := json.RawMessage(`{"location": "London"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}

	_, err := grpcTool.Execute(context.Background(), req)
	require.Error(t, err)

	errMsg := err.Error()
	t.Logf("Sanitized Error: %s", errMsg)

	// Verification 1: DSN Redaction
	assert.NotContains(t, errMsg, "supersecret", "DSN password should be redacted")
	assert.Contains(t, errMsg, "[REDACTED]", "Should contain redaction placeholder")

	// Verification 2: JSON Redaction
	assert.NotContains(t, errMsg, "jsonSecret", "JSON password should be redacted")

	// Verification 3: Truncation
	assert.True(t, len(errMsg) <= 600, "Error message should be truncated (allow some buffer for prefix)")
	assert.Contains(t, errMsg, "... (truncated)", "Should contain truncation marker")
}
