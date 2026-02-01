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

func TestLocalCommandTool_ArgumentReplacement(t *testing.T) {
	tool := v1.Tool_builder{
		Name: proto.String("test-arg-replacement"),
	}.Build()

	// Use echo to verify what arguments it receives
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		Local:   proto.Bool(true),
	}.Build()

	// Argument with two placeholders
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{p1}}", "{{p2}}", "{{p1}}{{p2}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("p1")}.Build(),
			}.Build(),
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("p2")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "repro-call-id")

	req := &ExecutionRequest{
		ToolName: "test-arg-replacement",
		Arguments: map[string]interface{}{
			"p1": "AAA",
			"p2": "BBB",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	output, ok := resultMap["stdout"].(string)
	assert.True(t, ok)

	// Expect proper replacement of BOTH.
	assert.Contains(t, output, "AAABBB")
}

func TestLocalCommandTool_ArgumentReplacement_NoRecursiveInjection(t *testing.T) {
	tool := v1.Tool_builder{
		Name: proto.String("test-arg-recursive"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{p1}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("p1")}.Build(),
			}.Build(),
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("p2")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "recursive-call-id")

	req := &ExecutionRequest{
		ToolName: "test-arg-recursive",
		Arguments: map[string]interface{}{
			"p1": "{{p2}}",
			"p2": "SECRET",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	output, ok := resultMap["stdout"].(string)
	assert.True(t, ok)

	// Expect literal "{{p2}}" not "SECRET"
	assert.Contains(t, output, "{{p2}}")
	assert.NotContains(t, output, "SECRET")
}
