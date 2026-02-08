// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerlArrayInterpolationInjection(t *testing.T) {
	// This test reproduces an RCE vulnerability where Perl code can be injected
	// into a double-quoted argument string using array interpolation @{[ ... ]}.

	service := &configv1.CommandLineUpstreamService{}
    service.SetCommand("perl")

	callDef := &configv1.CommandLineCallDefinition{}
    // Use DOUBLE QUOTES in the perl script
    callDef.SetArgs([]string{"-e", "print \"Hello, {{name}}!\""})

    // Parameters
    schema := &configv1.ParameterSchema{}
    schema.SetName("name")

    param := &configv1.CommandLineParameterMapping{}
    param.SetSchema(schema)

    callDef.SetParameters([]*configv1.CommandLineParameterMapping{param})

	toolProto := &pb.Tool{}
    toolProto.SetName("perl_hello")

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Payload: Perl array interpolation.
    // We avoid " (double quote) because it's blocked by Level 1 check.
    // We use qw() for strings.
	payload := "@{[ system(qw(echo RCE_SUCCESS)) ]}"

    inputs := map[string]string{"name": payload}
    inputsBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName: "perl_hello",
		ToolInputs: inputsBytes,
	}

	result, err := tool.Execute(context.Background(), req)

	if err != nil {
		t.Logf("Execution blocked: %v", err)
        if assert.Contains(t, err.Error(), "injection detected") {
             t.Log("Blocked by injection detection (Good)")
        } else {
             t.Logf("Blocked by unexpected error: %v", err)
        }
	} else {
		resMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		stdout, ok := resMap["stdout"].(string)
		require.True(t, ok)

		t.Logf("Stdout: %s", stdout)

		// Note: system() output goes to stdout (or merged), so we should see RCE_SUCCESS
        // The print output will contain the return value of system (usually 0).
		if assert.Contains(t, stdout, "RCE_SUCCESS") {
			t.Fatal("VULNERABILITY CONFIRMED: Perl code injection successful")
		}
	}
}
