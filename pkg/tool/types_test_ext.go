/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
