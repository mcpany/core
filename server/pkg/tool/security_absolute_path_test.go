package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
    "google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_AbsolutePath_Vulnerability(t *testing.T) {
    // This test demonstrates that absolute paths are currently allowed,
    // which effectively bypasses any working directory restrictions.

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"args": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := &v1.Tool{
		Name:        proto.String("test-cat"),
		Description: proto.String("Cat tool"),
        InputSchema: inputSchema,
	}
	service := &configv1.CommandLineUpstreamService{}
	service.Command = proto.String("cat")
	service.Local = proto.Bool(true)
    // We set a working directory to show that absolute path ignores it
    service.WorkingDirectory = proto.String("/tmp")

	callDef := &configv1.CommandLineCallDefinition{
		Parameters: []*configv1.CommandLineParameterMapping{
			{Schema: &configv1.ParameterSchema{Name: proto.String("args")}},
		},
	}

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

    // Try to access /etc/hosts using absolute path
	req := &ExecutionRequest{
		ToolName: "test-cat",
		Arguments: map[string]interface{}{
			"args": []interface{}{"/etc/hosts"},
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)

    // We expect this to fail now
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "absolute path not allowed")
	assert.Nil(t, result)
}
