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
	"google.golang.org/protobuf/types/known/structpb"
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

	tool := NewLocalCommandTool(
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

	tool := NewLocalCommandTool(
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

	tool := NewLocalCommandTool(
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

func TestSentinelRCE_QuoteParsingBypass(t *testing.T) {
	// 1. Configure a tool that uses sh -c with a vulnerable template (accidentally using \' inside single quotes)
	// Template: echo 'foo\'{{input}}'
	// This template is parsed by Bash as: 'foo\' (string) then {{input}} (unquoted) then ' (string start)
	svc := configv1.CommandLineUpstreamService_builder{
		Command:          proto.String("sh"),
		WorkingDirectory: proto.String("."),
	}.Build()

	stringType := configv1.ParameterType(configv1.ParameterType_value["STRING"])

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo 'foo\\'{{input}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("input"),
					Type: &stringType,
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := &v1.Tool{}
	toolProto.SetName("vulnerable_template")

	tool := NewLocalCommandTool(
		toolProto,
		svc,
		callDef,
		nil,
		"vulnerable_template",
	)

	// 2. Craft a malicious input that exploits the unquoted context
	// If the parser thinks it's single-quoted, it would allow semicolon.
	// But since it's actually unquoted, semicolon allows RCE.
	payload := "; echo pwned"

	inputMap := map[string]interface{}{
		"input": payload,
	}
	inputBytes, _ := json.Marshal(inputMap)

	req := &ExecutionRequest{
		ToolName: "vulnerable_template",
		ToolInputs: inputBytes,
	}

	// 3. Execute - Expect Error
	res, err := tool.Execute(context.Background(), req)

	// 4. Assert
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "shell injection detected")
	// It should fail because it detected dangerous character (semicolon) in unquoted context
	assert.Contains(t, err.Error(), ";")
}

func TestLocalCommandTool_InterpreterInjection_DoubleQuotes_Bypass(t *testing.T) {
	t.Parallel()

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type": structpb.NewStringValue("object"),
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"code": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := v1.Tool_builder{
		Name:        proto.String("python-eval-tool"),
		InputSchema: inputSchema,
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run: python -c eval("{{code}}")
	// Note the double quotes around {{code}}
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "eval(\"{{code}}\")"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: __import__('os').system('id')
	// This uses single quotes, which are ALLOWED in double quotes context (Level 1).
	// It contains "system(", which is BLOCKED in single quotes context (Level 2), but was NOT in Level 1 prior to fix.
	payload := "__import__('os').system('id')"

	req := &ExecutionRequest{
		ToolName: "python-eval-tool",
		Arguments: map[string]interface{}{
			"code": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	// We expect the execution to be BLOCKED by the new security check.
	// err should contain "shell injection detected"

	if err == nil {
		t.Fatal("Execution succeeded but should have been blocked")
	}

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shell injection detected")
	assert.Contains(t, err.Error(), "dangerous function call")
}

func TestSentinelRCE_FalsePositive_Filesystem(t *testing.T) {
	// Verify that valid English words containing "system" are not blocked.
	// E.g. "filesystem(s)" contains "system(".
	// We use 'sh' to trigger checkForShellInjection.

	svc := configv1.CommandLineUpstreamService_builder{
		Command:          proto.String("sh"),
		WorkingDirectory: proto.String("."),
	}.Build()

	stringType := configv1.ParameterType(configv1.ParameterType_value["STRING"])

	// Test case 1: Single Quoted
	callDefSingle := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo '{{text}}'"},
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
	toolProto.SetName("sh_wrapper_single")

	toolSingle := NewLocalCommandTool(
		toolProto,
		svc,
		callDefSingle,
		nil,
		"sh_wrapper_single",
	)

	payload := "Check the filesystem(s)"

	inputMap := map[string]interface{}{
		"text": payload,
	}
	inputBytes, _ := json.Marshal(inputMap)

	reqSingle := &ExecutionRequest{
		ToolName:   "sh_wrapper_single",
		ToolInputs: inputBytes,
	}

	res, err := toolSingle.Execute(context.Background(), reqSingle)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	// Test case 2: Double Quoted
	callDefDouble := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo \"{{text}}\""},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("text"),
					Type: &stringType,
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProtoDouble := &v1.Tool{}
	toolProtoDouble.SetName("sh_wrapper_double")

	toolDouble := NewLocalCommandTool(
		toolProtoDouble,
		svc,
		callDefDouble,
		nil,
		"sh_wrapper_double",
	)

	reqDouble := &ExecutionRequest{
		ToolName:   "sh_wrapper_double",
		ToolInputs: inputBytes,
	}

	resDouble, errDouble := toolDouble.Execute(context.Background(), reqDouble)
	assert.NoError(t, errDouble)
	assert.NotNil(t, resDouble)
}
