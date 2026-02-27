// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// createLargeFile creates a temporary file with 'a' repeated size times.
func createLargeFile(t *testing.T, size int) string {
	dir := t.TempDir()
	path := filepath.Join(dir, "large_file.txt")

	f, err := os.Create(path)
	assert.NoError(t, err)
	defer f.Close()

	// Write in chunks
	chunkSize := 1024 * 1024
	chunk := []byte(strings.Repeat("a", chunkSize))

	remaining := size
	for remaining > 0 {
		if remaining >= chunkSize {
			_, err := f.Write(chunk)
			assert.NoError(t, err)
			remaining -= chunkSize
		} else {
			_, err := f.Write([]byte(strings.Repeat("a", remaining)))
			assert.NoError(t, err)
			remaining = 0
		}
	}
	return path
}

func TestLocalCommandTool_Execute_LargeOutput(t *testing.T) {
	size := consts.DefaultMaxCommandOutputBytes
	filePath := createLargeFile(t, size)

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"properties": map[string]interface{}{
			"args": map[string]interface{}{},
		},
	})
	toolDef := mcp_router_v1.Tool_builder{
		Name:        proto.String("test-tool-large"),
		InputSchema: inputSchema,
	}.Build()

	cmdName := "cat"
	if runtime.GOOS == "windows" {
		cmdName = "type"
	}

	// Resolve full path to avoid PATH issues in restricted environment
	if path, err := exec.LookPath(cmdName); err == nil {
		cmdName = path
	}

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String(cmdName),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{filePath},
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

	// Create a file larger than the limit
	targetSize := 2048
	filePath := createLargeFile(t, targetSize)

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"properties": map[string]interface{}{
			"args": map[string]interface{}{},
		},
	})
	toolDef := mcp_router_v1.Tool_builder{
		Name:        proto.String("test-tool-large-truncated"),
		InputSchema: inputSchema,
	}.Build()

	cmdName := "cat"
	if runtime.GOOS == "windows" {
		cmdName = "type"
	}

	// Resolve full path to avoid PATH issues in restricted environment
	if path, err := exec.LookPath(cmdName); err == nil {
		cmdName = path
	}

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String(cmdName),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{filePath},
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
