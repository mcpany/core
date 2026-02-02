package tool

import (
	"context"
	"encoding/json"
	"testing"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestSentinel_ShellInjection_UnquotedSpace_Bash(t *testing.T) {
	t.Parallel()
	tool := v1.Tool_builder{
		Name: proto.String("test-bash-space"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("bash"),
		Local:   proto.Bool(true),
	}.Build()

	// args: ["-c", "{{msg}}"]
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build()}.Build(),
		},
		Args: []string{"-c", "{{msg}}"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "test-bash-space",
		Arguments: map[string]interface{}{
			"msg": "echo hello",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// This should FAIL because bash is a shell and space is dangerous in unquoted context (sh -c)
	_, err := localTool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shell injection detected")
}

func TestSentinel_ShellInjection_UnquotedSpace_Git(t *testing.T) {
	t.Parallel()
	tool := v1.Tool_builder{
		Name: proto.String("test-git-space"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"), // git is in the shell list
		Local:   proto.Bool(true),
	}.Build()

	// args: ["commit", "-m", "{{msg}}"]
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build()}.Build(),
		},
		Args: []string{"commit", "-m", "{{msg}}"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "test-git-space",
		Arguments: map[string]interface{}{
			"msg": "initial commit",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// This should NOT fail with "shell injection detected".
	// It might succeed or fail depending on environment (e.g. if git is installed and cwd is a repo).
	_, err := localTool.Execute(context.Background(), req)
	if err != nil {
		assert.False(t, strings.Contains(err.Error(), "shell injection detected"), "Should not detect shell injection for spaces")
	}
}

func TestSentinel_ShellInjection_Semicolon_Blocked(t *testing.T) {
	t.Parallel()
	tool := v1.Tool_builder{
		Name: proto.String("test-bash-semicolon"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("bash"),
		Local:   proto.Bool(true),
	}.Build()

	// args: ["-c", "{{msg}}"]
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build()}.Build(),
		},
		Args: []string{"-c", "{{msg}}"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "test-bash-semicolon",
		Arguments: map[string]interface{}{
			"msg": "echo hello; rm -rf /",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// This MUST fail
	_, err := localTool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shell injection detected")
}
