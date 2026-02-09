package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestNiceBypass(t *testing.T) {
	// Setup a LocalCommandTool that uses "nice"
	toolDef := v1.Tool_builder{
		Name: proto.String("nice_bypass"),
	}.Build()

	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("nice"),
	}).Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"sh", "-c", "{{cmd}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("cmd"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	// Initialize the tool
	cmdTool := NewLocalCommandTool(toolDef, service, callDef, nil, "call-id")

	// Payload that contains shell metacharacters which SHOULD be blocked if checks were applied
	payload := "echo hacked; echo pwned"

	// Execute
	req := &ExecutionRequest{
		ToolName: "nice_bypass",
		ToolInputs: []byte(`{"cmd": "` + payload + `"}`),
	}

	// This should return an error if security checks are working.
	_, err := cmdTool.Execute(context.Background(), req)

	if err == nil {
		t.Log("Vulnerability confirmed: 'nice' bypassed shell injection checks")
        t.Fail()
	} else {
		t.Logf("Result: %v", err)
		if assert.Contains(t, err.Error(), "shell injection detected") {
			t.Log("Vulnerability Mitigated: Blocked as expected")
		} else {
            t.Log("Vulnerability confirmed: Error was not about shell injection")
            t.Fail()
        }
	}
}
