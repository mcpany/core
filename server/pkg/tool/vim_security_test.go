package tool

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_VimInjection_Prevention(t *testing.T) {
	t.Parallel()
	// This test verifies that editors like vim are treated as shells
	// because they allow command execution.

	tool := v1.Tool_builder{
		Name: proto.String("test-tool-vim"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("vim"),
		Local:   proto.Bool(true),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("file")}.Build()}.Build(),
		},
		Args: []string{"{{file}}"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Injection attempt: Use vim's +command syntax to execute a shell command
	// We use '!' which is a dangerous character in our shell injection check.
	// If 'vim' is detected as a shell, this should be blocked.
	reqAttack := &ExecutionRequest{
		ToolName: "test-tool-vim",
		Arguments: map[string]interface{}{
			"file": "+!ls",
		},
	}
	reqAttack.ToolInputs, _ = json.Marshal(reqAttack.Arguments)

	// Set a short timeout because if vim starts, it might hang.
	// We only care if the security check blocks it BEFORE execution.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := localTool.Execute(ctx, reqAttack)

	// Expect error "shell injection detected"
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "shell injection detected")
	}
}
