package tool

import (
	"context"
	"fmt"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestPythonCodeInjection(t *testing.T) {
	cmd := "python3"

	// Tool: python3 -c "print('Hello {{name}}')"
	toolDef := v1.Tool_builder{Name: proto.String("python-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "print('Hello {{name}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("name")}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	// Payload: Break out of print('...') and execute system
	injection := "'); import os; s=os.system; s(\"echo RCE_SUCCESS\"); print('"

    // Escape for JSON
    escapedInjection := strings.ReplaceAll(injection, "\"", "\\\"")

	req := &ExecutionRequest{
		ToolName: "python-tool",
		ToolInputs: []byte(fmt.Sprintf(`{"name": "%s"}`, escapedInjection)),
	}

	result, err := tool.Execute(context.Background(), req)
	if err != nil {
		// Expect security error
		if strings.Contains(strings.ToLower(err.Error()), "security risk: template substitution is not allowed") {
			t.Logf("Success: Blocked by security hardening: %v", err)
			return
		}

		t.Fatalf("Execute failed with unexpected error: %v", err)
	}

	resMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not map: %v", result)
	}

	stdout, _ := resMap["stdout"].(string)
	if strings.Contains(stdout, "RCE_SUCCESS") {
		t.Errorf("FAIL: RCE succeeded! Python code executed.")
	} else {
		t.Logf("RCE seemingly failed (output not found). Check stderr: %v", resMap["stderr"])
		t.Errorf("FAIL: Should have been blocked by security hardening.")
	}
}

func TestNodeCodeInjection(t *testing.T) {
	cmd := "node"

	// Tool: node -e "console.log('Hello {{name}}')"
	toolDef := v1.Tool_builder{Name: proto.String("node-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "console.log('Hello {{name}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("name")}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	// Payload: Break out and execute
	injection := "'); require('child_process').execSync('echo RCE_SUCCESS').toString(); console.log('"

    escapedInjection := strings.ReplaceAll(injection, "\"", "\\\"")

	req := &ExecutionRequest{
		ToolName: "node-tool",
		ToolInputs: []byte(fmt.Sprintf(`{"name": "%s"}`, escapedInjection)),
	}

	result, err := tool.Execute(context.Background(), req)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "security risk: template substitution is not allowed") {
			t.Logf("Success: Blocked by security hardening: %v", err)
			return
		}
		t.Fatalf("Execute failed with unexpected error: %v", err)
	}

	resMap, _ := result.(map[string]interface{})
	stdout, _ := resMap["stdout"].(string)
	if strings.Contains(stdout, "RCE_SUCCESS") {
		t.Errorf("FAIL: RCE succeeded! Node code executed.")
	} else {
		t.Errorf("FAIL: Should have been blocked by security hardening.")
	}
}
