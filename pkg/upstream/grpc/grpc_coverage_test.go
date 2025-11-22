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

package grpc

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/resource"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestGRPCUpstream_Coverage(t *testing.T) {
	protoContent := `
syntax = "proto3";
package test;
service TestService {
  rpc GetData (GetDataRequest) returns (GetDataResponse);
}
message GetDataRequest {
  string query = 1;
}
message GetDataResponse {
  string result = 1;
}
`
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		GrpcService: configv1.GrpcUpstreamService_builder{
			Address: proto.String("localhost:50051"),
			ProtoDefinitions: []*configv1.ProtoDefinition{
				configv1.ProtoDefinition_builder{
					ProtoFile: configv1.ProtoFile_builder{
						FileName:    proto.String("test.proto"),
						FileContent: proto.String(protoContent),
					}.Build(),
				}.Build(),
			},
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("manual-tool"),
					CallId: proto.String("call1"),
				}.Build(),
			},
			Calls: map[string]*configv1.GrpcCallDefinition{
				"call1": configv1.GrpcCallDefinition_builder{
					Id:      proto.String("call1"),
					Service: proto.String("test.TestService"),
					Method:  proto.String("GetData"),
				}.Build(),
			},
			Prompts: []*configv1.PromptDefinition{
				configv1.PromptDefinition_builder{
					Name: proto.String("test-prompt"),
					Messages: []*configv1.PromptMessage{
						configv1.PromptMessage_builder{
							Role: configv1.PromptMessage_USER.Enum(),
							Text: configv1.TextContent_builder{
								Text: proto.String("Hello {{name}}"),
							}.Build(),
						}.Build(),
					},
				}.Build(),
			},
		}.Build(),
	}.Build()

	poolManager := pool.NewManager()
	upstream := NewGRPCUpstream(poolManager)
	tm := newMockToolManager()
	pm := newMockPromptManager()
	var rm resource.ResourceManagerInterface // nil is fine if not used

	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, pm, rm, false)
	require.NoError(t, err)
	assert.NotEmpty(t, serviceID)

	// Check "manual-tool"
	manualTool, ok := tm.GetTool("manual-tool")
	assert.True(t, ok, "manual-tool should be registered")
	assert.NotNil(t, manualTool)

	// Check "GetData" (from descriptor)
	getDataTool, ok := tm.GetTool("GetData")
	assert.True(t, ok, "GetData tool should be registered from descriptor")
	assert.NotNil(t, getDataTool)

	// Verify Prompts
	_, ok = pm.prompts["test-prompt"]
	assert.True(t, ok, "test-prompt should be registered")
}
