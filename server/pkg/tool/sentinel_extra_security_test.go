// Copyright 2026 Author(s) of MCP Any
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
	"google.golang.org/protobuf/proto"
)

func TestSentinelLFI_Awk_InShell(t *testing.T) {
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
	toolProto.SetName("awk_lfi_shell")

	tool := NewLocalCommandTool(
		toolProto,
		svc,
		callDef,
		nil,
		"awk_lfi_shell",
	)

	// 2. Craft a malicious input that uses awk's getline to read a file
	payload := `BEGIN { while((getline line < "types.go") > 0) print line }`

	inputMap := map[string]interface{}{
		"script": payload,
	}
	inputBytes, _ := json.Marshal(inputMap)

	req := &ExecutionRequest{
		ToolName:   "awk_lfi_shell",
		ToolInputs: inputBytes,
	}

	// 3. Execute - Expect Error
	res, err := tool.Execute(context.Background(), req)

	// 4. Assert
	assert.Error(t, err)
	assert.Nil(t, res)
	// It should block due to getline
	assert.Contains(t, err.Error(), "awk injection detected")
	assert.Contains(t, err.Error(), "getline")
}

func TestSentinelLFI_Python(t *testing.T) {
	// 1. Configure a tool that uses python -c with user input in single quotes (via sh -c)
	svc := configv1.CommandLineUpstreamService_builder{
		Command:          proto.String("python3"),
		WorkingDirectory: proto.String("."),
	}.Build()

	stringType := configv1.ParameterType(configv1.ParameterType_value["STRING"])

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "exec('{{script}}')"},
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
	toolProto.SetName("python_lfi")

	tool := NewLocalCommandTool(
		toolProto,
		svc,
		callDef,
		nil,
		"python_lfi",
	)

	// 2. Craft a malicious input that reads a file
	payload := `print(open("types.go").read())`

	inputMap := map[string]interface{}{
		"script": payload,
	}
	inputBytes, _ := json.Marshal(inputMap)

	req := &ExecutionRequest{
		ToolName:   "python_lfi",
		ToolInputs: inputBytes,
	}

	// 3. Execute - Expect Error
	res, err := tool.Execute(context.Background(), req)

	// 4. Assert
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "interpreter injection detected")
	assert.Contains(t, err.Error(), "open(")
}

func TestSentinelRCE_PythonSubprocessCall(t *testing.T) {
	// 1. Configure a tool that uses python -c to run a script with user input in single quotes
	svc := configv1.CommandLineUpstreamService_builder{
		Command:          proto.String("python3"),
		WorkingDirectory: proto.String("."),
	}.Build()

	stringType := configv1.ParameterType(configv1.ParameterType_value["STRING"])

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "exec('{{msg}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("msg"),
					Type: &stringType,
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := &v1.Tool{}
	toolProto.SetName("python_wrapper")

	tool := NewLocalCommandTool(
		toolProto,
		svc,
		callDef,
		nil,
		"python_wrapper",
	)

	// 2. Craft a malicious input that uses subprocess.call WITHOUT single quotes
	payload := `import subprocess; subprocess.call(["echo", "pwned"])`

	inputMap := map[string]interface{}{
		"msg": payload,
	}
	inputBytes, _ := json.Marshal(inputMap)

	req := &ExecutionRequest{
		ToolName:   "python_wrapper",
		ToolInputs: inputBytes,
	}

	// 3. Execute - Expect Error
	res, err := tool.Execute(context.Background(), req)

	// 4. Assert
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "interpreter injection detected")
	// It matches "subprocess" or "import"
	assert.True(t, strings.Contains(err.Error(), "subprocess") || strings.Contains(err.Error(), "import"))
}

func TestSentinel_FalsePositives(t *testing.T) {
	// 1. Configure a tool that uses python -c to echo text
	svc := configv1.CommandLineUpstreamService_builder{
		Command:          proto.String("python3"),
		WorkingDirectory: proto.String("."),
	}.Build()

	stringType := configv1.ParameterType(configv1.ParameterType_value["STRING"])

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "print('{{text}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("text"),
					Type: &stringType,
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := &v1.Tool{}
	toolProto.SetName("python_echo")

	tool := NewLocalCommandTool(
		toolProto,
		svc,
		callDef,
		nil,
		"python_echo",
	)

	// 2. Craft input with words that contain dangerous substrings but are safe
	payload := "This is an important message about openness."

	inputMap := map[string]interface{}{
		"text": payload,
	}
	inputBytes, _ := json.Marshal(inputMap)

	req := &ExecutionRequest{
		ToolName:   "python_echo",
		ToolInputs: inputBytes,
	}

	// 3. Execute - Expect Success
	res, err := tool.Execute(context.Background(), req)

	// 4. Assert
	assert.NoError(t, err)
	assert.NotNil(t, res)

	resMap, _ := res.(map[string]interface{})
	stdout, _ := resMap["stdout"].(string)
	assert.Contains(t, stdout, "important")
	assert.Contains(t, stdout, "openness")
}
