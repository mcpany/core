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

func TestLocalCommandTool_Git_C_Injection_Repro(t *testing.T) {
	// This test attempts to demonstrate that git allows command execution via -c configuration.
	// Specifically core.sshCommand.

	toolProto := v1.Tool_builder{
		Name: proto.String("git-ssh-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		Local:   proto.Bool(true),
	}.Build()

	// git -c core.sshCommand={{ssh}} fetch ssh://localhost/repo.git
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "core.sshCommand={{ssh}}", "fetch", "ssh://localhost/repo.git"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("ssh"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// The payload: sh -c "echo PWNED >&2; exit 1"
	// This replaces the ssh command.
	payload := "sh -c 'echo PWNED >&2; exit 1'"
	// We need to escape quotes for JSON
	inputs := fmt.Sprintf(`{"ssh": "%s"}`, payload)

	req := &ExecutionRequest{
		ToolName:   "git-ssh-tool",
		ToolInputs: []byte(inputs),
	}

	result, err := tool.Execute(context.Background(), req)

    // Expected behavior: execution blocked due to shell injection detection
    if err == nil {
        t.Errorf("Expected execution error, got nil")
    } else {
        if !strings.Contains(err.Error(), "shell injection detected") {
             t.Errorf("Expected 'shell injection detected' error, got: %v", err)
        } else {
             t.Logf("Security check triggered successfully: %v", err)
        }
    }

	var output string
	if result != nil {
		resultMap, ok := result.(map[string]interface{})
		if ok {
			if combined, ok := resultMap["combined_output"].(string); ok {
				output += combined
			}
		}
	} else if err != nil {
        // Sometimes error contains output?
        output += err.Error()
    }

	if strings.Contains(output, "PWNED") {
		t.Errorf("VULNERABILITY CONFIRMED: RCE via git -c core.sshCommand. Output: %s", output)
	}
}
