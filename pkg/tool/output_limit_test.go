// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/consts"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_Execute_LargeOutput(t *testing.T) {
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"args": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := &v1.Tool{
		Name:        proto.String("test-tool-large"),
		InputSchema: inputSchema,
	}
	service := &configv1.CommandLineUpstreamService{}
	// Use python3 to print 10MB of data
	service.Command = proto.String("python3")
	service.Local = proto.Bool(true)
	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"-c", "print('a' * 10 * 1024 * 1024)"},
	}

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName:  "test-tool-large",
		Arguments: map[string]interface{}{},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	start := time.Now()
	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	stdout := resultMap["stdout"].(string)

	// Default limit is 10MB. The command outputs 10MB + 1 byte (newline).
	// So it should be truncated to exactly 10MB.
	assert.Equal(t, consts.DefaultMaxCommandOutputBytes, len(stdout))
	t.Logf("Execution took %v, output size: %d", time.Since(start), len(stdout))
}

func TestLocalCommandTool_Execute_LargeOutput_Truncated(t *testing.T) {
	// Set the env var to limit output to 1KB
	os.Setenv("MCPANY_MAX_COMMAND_OUTPUT_SIZE", "1024")
	defer os.Unsetenv("MCPANY_MAX_COMMAND_OUTPUT_SIZE")

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"args": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := &v1.Tool{
		Name:        proto.String("test-tool-large-truncated"),
		InputSchema: inputSchema,
	}
	service := &configv1.CommandLineUpstreamService{}
	service.Command = proto.String("python3")
	service.Local = proto.Bool(true)
	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"-c", "print('a' * 2048)"},
	}

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName:  "test-tool-large-truncated",
		Arguments: map[string]interface{}{},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	stdout := resultMap["stdout"].(string)

	// Should be truncated to 1024
	assert.Equal(t, 1024, len(stdout))
}
