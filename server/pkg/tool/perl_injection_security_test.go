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
