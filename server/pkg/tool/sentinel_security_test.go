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
