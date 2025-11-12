
package tool

import (
	"context"
	"encoding/json"
	"testing"

	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestOpenAPITool_GetCacheConfig(t *testing.T) {
	tool := &OpenAPITool{}
	assert.Nil(t, tool.GetCacheConfig(), "GetCacheConfig should return nil")
}

func TestOpenAPITool_Tool(t *testing.T) {
	toolProto := &v1.Tool{}
	toolProto.SetName("test-tool")
	tool := &OpenAPITool{tool: toolProto}
	assert.Equal(t, toolProto, tool.Tool(), "Tool() should return the tool proto")
}

func TestOpenAPITool_Execute_InvalidInputSchema(t *testing.T) {
	toolProto := &v1.Tool{}
	toolProto.SetName("test-tool")
	toolProto.SetInputSchema(&structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type":    structpb.NewStringValue("object"),
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"param": structpb.NewStructValue(&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"type": structpb.NewStringValue("invalid-type"),
						},
					}),
				},
			}),
		},
	})
	tool := &OpenAPITool{tool: toolProto}
	req := &ExecutionRequest{
		ToolInputs: json.RawMessage(`{"param":"value"}`),
	}
	_, err := tool.Execute(context.Background(), req)
	assert.Error(t, err, "Execute should fail with an invalid input schema")
}

func TestWebsocketTool_GetCacheConfig(t *testing.T) {
	tool := &WebsocketTool{}
	assert.Nil(t, tool.GetCacheConfig(), "GetCacheConfig should return nil")
}
