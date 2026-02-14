package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_PerlRCE_Unquoted(t *testing.T) {
	// This test demonstrates that Perl execution via "exec ls" is blocked.
	// Previously, it was allowed because spaces were permitted in unquoted context for interpreters.

	tool := v1.Tool_builder{
		Name: proto.String("perl-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
		Local:   proto.Bool(true),
	}.Build()

	// Perl script that takes unquoted input: -e {{input}}
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{input}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: exec ls
	// This uses only alphanumeric characters and space.
	payload := "exec ls"

	req := &ExecutionRequest{
		ToolName: "perl-tool",
		Arguments: map[string]interface{}{
			"input": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	// We expect this to fail with a security error
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "interpreter injection detected", "Expected interpreter injection error")
	}
}
