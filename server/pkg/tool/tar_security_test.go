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

func TestTarInjection_Vulnerability(t *testing.T) {
	// This test demonstrates that "tar" IS NOW treated as a dangerous command,
	// preventing argument injection that can lead to RCE via --checkpoint-action.

	// Verify tar is considered a shell command (interpreter)
	assert.True(t, isShellCommand("tar"), "tar should be considered a shell command (interpreter)")

	toolDef := v1.Tool_builder{
		Name: proto.String("tar-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("tar"),
		Local:   proto.Bool(true),
	}.Build()

	// Configuration where the user input is substituted into a flag
	callDefVulnerable := configv1.CommandLineCallDefinition_builder{
		Args: []string{"cf", "archive.tar", "--checkpoint-action={{action}}", "file.txt"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("action")}.Build(),
			}.Build(),
		},
	}.Build()

	localToolVulnerable := NewLocalCommandTool(toolDef, service, callDefVulnerable, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "tar-tool",
		Arguments: map[string]interface{}{
			"action": "exec=sh", // Malicious payload
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute
	_, err := localToolVulnerable.Execute(context.Background(), req)

	// Expect security error
	assert.Error(t, err)
	if err != nil {
		isSecurityError := strings.Contains(err.Error(), "injection detected") || strings.Contains(err.Error(), "blocked")
		assert.True(t, isSecurityError, "Security check should trigger for tar injection")
		t.Logf("Got expected security error: %v", err)
	}
}
