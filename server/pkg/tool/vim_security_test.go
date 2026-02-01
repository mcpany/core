// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_VimInjection_Prevention(t *testing.T) {
	t.Parallel()
	// This test verifies that editors like vim are treated as shells
	// because they allow command execution.

	tool := v1.Tool_builder{
		Name: proto.String("test-tool-vim"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("vim"),
		Local:   proto.Bool(true),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("file")}.Build()}.Build(),
		},
		Args: []string{"{{file}}"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Injection attempt: Use vim's +command syntax to execute a shell command
	// We use '!' which is a dangerous character in our shell injection check.
	// If 'vim' is detected as a shell, this should be blocked.
	reqAttack := &ExecutionRequest{
		ToolName: "test-tool-vim",
		Arguments: map[string]interface{}{
			"file": "+!ls",
		},
	}
	reqAttack.ToolInputs, _ = json.Marshal(reqAttack.Arguments)

	// Set a short timeout because if vim starts, it might hang.
	// We only care if the security check blocks it BEFORE execution.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := localTool.Execute(ctx, reqAttack)

	// Expect error "argument injection detected" (caught by '+' check) or "shell injection detected"
	assert.Error(t, err)
	if err != nil {
		// We accept either because both are valid security blocks.
		// The argument injection check (for +) runs before shell injection check.
		msg := err.Error()
		if !strings.Contains(msg, "argument injection detected") {
			assert.Contains(t, msg, "shell injection detected")
		}
	}
}
