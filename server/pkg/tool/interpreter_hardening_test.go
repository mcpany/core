package tool

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestCommandTool_SecurityHardening(t *testing.T) {
	// Setup LocalCommandTool with python command
	toolDef := v1.Tool_builder{
		Name: proto.String("python-echo"),
	}.Build()

	cmdService := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
	}).Build()

	// Scenario 1: Multiple placeholders
	// This verifies the fix for the logic bug where subsequent replacements overwrote previous ones
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "print('{{a}} {{b}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("a")}.Build(),
			}.Build(),
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("b")}.Build(),
			}.Build(),
		},
	}.Build()

	cmdTool := NewLocalCommandTool(toolDef, cmdService, callDef, nil, "test-call-id")

	// Verify multiple replacement fix
	inputMap := map[string]interface{}{
		"a": "Hello",
		"b": "World",
	}
	jsonBytes, _ := json.Marshal(inputMap)
	req := &ExecutionRequest{ToolInputs: jsonBytes}

	result, err := cmdTool.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}
	resMap, _ := result.(map[string]interface{})
	stdout := resMap["stdout"].(string)
	if !strings.Contains(stdout, "Hello World") {
		t.Errorf("Multiple replacement failed. Output: %q", stdout)
	}

	// Scenario 2: Hardening check - getattr
	// We use 'exec' to simulate code execution context where checkInterpreterFunctionCalls matters
	// This verifies that 'getattr' is now in the dangerous keywords list
	callDefExec := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "exec('{{code}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build(),
			}.Build(),
		},
	}.Build()
	cmdToolExec := NewLocalCommandTool(toolDef, cmdService, callDefExec, nil, "test-exec-id")

	payloadGetattr := "import os; getattr(os, 'system')('echo PWNED')"
	inputGetattr := map[string]interface{}{"code": payloadGetattr}
	jsonBytesGetattr, _ := json.Marshal(inputGetattr)
	reqGetattr := &ExecutionRequest{ToolInputs: jsonBytesGetattr}

	_, err = cmdToolExec.Execute(context.Background(), reqGetattr)
	if err == nil {
		t.Errorf("Exploit succeeded! getattr was not blocked.")
	} else {
		if !strings.Contains(err.Error(), "getattr") && !strings.Contains(err.Error(), "os.") {
			t.Errorf("Blocked but not by expected rule? Error: %v", err)
		}
	}

	// Scenario 3: Hardening check - subprocess.run
	// Payload: import subprocess; subprocess.run(['ls'])
	// Previously bypassed because 'subprocess(' was checked, but 'subprocess.run' doesn't match.
	// Now should be blocked by 'subprocess.' check.

	payloadSubprocess := "import subprocess; subprocess.run(['echo', 'PWNED'])"
	inputSubprocess := map[string]interface{}{"code": payloadSubprocess}
	jsonBytesSubprocess, _ := json.Marshal(inputSubprocess)
	reqSubprocess := &ExecutionRequest{ToolInputs: jsonBytesSubprocess}

	_, err = cmdToolExec.Execute(context.Background(), reqSubprocess)
	if err == nil {
		t.Errorf("Exploit succeeded! subprocess.run was not blocked.")
	} else {
		if !strings.Contains(err.Error(), "subprocess") {
			t.Errorf("Blocked but not by expected rule? Error: %v", err)
		}
	}
}
