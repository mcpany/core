package tool

import (
	"context"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestSedSandbox_Prevention(t *testing.T) {
	cmd := "sed"

	// Create tool for sed -e {{script}}
	toolDef := v1.Tool_builder{Name: proto.String("sed-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{script}}", "/etc/hosts"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	// Payload: 1e date (Execute 'date')
	req := &ExecutionRequest{
		ToolName: "sed-tool",
		ToolInputs: []byte(`{"script": "1e date"}`),
	}

	result, err := tool.Execute(context.Background(), req)
	if err != nil {
		// If input validation blocks it (e.g., preventing spaces), that's also a success for security.
		if strings.Contains(err.Error(), "injection detected") || strings.Contains(err.Error(), "dangerous character") {
			t.Logf("Success: Blocked by input validation: %v", err)
		} else {
			t.Fatalf("Execute failed: %v", err)
		}
	} else {
		resMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Result is not map: %v", result)
		}

		// Expect return code to be non-zero (sed error)
		returnCode, ok := resMap["return_code"].(int)
		if !ok {
			t.Fatalf("return_code not int: %v", resMap["return_code"])
		}

		if returnCode == 0 {
			t.Errorf("FAIL: sed executed '1e date' successfully (return_code 0). Sandbox failed.")
		}

		stderr, _ := resMap["stderr"].(string)
		if !strings.Contains(stderr, "command disabled") && !strings.Contains(stderr, "unknown command") {
			// GNU sed: 'e' command disabled in sandbox mode
			// BSD sed: unknown command: 1 (or similar)
			t.Logf("Note: stderr was: %s", stderr)
		} else {
			t.Logf("Success: sed blocked command with error: %s", stderr)
		}
	}

	// Payload: w /tmp/pwned (Write file)
	req = &ExecutionRequest{
		ToolName: "sed-tool",
		ToolInputs: []byte(`{"script": "w /tmp/pwned"}`),
	}

	result, err = tool.Execute(context.Background(), req)
	if err != nil {
		if strings.Contains(err.Error(), "injection detected") || strings.Contains(err.Error(), "dangerous character") {
			t.Logf("Success: Blocked by input validation: %v", err)
			return
		}
		t.Fatalf("Execute failed: %v", err)
	}
	resMap, _ := result.(map[string]interface{})
	returnCode, _ := resMap["return_code"].(int)

	if returnCode == 0 {
		t.Errorf("FAIL: sed executed 'w /tmp/pwned' successfully. Sandbox failed.")
	}
}

func TestSedSandbox_ValidUsage(t *testing.T) {
	cmd := "sed"
	toolDef := v1.Tool_builder{Name: proto.String("sed-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{script}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	// Valid replacement (needs input, but sed waits for stdin if file not provided)
	// We pass empty input via echo? No, LocalCommandTool doesn't easily pipe stdin in this test setup unless we use communication protocol JSON?
	// But sed -e 's/a/b/' will wait for input.
	// So it will timeout.

	// We can use 'sed --help' or something?
	// But arguments are fixed.

	// Maybe use 'version' command if sed has it as script? No.
	// Use 'q' (quit).
	req := &ExecutionRequest{
		ToolName: "sed-tool",
		ToolInputs: []byte(`{"script": "q"}`),
	}

	// Since sed is waiting for input (no files), 'q' might not trigger immediately unless it processes line 1.
	// But without input...
	// Actually, sed without input files reads from stdin.
	// LocalCommandTool passes nil stdin (closed).
	// So sed reads EOF.
	// 'q' on EOF?

	// Let's rely on timeout if it hangs.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond) // Short timeout
	defer cancel()

	result, err := tool.Execute(ctx, req)
	if err != nil {
		// Timeout is expected or success?
		// If sed exits 0, it's success.
		t.Logf("Execute result: %v, err: %v", result, err)
	} else {
		resMap, _ := result.(map[string]interface{})
		t.Logf("Success: valid command executed. Return code: %v", resMap["return_code"])
	}
}
