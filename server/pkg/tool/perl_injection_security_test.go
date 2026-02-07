package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtrLocal(s string) *string { return &s }

func TestPerlInterpolationInjection(t *testing.T) {
	service := configv1.CommandLineUpstreamService_builder{
		Command: strPtrLocal("perl"),
	}.Build()

	// Perl double quoted string interpolation
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "print \"Hello, {{name}}!\\n\""},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: strPtrLocal("name"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := pb.Tool_builder{
		Name: strPtrLocal("perl_hello"),
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Payload: Perl array interpolation execution
	// @{[ system('echo RCE_SUCCESS') ]}
	// We check for "RCE_SUCCESS" in stdout.

	payload := "@{[ system('echo RCE_SUCCESS') ]}"

    inputMap := map[string]string{"name": payload}
    inputBytes, _ := json.Marshal(inputMap)

	req := &ExecutionRequest{
		ToolName: "perl_hello",
		ToolInputs: inputBytes,
	}

	result, err := tool.Execute(context.Background(), req)

	if err != nil {
		t.Logf("Execution blocked (good): %v", err)
		if !assert.Contains(t, err.Error(), "injection detected") {
			t.Errorf("Unexpected error: %v", err)
		}
	} else {
		resMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		stdout, ok := resMap["stdout"].(string)
		require.True(t, ok)

		t.Logf("Stdout: %s", stdout)

		if assert.Contains(t, stdout, "RCE_SUCCESS") {
			t.Fatal("VULNERABILITY CONFIRMED: Perl code injection successful")
		}
	}
}
