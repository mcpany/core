/*
 * Copyright 2025 Author(s) of MCP-XY
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
	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/pool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

type mockToolManager struct {
	tool.ToolManagerInterface
	tools       map[string]tool.Tool
	serviceInfo map[string]*tool.ServiceInfo
}

func (m *mockToolManager) AddTool(t tool.Tool) error {
	m.tools[t.Tool().GetName()] = t
	return nil
}

func (m *mockToolManager) GetTool(name string) (tool.Tool, bool) {
	t, ok := m.tools[name]
	return t, ok
}

func (m *mockToolManager) AddServiceInfo(serviceKey string, info *tool.ServiceInfo) {
	m.serviceInfo[serviceKey] = info
}

func (m *mockToolManager) GetServiceInfo(serviceKey string) (*tool.ServiceInfo, bool) {
	info, ok := m.serviceInfo[serviceKey]
	return info, ok
}

func newMockToolManager() *mockToolManager {
	return &mockToolManager{
		tools:       make(map[string]tool.Tool),
		serviceInfo: make(map[string]*tool.ServiceInfo),
	}
}

func TestGRPCUpstream_Register_WithProtoContent(t *testing.T) {
	protoContent := `
syntax = "proto3";
package test3;

service TestService3 {
  rpc TestMethod3(TestRequest3) returns (TestResponse3);
}

message TestRequest3 {
  string name = 1;
}

message TestResponse3 {
  string message = 1;
}
`
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		GrpcService: configv1.GrpcUpstreamService_builder{
			Address:      proto.String("localhost:50051"),
			UseReflection: proto.Bool(false),
			ProtoDefinitions: []*configv1.ProtoDefinition{
				configv1.ProtoDefinition_builder{
					ProtoFile: configv1.ProtoFile_builder{
						FileName: proto.String("test3.proto"),
						FileContent: proto.String(protoContent),
					}.Build(),
				}.Build(),
			},
		}.Build(),
	}.Build()

	tm := newMockToolManager()
	pm := pool.NewManager()
	upstream := NewGRPCUpstream(pm)

	serviceKey, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	// Check if the service info was added
	_, ok := tm.GetServiceInfo(serviceKey)
	assert.True(t, ok, "Service info should be added to the tool manager")

	// Check if the tool was added
	_, ok = tm.GetTool("TestMethod3")
	assert.True(t, ok, "Tool 'TestMethod3' should be added to the tool manager")
}
