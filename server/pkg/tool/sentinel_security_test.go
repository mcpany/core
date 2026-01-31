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

func TestSentinelRCE_AwkInShell(t *testing.T) {
	// 1. Configure a tool that uses sh -c to run awk with user input in single quotes
	svc := configv1.CommandLineUpstreamService_builder{
		Command:          proto.String("sh"),
		WorkingDirectory: proto.String("."),
	}.Build()

	stringType := configv1.ParameterType(configv1.ParameterType_value["STRING"])

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "awk '{{script}}'"},
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
	toolProto.SetName("awk_wrapper")

	// Sentinel Security: We use NewCommandTool directly as it wraps local execution
	tool := NewCommandTool(
		toolProto,
		svc,
		callDef,
		nil,
		"awk_wrapper",
	)

	// 2. Craft a malicious input that uses awk's system() function
	payload := `BEGIN { system("echo pwned") }`

	inputMap := map[string]interface{}{
		"script": payload,
	}
	inputBytes, _ := json.Marshal(inputMap)

	req := &ExecutionRequest{
		ToolName: "awk_wrapper",
		ToolInputs: inputBytes,
	}

	// 3. Execute - Expect Error
	res, err := tool.Execute(context.Background(), req)

	// 4. Assert
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "shell injection detected")
	assert.Contains(t, err.Error(), "system(")
}

func TestSentinelRCE_Backticks(t *testing.T) {
	// 1. Configure a tool that uses sh -c to run perl with user input in single quotes
	svc := configv1.CommandLineUpstreamService_builder{
		Command:          proto.String("sh"),
		WorkingDirectory: proto.String("."),
	}.Build()

	stringType := configv1.ParameterType(configv1.ParameterType_value["STRING"])

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "perl -e '{{script}}'"},
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
	toolProto.SetName("perl_wrapper")

	tool := NewCommandTool(
		toolProto,
		svc,
		callDef,
		nil,
		"perl_wrapper",
	)

	// 2. Craft a malicious input that uses backticks
	// In perl, print `id` executes id.
	payload := "print `id`"

	inputMap := map[string]interface{}{
		"script": payload,
	}
	inputBytes, _ := json.Marshal(inputMap)

	req := &ExecutionRequest{
		ToolName: "perl_wrapper",
		ToolInputs: inputBytes,
	}

	// 3. Execute - Expect Error
	res, err := tool.Execute(context.Background(), req)

	// 4. Assert
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "shell injection detected")
	assert.Contains(t, err.Error(), "backtick")
}

func TestSentinelRCE_WhitespaceEvasion(t *testing.T) {
	// 1. Configure a tool that uses sh -c to run awk with user input in single quotes
	svc := configv1.CommandLineUpstreamService_builder{
		Command:          proto.String("sh"),
		WorkingDirectory: proto.String("."),
	}.Build()

	stringType := configv1.ParameterType(configv1.ParameterType_value["STRING"])

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "awk '{{script}}'"},
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
	toolProto.SetName("awk_wrapper_evasion")

	tool := NewCommandTool(
		toolProto,
		svc,
		callDef,
		nil,
		"awk_wrapper_evasion",
	)

	// 2. Craft a malicious input that uses whitespace to evade "system(" check
	payload := `BEGIN { system  ("echo pwned") }`

	inputMap := map[string]interface{}{
		"script": payload,
	}
	inputBytes, _ := json.Marshal(inputMap)

	req := &ExecutionRequest{
		ToolName: "awk_wrapper_evasion",
		ToolInputs: inputBytes,
	}

	// 3. Execute - Expect Error
	res, err := tool.Execute(context.Background(), req)

	// 4. Assert
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "shell injection detected")
	assert.Contains(t, err.Error(), "system(")
}
