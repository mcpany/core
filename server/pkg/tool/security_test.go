// Copyright 2025 Author(s) of MCP Any
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

func TestLocalCommandTool_ArgumentInjection_Prevention(t *testing.T) {
	t.Parallel()
	// This test verifies that argument injection via placeholders is prevented.

	tool := v1.Tool_builder{
		Name:        proto.String("test-tool-cat"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("cat"),
		Local:   proto.Bool(true),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("file")}.Build()}.Build(),
		},
		Args: []string{"{{file}}"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Case 1: Safe input (relative path)
	reqSafe := &ExecutionRequest{
		ToolName: "test-tool-cat",
		Arguments: map[string]interface{}{
			"file": "hosts",
		},
	}
	reqSafe.ToolInputs, _ = json.Marshal(reqSafe.Arguments)

	_, err := localTool.Execute(context.Background(), reqSafe)
	if err != nil {
		assert.NotContains(t, err.Error(), "argument injection")
		assert.NotContains(t, err.Error(), "absolute path detected")
	}

	// Case 2: Argument Injection
	reqAttack := &ExecutionRequest{
		ToolName: "test-tool-cat",
		Arguments: map[string]interface{}{
			"file": "-n", // Attempt to inject a flag
		},
	}
	reqAttack.ToolInputs, _ = json.Marshal(reqAttack.Arguments)

	_, err = localTool.Execute(context.Background(), reqAttack)

	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "argument injection")
	}

	// Case 3: Negative number (should be allowed)
	reqNegative := &ExecutionRequest{
		ToolName: "test-tool-cat",
		Arguments: map[string]interface{}{
			"file": "-5",
		},
	}
	reqNegative.ToolInputs, _ = json.Marshal(reqNegative.Arguments)

	_, err = localTool.Execute(context.Background(), reqNegative)
	if err != nil {
		assert.NotContains(t, err.Error(), "argument injection")
	}
}

func TestLocalCommandTool_ShellInjection_Prevention(t *testing.T) {
	t.Parallel()
	// Test Case 1: Unquoted Placeholder (Vulnerable configuration)
	t.Run("Unquoted Placeholder", func(t *testing.T) {
		tool := v1.Tool_builder{Name: proto.String("test-tool-sh")}.Build()
		service := configv1.CommandLineUpstreamService_builder{
			Command: proto.String("sh"),
			Local:   proto.Bool(true),
		}.Build()
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build()}.Build(),
			},
			Args: []string{"-c", "echo {{msg}}"},
		}.Build()
		localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

		// Injection attempt
		reqAttack := &ExecutionRequest{
			ToolName: "test-tool-sh",
			Arguments: map[string]interface{}{
				"msg": "hello; cat /etc/passwd",
			},
		}
		reqAttack.ToolInputs, _ = json.Marshal(reqAttack.Arguments)

		_, err := localTool.Execute(context.Background(), reqAttack)
		assert.Error(t, err)
		if err != nil {
			// Now blocked by interpreter hardening
			assert.Contains(t, strings.ToLower(err.Error()), "security risk: template substitution is not allowed")
		}

		// Safe input but with special chars that are dangerous in unquoted context
		reqSafeish := &ExecutionRequest{
			ToolName: "test-tool-sh",
			Arguments: map[string]interface{}{
				"msg": "Law & Order",
			},
		}
		reqSafeish.ToolInputs, _ = json.Marshal(reqSafeish.Arguments)
		_, err = localTool.Execute(context.Background(), reqSafeish)
		// Should fail because & is dangerous unquoted
		assert.Error(t, err)
	})

	// Test Case 2: Single Quoted Placeholder (Safer configuration)
	t.Run("Single Quoted Placeholder", func(t *testing.T) {
		tool := v1.Tool_builder{Name: proto.String("test-tool-sh-quoted")}.Build()
		service := configv1.CommandLineUpstreamService_builder{
			Command: proto.String("sh"),
			Local:   proto.Bool(true),
		}.Build()
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build()}.Build(),
			},
			Args: []string{"-c", "echo '{{msg}}'"},
		}.Build()
		localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

		// Safe input with special chars
		reqSafe := &ExecutionRequest{
			ToolName: "test-tool-sh-quoted",
			Arguments: map[string]interface{}{
				"msg": "Law & Order",
			},
		}
		reqSafe.ToolInputs, _ = json.Marshal(reqSafe.Arguments)

		_, err := localTool.Execute(context.Background(), reqSafe)
		// Should PASS because it's quoted
		if err != nil {
			assert.NotContains(t, err.Error(), "shell injection detected")
		}

		// Breakout attempt
		reqBreakout := &ExecutionRequest{
			ToolName: "test-tool-sh-quoted",
			Arguments: map[string]interface{}{
				"msg": "foo'bar",
			},
		}
		reqBreakout.ToolInputs, _ = json.Marshal(reqBreakout.Arguments)
		_, err = localTool.Execute(context.Background(), reqBreakout)
		// Should FAIL because it contains single quote
		assert.Error(t, err)
	})

	// Test Case 3: Double Quoted Placeholder
	t.Run("Double Quoted Placeholder", func(t *testing.T) {
		tool := v1.Tool_builder{Name: proto.String("test-tool-sh-dquoted")}.Build()
		service := configv1.CommandLineUpstreamService_builder{
			Command: proto.String("sh"),
			Local:   proto.Bool(true),
		}.Build()
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build()}.Build(),
			},
			Args: []string{"-c", "echo \"{{msg}}\""},
		}.Build()
		localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

		// Safe input
		reqSafe := &ExecutionRequest{
			ToolName: "test-tool-sh-dquoted",
			Arguments: map[string]interface{}{
				"msg": "Hello World",
			},
		}
		reqSafe.ToolInputs, _ = json.Marshal(reqSafe.Arguments)
		_, err := localTool.Execute(context.Background(), reqSafe)
		// Now blocked by strict interpreter hardening
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, strings.ToLower(err.Error()), "security risk: template substitution is not allowed")
		}

		// Variable expansion attempt (dangerous in double quotes)
		reqVar := &ExecutionRequest{
			ToolName: "test-tool-sh-dquoted",
			Arguments: map[string]interface{}{
				"msg": "$HOME",
			},
		}
		reqVar.ToolInputs, _ = json.Marshal(reqVar.Arguments)
		_, err = localTool.Execute(context.Background(), reqVar)
		assert.Error(t, err) // Should block $

		// Breakout attempt
		reqBreakout := &ExecutionRequest{
			ToolName: "test-tool-sh-dquoted",
			Arguments: map[string]interface{}{
				"msg": "foo\"bar",
			},
		}
		reqBreakout.ToolInputs, _ = json.Marshal(reqBreakout.Arguments)
		_, err = localTool.Execute(context.Background(), reqBreakout)
		assert.Error(t, err)
		if err != nil {
			// Now blocked by interpreter hardening
			assert.Contains(t, strings.ToLower(err.Error()), "security risk: template substitution is not allowed")
		}

		// Backslash escape attempt (to escape closing quote)
		reqBackslash := &ExecutionRequest{
			ToolName: "test-tool-sh-dquoted",
			Arguments: map[string]interface{}{
				"msg": "foo\\",
			},
		}
		reqBackslash.ToolInputs, _ = json.Marshal(reqBackslash.Arguments)
		_, err = localTool.Execute(context.Background(), reqBackslash)
		assert.Error(t, err) // Should block \

		// Windows CMD injection attempt (dangerous in double quotes on Windows)
		reqWinCmd := &ExecutionRequest{
			ToolName: "test-tool-sh-dquoted",
			Arguments: map[string]interface{}{
				"msg": "%PATH%",
			},
		}
		reqWinCmd.ToolInputs, _ = json.Marshal(reqWinCmd.Arguments)
		_, err = localTool.Execute(context.Background(), reqWinCmd)
		assert.Error(t, err) // Should block %
	})

	// Test Case 4: Non-Shell Command
	t.Run("Non-Shell Command", func(t *testing.T) {
		tool := v1.Tool_builder{Name: proto.String("test-tool-echo")}.Build()
		service := configv1.CommandLineUpstreamService_builder{
			Command: proto.String("echo"), // Not a shell
			Local:   proto.Bool(true),
		}.Build()
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build()}.Build(),
			},
			Args: []string{"{{msg}}"},
		}.Build()
		localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

		// Input with shell chars - should be allowed for non-shell command
		reqSafe := &ExecutionRequest{
			ToolName: "test-tool-echo",
			Arguments: map[string]interface{}{
				"msg": "Law & Order",
			},
		}
		reqSafe.ToolInputs, _ = json.Marshal(reqSafe.Arguments)
		_, err := localTool.Execute(context.Background(), reqSafe)
		assert.NoError(t, err)
	})
}

