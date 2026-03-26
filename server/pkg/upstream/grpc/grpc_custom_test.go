// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type mockToolManager struct {
	tool.ManagerInterface
	tools       map[string]tool.Tool
	serviceInfo map[string]*tool.ServiceInfo
}

// AddTool ...
// Summary: AddTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.tools[t.Tool().GetName()] = t
	return nil
}

// GetTool ...
// Summary: GetTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t, ok := m.tools[name]
	return t, ok
}

// AddServiceInfo ...
// Summary: AddServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.serviceInfo[serviceID] = info
}

// GetServiceInfo ...
// Summary: GetServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	info, ok := m.serviceInfo[serviceID]
	return info, ok
}

// ListServices ...
// Summary: ListServices
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	services := make([]*tool.ServiceInfo, 0, len(m.serviceInfo))
	for _, info := range m.serviceInfo {
		services = append(services, info)
	}
	return services
}

// SetProfiles ...
// Summary: SetProfiles
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

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

// AddPrompt ...
// Summary: AddPrompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.prompts[p.Prompt().Name] = p
}

func newMockPromptManager() *mockPromptManager {
	return &mockPromptManager{
		prompts: make(map[string]prompt.Prompt),
	}
}

// TestGRPCUpstream_Register_WithProtoContent ...
// Summary: TestGRPCUpstream_Register_WithProtoContent
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
