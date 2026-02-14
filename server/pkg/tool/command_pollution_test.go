package tool_test

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestCommandTool_ParameterPollution(t *testing.T) {
	// Define a tool with ONE allowed parameter "allowed".
	// But the args template uses "{{allowed}}" AND "{{hidden}}".
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{allowed}}", "{{hidden}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("allowed")}.Build(),
			}.Build(),
		},
	}.Build()

	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
	}).Build()

	properties := make(map[string]*structpb.Value)
	properties["allowed"] = structpb.NewStructValue(&structpb.Struct{})

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: properties,
			}),
		},
	}

	cmdTool := tool.NewLocalCommandTool(
		v1.Tool_builder{InputSchema: inputSchema}.Build(),
		service,
		callDef,
		nil,
		"call-id",
	)

	// User sends "hidden" parameter which is NOT in the schema.
	inputData := map[string]interface{}{
		"allowed": "safe",
		"hidden":  "injected",
	}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)
	req := &tool.ExecutionRequest{ToolInputs: inputs}

	result, err := cmdTool.Execute(context.Background(), req)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	// Since we fixed the vulnerability, "injected" should NOT appear.
	// The placeholder {{hidden}} should remain unsubstituted.
	args := resultMap["args"].([]string)
	assert.Contains(t, args, "safe")
	assert.NotContains(t, args, "injected", "Vulnerability fixed: 'hidden' parameter was ignored because it is not in the schema")
	assert.Contains(t, args, "{{hidden}}", "Placeholder should remain unsubstituted")
}
