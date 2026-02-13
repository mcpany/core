// Copyright 2025 Author(s) of MCP Any
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
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSedInjectionRepro(t *testing.T) {
	// This test reproduces a vulnerability where sed commands (e.g., 'e' for execute)
	// can be injected when sed is used inside a shell command string.

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("bash"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo input | sed \"{{pattern}}\""},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("pattern"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := v1.Tool_builder{
		Name: proto.String("sed-tool"),
	}.Build()

	tool := NewLocalCommandTool(
		toolProto,
		service,
		callDef,
		nil, // Policies
		"test-call",
	)

	// Malicious input: executes 'echo pwned' via sed's 'e' command
	// The sed command '1e echo pwned' executes 'echo pwned' for the first line.
	// Since we are processing input (from echo input), it will trigger.

	inputs := map[string]interface{}{
		"pattern": "1e echo pwned",
	}

	inputBytes, _ := json.Marshal(inputs)
	req := &ExecutionRequest{
		ToolName:   "sed-tool",
		ToolInputs: inputBytes,
	}

	// We expect this to FAIL with a security error now that it is fixed.
	result, err := tool.Execute(context.Background(), req)

	// Assert that execution is BLOCKED.
	if err == nil {
		resMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		stdout, ok := resMap["stdout"].(string)
		if ok && strings.Contains(stdout, "pwned") {
			t.Fatalf("Vulnerability NOT blocked: 'pwned' found in output.")
		}
		t.Fatalf("Command executed successfully (unexpected). Result: %v", result)
	}

	// Check if it's the expected security error.
	assert.Contains(t, err.Error(), "sed injection detected", "Expected sed injection detection error")
}
