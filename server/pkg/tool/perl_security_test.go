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

func TestSentinelRCE_PerlArrayInterpolation(t *testing.T) {
	// 1. Configure a tool that uses perl with user input in double quotes
	svc := configv1.CommandLineUpstreamService_builder{
		Command:          proto.String("perl"),
		WorkingDirectory: proto.String("."),
	}.Build()

	stringType := configv1.ParameterType(configv1.ParameterType_value["STRING"])

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "print \"{{script}}\""},
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
	toolProto.SetName("perl_double_quote")

	tool := NewLocalCommandTool(
		toolProto,
		svc,
		callDef,
		nil,
		"perl_double_quote",
	)

	// 2. Craft a malicious input that uses perl array interpolation @{[...]}
	payload := "@{[system('echo pwned')]}"

	inputMap := map[string]interface{}{
		"script": payload,
	}
	inputBytes, _ := json.Marshal(inputMap)

	req := &ExecutionRequest{
		ToolName: "perl_double_quote",
		ToolInputs: inputBytes,
		DryRun: true, // Use DryRun to avoid actual execution but trigger validation
	}

	// 3. Execute - Expect Error from Validator
	_, err := tool.Execute(context.Background(), req)

	// 4. Assert
	if err == nil {
		t.Fatal("Expected validation error for Perl array interpolation, but got nil")
	}
	assert.Contains(t, err.Error(), "injection detected")
}
