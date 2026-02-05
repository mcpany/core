// Copyright 2025 Author(s) of MCP Any
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

func TestNiceInjection_Bypass(t *testing.T) {
	// 1. Configure a tool that uses 'nice' to run 'sh -c' with user input.
	// Since 'nice' was missing from isShellCommand, checks were skipped.
	svc := configv1.CommandLineUpstreamService_builder{
		Command:          proto.String("nice"),
		WorkingDirectory: proto.String("."),
	}.Build()

	stringType := configv1.ParameterType(configv1.ParameterType_value["STRING"])

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"sh", "-c", "{{script}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("script"),
					Type: &stringType,
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := &v1.Tool{}
	toolProto.SetName("nice_wrapper")

	tool := NewLocalCommandTool(
		toolProto,
		svc,
		callDef,
		nil,
		"nice_wrapper",
	)

	// 2. Craft a malicious input that uses spaces in an unquoted context.
	// This should be blocked by checkUnquotedInjection if 'nice' is recognized as a shell command.
	payload := `echo pwned`

	inputMap := map[string]interface{}{
		"script": payload,
	}
	inputBytes, _ := json.Marshal(inputMap)

	req := &ExecutionRequest{
		ToolName: "nice_wrapper",
		ToolInputs: inputBytes,
	}

	// 3. Execute - Expect Error (Security Block)
	res, err := tool.Execute(context.Background(), req)

	// 4. Assert
	// If the vulnerability exists, err will be nil (execution allowed).
	// If fixed, err will be "shell injection detected".
	assert.Error(t, err, "Expected injection to be blocked")
	if err != nil {
		assert.Contains(t, err.Error(), "shell injection detected")
	} else {
		t.Logf("Vulnerability confirmed: 'nice' allowed execution of '%s'", payload)
		t.Logf("Result: %v", res)
	}
}
