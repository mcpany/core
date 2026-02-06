// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type mockExecutor struct {
	capturedArgs []string
}

func (m *mockExecutor) Execute(ctx context.Context, cmd string, args []string, workingDir string, env []string) (io.ReadCloser, io.ReadCloser, <-chan int, error) {
	m.capturedArgs = args
	ch := make(chan int, 1)
	ch <- 0
	close(ch)
	return io.NopCloser(strings.NewReader("")), io.NopCloser(strings.NewReader("")), ch, nil
}

func (m *mockExecutor) ExecuteWithStdIO(ctx context.Context, cmd string, args []string, workingDir string, env []string) (io.WriteCloser, io.ReadCloser, io.ReadCloser, <-chan int, error) {
	return nil, nil, nil, nil, nil
}

func TestLargeIntPrecisionLoss(t *testing.T) {
	t.Parallel()
	// Large integer that cannot be represented exactly as float64
	// 2^63 - 1 = 9223372036854775807
	// float64 has 53 bits of mantissa.
	largeIntStr := "9223372036854775807"

	mockExec := &mockExecutor{}

	// Define tool
	toolDef := v1.Tool_builder{
		Name: proto.String("test-tool"),
		InputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"type": structpb.NewStringValue("object"),
				"properties": structpb.NewStructValue(&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"id": {Kind: &structpb.Value_StringValue{StringValue: "string"}}, // Treat as string in schema? No, user might send number.
						// Even if schema says number, JSON decoder into interface{} makes it float64.
					},
				}),
			},
		},
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{id}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("id")}.Build(),
			}.Build(),
		},
	}.Build()

	// Create CommandTool manually
	ct := &CommandTool{
		tool:           toolDef,
		service:        service,
		callDefinition: callDef,
		executorFactory: func(_ *configv1.ContainerEnvironment) command.Executor {
			return mockExec
		},
	}

	// Execute
	inputsJSON := `{"id": ` + largeIntStr + `}`
	req := &ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: json.RawMessage([]byte(inputsJSON)),
	}

	_, err := ct.Execute(context.Background(), req)
	require.NoError(t, err)

	require.Len(t, mockExec.capturedArgs, 1)
	captured := mockExec.capturedArgs[0]

	// Check if captured argument matches original string
	assert.Equal(t, largeIntStr, captured, "Large integer should be preserved exactly")
}
