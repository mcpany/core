package tool

import (
	"context"
	"fmt"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
    "github.com/stretchr/testify/assert"
)

func TestLocalCommandTool_Git_C_Injection_Security(t *testing.T) {
	// This test verifies that git command execution via -c configuration is protected
    // against shell injection by treating git as a shell command for template substitution.

	toolProto := v1.Tool_builder{
		Name: proto.String("git-ssh-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		Local:   proto.Bool(true),
	}.Build()

	// git -c core.sshCommand={{ssh}} ls-remote ssh://user@localhost:12345/repo
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "core.sshCommand={{ssh}}", "ls-remote", "ssh://user@localhost:12345/repo"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("ssh"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Payload attempting to inject a shell command via spaces
	payload := "sh -c 'echo VULN >&2'"
	// We need to escape quotes for JSON
	inputs := fmt.Sprintf(`{"ssh": "%s"}`, payload)

	req := &ExecutionRequest{
		ToolName:   "git-ssh-tool",
		ToolInputs: []byte(inputs),
	}

	result, err := tool.Execute(context.Background(), req)

    // We expect an error because spaces are blocked in unquoted context for shell commands
    assert.Error(t, err)
    assert.Nil(t, result)
    if err != nil {
        assert.Contains(t, err.Error(), "shell injection detected")
        assert.Contains(t, err.Error(), "contains dangerous character ' '")
    }
}
