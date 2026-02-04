// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestBacktickInjection_RCE_Reproduction(t *testing.T) {
	t.Parallel()
	tool := v1.Tool_builder{
		Name:        proto.String("rce-tool"),
		Description: proto.String("A tool vulnerable to RCE via backticks"),
	}.Build()

	// We use ruby because it uses backticks for shell execution and is in the isInterpreter list
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
		Local:   proto.Bool(true),
	}.Build()

	// The argument uses backticks.
	// If the security check is weak, injecting "echo vulnerable" will execute the command.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "print `{{msg}}`"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "rce-call-id")

	// Payload that is a valid shell command but contains characters that should be blocked
	// if we were enforcing strict checks.
	// We use 'echo vulnerable' which contains a space (blocked by checkUnquotedInjection/checkBacktickInjection normally).
	req := &ExecutionRequest{
		ToolName: "rce-tool",
		Arguments: map[string]interface{}{
			"msg": "echo vulnerable",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute the tool
	_, err := localTool.Execute(context.Background(), req)

	// AFTER FIX: The execution should fail with a security error
	assert.Error(t, err, "Expected security error blocking backtick injection")
	if err != nil {
		assert.Contains(t, err.Error(), "shell injection detected", "Error message should indicate injection detection")
	}
}
