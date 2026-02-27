// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// createHelperProgram creates a Go program that prints 'a' n times to stdout.
func createHelperProgram(t *testing.T) string {
	dir := t.TempDir()
	path := filepath.Join(dir, "printer.go")
	content := `package main
import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		return
	}
	n, _ := strconv.Atoi(os.Args[1])
	// Print in chunks to avoid memory issues with huge strings
	chunkSize := 1024 * 1024
	chunk := strings.Repeat("a", chunkSize)

	for n > 0 {
		if n >= chunkSize {
			fmt.Print(chunk)
			n -= chunkSize
		} else {
			fmt.Print(strings.Repeat("a", n))
			n = 0
		}
	}
}
`
	err := os.WriteFile(path, []byte(content), 0600)
	assert.NoError(t, err)
	return path
}

func TestLocalCommandTool_Execute_LargeOutput(t *testing.T) {
	helperPath := createHelperProgram(t)
	size := consts.DefaultMaxCommandOutputBytes

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
		Command: proto.String("go"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"run", helperPath, fmt.Sprintf("%d", size)},
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

		if !assert.Equal(t, size, len(stdout)) {
			t.Logf("Stdout length mismatch. Expected: %d, Actual: %d. ExitCode: %d, Stderr: %s", size, len(stdout), exitCode, stderr)
		}
		t.Logf("Execution took %v, output size: %d", time.Since(start), len(stdout))
	}
}

func TestLocalCommandTool_Execute_LargeOutput_Truncated(t *testing.T) {
	os.Setenv("MCPANY_MAX_COMMAND_OUTPUT_SIZE", "1024")
	defer os.Unsetenv("MCPANY_MAX_COMMAND_OUTPUT_SIZE")

	helperPath := createHelperProgram(t)
	targetSize := 2048

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
		Command: proto.String("go"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"run", helperPath, fmt.Sprintf("%d", targetSize)},
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
