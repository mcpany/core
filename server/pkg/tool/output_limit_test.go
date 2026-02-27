// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_Execute_LargeOutput(t *testing.T) {
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"properties": map[string]interface{}{
			"args": map[string]interface{}{},
		},
	})
	toolDef := mcp_router_v1.Tool_builder{
		Name:        proto.String("test-tool-large"),
		InputSchema: inputSchema,
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("dd"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"if=/dev/zero", "bs=10485760", "count=1"},
	}.Build()

	localTool := NewLocalCommandTool(toolDef, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName:  "test-tool-large",
		Arguments: map[string]interface{}{},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	start := time.Now()
	result, err := localTool.Execute(context.Background(), req)
	if !assert.NoError(t, err) {
		t.Logf("Execute failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if assert.True(t, ok, "result should be a map") {
		stdout, _ := resultMap["stdout"].(string)
		stderr, _ := resultMap["stderr"].(string)
		exitCode, _ := resultMap["return_code"].(int)

		if !assert.Equal(t, consts.DefaultMaxCommandOutputBytes, len(stdout)) {
			t.Logf("Stdout length mismatch. ExitCode: %d, Stderr: %s", exitCode, stderr)
		}
		t.Logf("Execution took %v, output size: %d", time.Since(start), len(stdout))
	}
}

func TestLocalCommandTool_Execute_LargeOutput_Truncated(t *testing.T) {
	os.Setenv("MCPANY_MAX_COMMAND_OUTPUT_SIZE", "1024")
	defer os.Unsetenv("MCPANY_MAX_COMMAND_OUTPUT_SIZE")

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"properties": map[string]interface{}{
			"args": map[string]interface{}{},
		},
	})
	toolDef := mcp_router_v1.Tool_builder{
		Name:        proto.String("test-tool-large-truncated"),
		InputSchema: inputSchema,
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("dd"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"if=/dev/zero", "bs=2048", "count=1"},
	}.Build()

	localTool := NewLocalCommandTool(toolDef, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName:  "test-tool-large-truncated",
		Arguments: map[string]interface{}{},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	if !assert.NoError(t, err) {
		t.Logf("Execute failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if assert.True(t, ok, "result should be a map") {
		stdout, _ := resultMap["stdout"].(string)
		stderr, _ := resultMap["stderr"].(string)
		exitCode, _ := resultMap["return_code"].(int)

		if !assert.Equal(t, 1024, len(stdout)) {
			t.Logf("Stdout length mismatch. ExitCode: %d, Stderr: %s", exitCode, stderr)
		}
	}
}
