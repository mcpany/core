package tool

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestSentinel_InterpreterInjection_Lua_Execute(t *testing.T) {
	// We use "lua" as the command. Even if not installed, we test the input validation logic.
	cmd := "lua"
	toolDef := v1.Tool_builder{Name: proto.String("lua-tool")}.Build()

	// Create a service config for a lua command tool
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()

	// Define arguments: lua -e '{{code}}'
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "'{{code}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-lua")

	// Payload: os.execute("id")
	// JSON escape quotes: os.execute(\"id\")
	payload := `os.execute(\"id\")`

	req := &ExecutionRequest{
		ToolName:   "lua-tool",
		ToolInputs: []byte(`{"code": "` + payload + `"}`),
	}

	result, err := tool.Execute(context.Background(), req)

	if err != nil {
		// If the error is about "interpreter injection detected", then it's SECURE.
		if strings.Contains(err.Error(), "interpreter injection detected") {
			t.Logf("Secure: Blocked by input validation: %v", err)
			return
		}
		// If the error is about "executable file not found", it means validation PASSED -> VULNERABLE.
		if strings.Contains(err.Error(), "executable file not found") || strings.Contains(err.Error(), "no such file or directory") {
			t.Fatalf("VULNERABLE: Validation passed, attempted to execute lua! Error: %v", err)
		}

		t.Logf("Other error: %v", err)
	} else {
		resMap, _ := result.(map[string]interface{})
		t.Fatalf("VULNERABLE: Command executed successfully! Result: %v", resMap)
	}
}

func TestSentinel_InterpreterInjection_Python_Kill(t *testing.T) {
	cmd := "python"
	toolDef := v1.Tool_builder{Name: proto.String("python-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{Command: &cmd}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "'{{code}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-python-kill")

	// Payload: import os; os.kill(1, 9)
	payload := `import os; os.kill(1, 9)`

	req := &ExecutionRequest{
		ToolName:   "python-tool",
		ToolInputs: []byte(`{"code": "` + payload + `"}`),
	}

	result, err := tool.Execute(context.Background(), req)

	if err != nil {
		if strings.Contains(err.Error(), "interpreter injection detected") {
			t.Logf("Secure: Blocked by input validation: %v", err)
			return
		}
		if strings.Contains(err.Error(), "executable file not found") || strings.Contains(err.Error(), "no such file or directory") {
			t.Fatalf("VULNERABLE: Validation passed, attempted to execute python! Error: %v", err)
		}
		t.Logf("Other error: %v", err)
	} else {
		resMap, _ := result.(map[string]interface{})
		t.Fatalf("VULNERABLE: Command executed successfully! Result: %v", resMap)
	}
}
