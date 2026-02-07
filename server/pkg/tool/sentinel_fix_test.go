// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestSentinel_Fix_GitSpaceAllowed(t *testing.T) {
	t.Parallel()

	// 1. Setup a tool definition for "git"
	tool := v1.Tool_builder{
		Name:        proto.String("git-test"),
		Description: proto.String("A git wrapper"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"rev-parse", "{{rev}}"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// 2. Input with SPACE
	req := &ExecutionRequest{
		ToolName: "git-test",
		Arguments: map[string]interface{}{
			"rev": "HEAD HEAD", // Space!
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// 3. Execute
	_, err := localTool.Execute(context.Background(), req)

	// 4. Assert: Should NOT be blocked by "shell injection detected"
	// It might fail execution (exit code 128), but not validation error.
	if err != nil {
		assert.False(t, strings.Contains(err.Error(), "shell injection detected"), "Did not expect shell injection error for space in git arg, got: %v", err)
	}
}

func TestSentinel_Fix_BashInjectionBlocked(t *testing.T) {
	t.Parallel()

	// 1. Setup a tool definition for "bash"
	tool := v1.Tool_builder{
		Name:        proto.String("bash-test"),
		Description: proto.String("A bash wrapper"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("bash"), // bash IS a shell
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "{{script}}"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// 2. Input with Injection Chars
	req := &ExecutionRequest{
		ToolName: "bash-test",
		Arguments: map[string]interface{}{
			"script": "echo hello; rm -rf /",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// 3. Execute
	_, err := localTool.Execute(context.Background(), req)

	// 4. Assert: Should BE blocked
	assert.Error(t, err)
	if err != nil {
		assert.True(t, strings.Contains(err.Error(), "shell injection detected"), "Expected shell injection error for bash injection, got: %v", err)
	}
}

func TestSentinel_Fix_BashSpaceBlocked(t *testing.T) {
	t.Parallel()

	// 1. Setup a tool definition for "bash"
	tool := v1.Tool_builder{
		Name:        proto.String("bash-space-test"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("bash"), // bash IS a shell
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "{{script}}"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// 2. Input with Space (Unquoted in config)
	// Because config uses {{script}} (unquoted), spaces are dangerous in shell context if not handled.
	// Although here it is passed as argument to -c.
	// bash -c "echo hello".
	// If input is "echo hello". Arg becomes "echo hello".
	// bash -c "echo hello". Safe.
	// BUT, current policy blocks spaces for shells to be super strict.
	req := &ExecutionRequest{
		ToolName: "bash-space-test",
		Arguments: map[string]interface{}{
			"script": "echo hello",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// 3. Execute
	_, err := localTool.Execute(context.Background(), req)

	// 4. Assert: Should BE blocked (preserving existing strict behavior for shells)
	assert.Error(t, err)
	if err != nil {
		assert.True(t, strings.Contains(err.Error(), "shell injection detected"), "Expected shell injection error for space in bash arg, got: %v", err)
	}
}
