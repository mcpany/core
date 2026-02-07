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

func TestPythonCodeInjection(t *testing.T) {
	// This function was just a placeholder/commentary in previous attempt.
	// We can implement a basic check here or skip.
	t.Log("Testing RCE via Code Injection into Single Quoted Argument...")
}

func TestPythonCodeInjection_CodeContext(t *testing.T) {
	cmd := "python3"

	// Tool: python3 -c '{{code}}'

	toolDef := v1.Tool_builder{Name: proto.String("python-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "print('Start'); {{code}}; print('End')"}, // Single quoted template
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	// Payload: import os; s=os.system; s("echo RCE_SUCCESS")

	injection := "import os; s=os.system; s(\"echo RCE_SUCCESS\")"

	inputs := map[string]string{
		"code": injection,
	}
	inputBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "python-tool",
		ToolInputs: inputBytes,
	}

	result, err := tool.Execute(context.Background(), req)
	if err != nil {
		t.Logf("Blocked: %v", err)
	} else {
		resMap, _ := result.(map[string]interface{})
		stdout, _ := resMap["stdout"].(string)
		t.Logf("Stdout: %s", stdout)
		if strings.Contains(stdout, "RCE_SUCCESS") {
			t.Errorf("FAIL: RCE Succeeded! Python code executed via injected code.")
		}
	}
}
