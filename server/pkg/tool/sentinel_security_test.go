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

func TestSentinelRCE_Sed_CommandTool(t *testing.T) {
	// 1. Configure a tool using "sed" with CommandTool (not LocalCommandTool)
	// implicit local execution (no container env)
	toolProto := v1.Tool_builder{
		Name: proto.String("sed-rce"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sed"),
		// Local is NOT set (defaults to false), so Upstream uses NewCommandTool
		// ContainerEnvironment is nil, so it uses local executor
	}.Build()

	// 2. Define call that allows passing arguments
	// We use "s,.,id,e" as expression.
	// Note: We need a valid input file for sed to run without erroring on file not found.
	// server/pkg/tool/testdata/input.txt might not exist.
	// We use "../../../go.mod" as it is in root.

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{expr}}", "../../../go.mod"},
        Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("expr")}.Build(),
			}.Build(),
        },
	}.Build()

	// 3. Create CommandTool manually
	cmdTool := NewCommandTool(toolProto, service, callDef, nil, "call-id")

	// 4. Payload: s,m,id,e
	req := &ExecutionRequest{
		ToolName: "sed-rce",
		Arguments: map[string]interface{}{
			"expr": "s,.,id,e",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := cmdTool.Execute(context.Background(), req)

	if err != nil {
		// If execution is blocked because sandbox is not supported, this is also secure (Fail Closed).
		if strings.Contains(err.Error(), "execution blocked for security") {
			t.Log("Secure: Execution blocked because sed sandbox is not supported")
			return
		}
		t.Fatalf("CommandTool.Execute failed unexpectedly: %v", err)
	}

	resMap, _ := result.(map[string]interface{})
	stderr, _ := resMap["stderr"].(string)
	stdout, _ := resMap["stdout"].(string)

	// Check if it is sandbox error in stderr
	if strings.Contains(stderr, "sandbox mode") {
		// Secure: Sandbox blocked the execution
		return
	}

	// If not sandboxed, check if RCE happened
	if strings.Contains(stdout, "uid=") {
		t.Log("VULNERABILITY REPRODUCED: sed executed 'id' command!")
	} else {
		t.Logf("Execution finished without sandbox error. Stdout: %q, Stderr: %q", stdout, stderr)
	}
	t.Fatal("Expected sandbox error in stderr, but got none (Vulnerable)")
}

func TestSentinelValidation_CommandTool_Decoded(t *testing.T) {
    // Test that CommandTool checks decoded values (which it currently misses)

    toolProto := v1.Tool_builder{Name: proto.String("cat-file")}.Build()
    service := configv1.CommandLineUpstreamService_builder{
        Command: proto.String("cat"),
    }.Build()

    callDef := configv1.CommandLineCallDefinition_builder{
        Args: []string{"{{file}}"},
        Parameters: []*configv1.CommandLineParameterMapping{
             configv1.CommandLineParameterMapping_builder{
                Schema: configv1.ParameterSchema_builder{Name: proto.String("file")}.Build(),
             }.Build(),
        },
    }.Build()

    cmdTool := NewCommandTool(toolProto, service, callDef, nil, "call-id")

    // Payload: %2e%2e/%2e%2e/%2e%2e/etc/passwd (encoded ../../../etc/passwd)
    // checkForPathTraversal handles encoded dots manually, so that one passes (or rather fails correctly).

    // We need something that passes non-decoded check but fails decoded check.
    // checkForArgumentInjection checks prefix "-".
    // %2d is "-".
    // "%2drf"

    req := &ExecutionRequest{
        ToolName: "cat-file",
        Arguments: map[string]interface{}{
            "file": "%2drf",
        },
    }
    req.ToolInputs, _ = json.Marshal(req.Arguments)

    _, err := cmdTool.Execute(context.Background(), req)

    // Expect error about argument injection (decoded)
    assert.Error(t, err)
    // Currently it will NOT error because CommandTool misses the check.
    if err != nil {
         assert.Contains(t, err.Error(), "argument injection")
    } else {
         t.Fatal("Expected argument injection error, but got success (Vulnerable)")
    }
}