func TestLocalCommandTool_Execute_PythonInjection(t *testing.T) {
	// This test demonstrates that python is not currently treated as a shell command,
	// allowing code injection via argument substitution.

	toolDef := v1.Tool_builder{
		Name: proto.String("python_tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "print('{{msg}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("msg"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	ct := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call-id")

	// Malicious input trying to break out of python string
	// msg = '); print("INJECTED"); print('
	// Resulting code: print(''); print("INJECTED"); print('')
	// Note: We need to escape quotes for JSON

	injectionPayload := "'); print(\"INJECTED\"); print('"
	jsonInput, _ := json.Marshal(map[string]string{"msg": injectionPayload})

	req := &ExecutionRequest{
		ToolName: "python_tool",
		ToolInputs: jsonInput,
	}

	// Execute
	_, err := ct.Execute(context.Background(), req)

	// Expect strict shell injection prevention to kick in for Python
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, strings.ToLower(err.Error()), "security risk: template substitution is not allowed")
	}
}

func TestLocalCommandTool_ShellInjection_ControlChars(t *testing.T) {
	t.Parallel()
	tool := v1.Tool_builder{
		Name: proto.String("test-tool-shell-control"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"), // This triggers shell injection checks
		Local:   proto.Bool(true),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo {{arg}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("arg")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Test cases for control characters
	testCases := []struct {
		name       string
		input      string
		shouldFail bool
	}{
		{"CarriageReturn", "hello\rworld", true},
		{"Tab", "hello\tworld", true}, // We want to block this as it can split args in unquoted context
		{"VerticalTab", "hello\vworld", true},
		{"FormFeed", "hello\fworld", true},
		{"Safe", "helloworld", true}, // Now blocked by strict interpreter hardening
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &ExecutionRequest{
				ToolName: "test-tool-shell-control",
				Arguments: map[string]interface{}{
					"arg": tc.input,
				},
			}
			req.ToolInputs, _ = json.Marshal(req.Arguments)

			_, err := localTool.Execute(context.Background(), req)
			if tc.shouldFail {
				if err == nil {
					t.Fatalf("Expected error for input %q, but got nil", tc.input)
				}
				// Now blocked by interpreter hardening
				if strings.Contains(strings.ToLower(err.Error()), "security risk") {
					assert.Contains(t, strings.ToLower(err.Error()), "security risk: template substitution is not allowed")
				} else {
					assert.Contains(t, err.Error(), "shell injection detected")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
