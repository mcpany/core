
package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_Execute(t *testing.T) {
	tool := &v1.Tool{
		Name:        proto.String("test-tool"),
		Description: proto.String("A test tool"),
	}
	service := &configv1.CommandLineUpstreamService{}
	service.SetCommand("echo")
	service.SetLocal(true)
	callDef := &configv1.CommandLineCallDefinition{}

	localTool := NewLocalCommandTool(tool, service, callDef)

	req := &ExecutionRequest{
		ToolName: "test-tool",
		Arguments: map[string]interface{}{
			"args": []interface{}{"hello", "world"},
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "hello world\n", resultMap["stdout"])
}
