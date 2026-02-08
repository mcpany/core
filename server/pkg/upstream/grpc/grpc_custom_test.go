package grpc

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type mockToolManager struct {
	tool.ManagerInterface
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

func (m *mockToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	m.serviceInfo[serviceID] = info
}

func (m *mockToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	info, ok := m.serviceInfo[serviceID]
	return info, ok
}

func (m *mockToolManager) ListServices() []*tool.ServiceInfo {
	services := make([]*tool.ServiceInfo, 0, len(m.serviceInfo))
	for _, info := range m.serviceInfo {
		services = append(services, info)
	}
	return services
}

func (m *mockToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

func newMockToolManager() *mockToolManager {
	return &mockToolManager{
		tools:       make(map[string]tool.Tool),
		serviceInfo: make(map[string]*tool.ServiceInfo),
	}
}

type mockPromptManager struct {
	prompt.ManagerInterface
	prompts map[string]prompt.Prompt
}

func (m *mockPromptManager) AddPrompt(p prompt.Prompt) {
	m.prompts[p.Prompt().Name] = p
}

func newMockPromptManager() *mockPromptManager {
	return &mockPromptManager{
		prompts: make(map[string]prompt.Prompt),
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
			Address:       proto.String("127.0.0.1:50051"),
			UseReflection: proto.Bool(false),
			ProtoDefinitions: []*configv1.ProtoDefinition{
				configv1.ProtoDefinition_builder{
					ProtoFile: configv1.ProtoFile_builder{
						FileName:    proto.String("test3.proto"),
						FileContent: proto.String(protoContent),
					}.Build(),
				}.Build(),
			},
		}.Build(),
	}.Build()

	tm := newMockToolManager()
	pm := pool.NewManager()
	mockPromptManager := newMockPromptManager()
	upstream := NewUpstream(pm)

	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, mockPromptManager, nil, false)
	require.NoError(t, err)

	// Check if the service info was added
	_, ok := tm.GetServiceInfo(serviceID)
	assert.True(t, ok, "Service info should be added to the tool manager")

	// Check if the tool was added
	_, ok = tm.GetTool("TestMethod3")
	assert.True(t, ok, "Tool 'TestMethod3' should be added to the tool manager")
}
