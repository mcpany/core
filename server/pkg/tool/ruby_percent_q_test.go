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

func TestLocalCommandTool_RubyExecution_PercentQ_Security(t *testing.T) {
	// This test asserts that Ruby execution BLOCKS bypassing restrictions
	// using %q{} to avoid quotes.
	// Initially expected to FAIL until fix is implemented.

	tool := v1.Tool_builder{
		Name: proto.String("ruby-percent-q-test"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `ruby -e '{{input}}'`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "'{{input}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: system %q{echo PWNED}
	payload := "system %q{echo PWNED}"

	req := &ExecutionRequest{
		ToolName: "ruby-percent-q-test",
		Arguments: map[string]interface{}{
			"input": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)

	if err == nil {
		resultMap, ok := result.(map[string]interface{})
		output := ""
		if ok {
			output = resultMap["combined_output"].(string)
		}
		t.Errorf("Security check failed: Expected error but got success. Output: %s", output)
	} else {
		assert.Contains(t, err.Error(), "interpreter injection detected", "Error should mention injection detection")
	}
}
