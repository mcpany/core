// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestPerlInjection(t *testing.T) {
	// Setup the tool configuration
	// We simulate a tool that runs a perl command with a double-quoted argument
	cmdService := &configv1.CommandLineUpstreamService{}
	cmdService.SetCommand("perl")

	paramSchema := &configv1.ParameterSchema{}
	paramSchema.SetName("name")
	paramSchema.SetType(configv1.ParameterType_STRING)

	paramMapping := &configv1.CommandLineParameterMapping{}
	paramMapping.SetSchema(paramSchema)

	callDef := &configv1.CommandLineCallDefinition{}
	callDef.SetArgs([]string{"-e", "print \"Hello {{name}}\""})
	callDef.SetParameters([]*configv1.CommandLineParameterMapping{paramMapping})

	// NewCommandTool returns a Tool interface
	toolStruct := &v1.Tool{}
	toolStruct.SetName("perl_hello")

	tool := NewCommandTool(
		toolStruct,
		cmdService,
		callDef,
		nil,
		"test-call-id",
	)

	ctx := context.Background()

	// Attack payload: Perl array interpolation execution
	// @{[ system('echo INJECTED') ]}
	// Note: We avoid characters blocked by quoteLevel=1: " $ ` \ %
	// The payload uses @, {, [, ], (, ), ', which are NOT blocked in quoteLevel=1
	payload := "@{[ system('echo INJECTED') ]}"

	req := &ExecutionRequest{
		ToolName: "perl_hello",
		ToolInputs: []byte(`{"name": "` + payload + `"}`),
	}

	// Execute
	result, err := tool.Execute(ctx, req)

	// If the injection is successful, the output will contain "INJECTED"
	// or the function might return an error if we successfully blocked it.
	// If we FAILED to block it, err will be nil (or success) and stdout will have INJECTED.

	if err != nil {
		t.Logf("Execution error (expected if blocked): %v", err)
		// If the error explicitly mentions injection detection, we are good.
		if assert.Contains(t, err.Error(), "injection detected") {
			return
		}
	} else {
		resMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		stdout := resMap["stdout"].(string)
		t.Logf("Stdout: %s", stdout)

		if assert.Contains(t, stdout, "INJECTED") {
			t.FailNow() // VULNERABILITY CONFIRMED
		}
	}
}

func TestPerlReadpipeInjection(t *testing.T) {
	cmd := "perl"
	cmdService := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()

	paramSchema := configv1.ParameterSchema_builder{
		Name: proto.String("name"),
		Type: configv1.ParameterType_STRING.Enum(),
	}.Build()

	paramMapping := configv1.CommandLineParameterMapping_builder{
		Schema: paramSchema,
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "print '{{name}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{paramMapping},
	}.Build()

	toolStruct := v1.Tool_builder{
		Name: proto.String("perl_rce"),
	}.Build()

	tool := NewLocalCommandTool(
		toolStruct,
		cmdService,
		callDef,
		nil,
		"test-call-id",
	)

	ctx := context.Background()

	// Attack payload: ' . readpipe('echo INJECTED') . '
	// This closes the single quote, executes readpipe, and opens another single quote.
	// Resulting Perl code: print '' . readpipe('echo INJECTED') . ''
	// This payload was NOT blocked before because checkInterpreterFunctionCalls was not run for quoteLevel=2 (Single Quoted).
	// Now it should be blocked.
	jsonInput := "{\"name\": \"' . readpipe('echo INJECTED') . '\"}"

	req := &ExecutionRequest{
		ToolName:   "perl_rce",
		ToolInputs: []byte(jsonInput),
	}

	_, err := tool.Execute(ctx, req)

	if err == nil {
        assert.Error(t, err, "Expected error for readpipe injection")
	} else {
		assert.Contains(t, err.Error(), "injection detected", "Expected injection detection error")
	}
}
