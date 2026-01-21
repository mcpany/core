/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

package mcpserver_test

import (
	"context"
	"encoding/json"
	"testing"

	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type spyToolManager struct {
	tool.Manager
	lastExecReq *tool.ExecutionRequest
}

// Stubs
func (m *spyToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}
func (m *spyToolManager) GetTool(_ string) (tool.Tool, bool)           { return nil, false }
func (m *spyToolManager) ListTools() []tool.Tool                       { return nil }
func (m *spyToolManager) ListMCPTools() []*mcp.Tool                    { return nil }
func (m *spyToolManager) ListServices() []*tool.ServiceInfo            { return nil }
func (m *spyToolManager) AddMiddleware(_ tool.ExecutionMiddleware)     {}
func (m *spyToolManager) SetMCPServer(_ tool.MCPServerProvider)        {}
func (m *spyToolManager) AddTool(_ tool.Tool) error                    { return nil }
func (m *spyToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) {
	return nil, false
}
func (m *spyToolManager) ClearToolsForService(_ string)                                {}
func (m *spyToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition)      {}
func (m *spyToolManager) IsServiceAllowed(_, _ string) bool                            { return true }
func (m *spyToolManager) GetAllowedServiceIDs(_ string) (map[string]bool, bool)        { return nil, false }

func (m *spyToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	m.lastExecReq = req
	return "success", nil
}

func TestServer_DryRun(t *testing.T) {
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager, nil)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	spyTM := &spyToolManager{}
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(factory, spyTM, promptManager, resourceManager, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, spyTM, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	router := server.GetRouter()
	handler, ok := router.GetHandler(consts.MethodToolsCall)
	require.True(t, ok)

	// Case 1: No dry run
	args := map[string]any{"arg1": "value"}
	argsBytes, _ := json.Marshal(args)
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "test-tool",
			Arguments: argsBytes,
		},
	}

	_, err = handler(ctx, req)
	require.NoError(t, err)
	assert.False(t, spyTM.lastExecReq.DryRun)
	assert.JSONEq(t, string(argsBytes), string(spyTM.lastExecReq.ToolInputs))

	// Case 2: Dry run with _dry_run=true
	argsDry := map[string]any{"arg1": "value", "_dry_run": true}
	argsDryBytes, _ := json.Marshal(argsDry)
	reqDry := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "test-tool",
			Arguments: argsDryBytes,
		},
	}

	_, err = handler(ctx, reqDry)
	require.NoError(t, err)
	assert.True(t, spyTM.lastExecReq.DryRun)

	// Verify _dry_run is removed from ToolInputs
	// Expected inputs should only contain arg1
	expectedArgs := map[string]any{"arg1": "value"}
	expectedBytes, _ := json.Marshal(expectedArgs)
	assert.JSONEq(t, string(expectedBytes), string(spyTM.lastExecReq.ToolInputs))
}
