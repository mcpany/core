// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"unsafe"

	configv1 "github.com/mcpany/core/proto/config/v1"
	routerv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// mockToolSimple is a simple mock for Tool interface
type mockToolSimple struct {
	executeFunc func(ctx context.Context, req *ExecutionRequest) (any, error)
	toolDef     *routerv1.Tool
	serviceID   string
}

func (m *mockToolSimple) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, req)
	}
	return "success", nil
}

func (m *mockToolSimple) Tool() *routerv1.Tool {
	if m.toolDef == nil {
		inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
		return routerv1.Tool_builder{
			Name:        proto.String("mock-tool"),
			ServiceId:   proto.String(m.serviceID),
			InputSchema: inputSchema,
		}.Build()
	}
	return m.toolDef
}

func (m *mockToolSimple) GetCacheConfig() *configv1.CacheConfig { return nil }

func (m *mockToolSimple) MCPTool() *mcp.Tool {
	t, _ := ConvertProtoToMCPTool(m.Tool())
	return t
}

type mockMiddleware struct {
	id     string
	called bool
}

func (m *mockMiddleware) Execute(ctx context.Context, req *ExecutionRequest, next ExecutionFunc) (any, error) {
	m.called = true
	return next(ctx, req)
}

func TestManager_ExecuteTool_Coverage(t *testing.T) {
	t.Parallel()

	b, _ := bus.NewProvider(nil)
	m := NewManager(b)

	// Case 1: Tool Not Found
	_, err := m.ExecuteTool(context.Background(), &ExecutionRequest{ToolName: "missing"})
	assert.Error(t, err)
	assert.Equal(t, ErrToolNotFound, err)

	// Add a tool
	mt := &mockToolSimple{serviceID: "s1"}
	_ = m.AddTool(mt)

	// Case 2: Success
	res, err := m.ExecuteTool(context.Background(), &ExecutionRequest{ToolName: "s1.mock-tool"})
	assert.NoError(t, err)
	assert.Equal(t, "success", res)

	// Case 3: Middleware execution
	mw := &mockMiddleware{id: "m1"}
	m.AddMiddleware(mw)
	_, err = m.ExecuteTool(context.Background(), &ExecutionRequest{ToolName: "s1.mock-tool"})
	assert.NoError(t, err)
	assert.True(t, mw.called)
}

func TestManager_ExecuteTool_Hooks_Coverage(t *testing.T) {
	t.Parallel()

	b, _ := bus.NewProvider(nil)
	m := NewManager(b)
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	mt := &mockToolSimple{
		serviceID: "s1",
		toolDef: routerv1.Tool_builder{
			Name:        proto.String("mock-tool"),
			ServiceId:   proto.String("s1"),
			InputSchema: inputSchema,
		}.Build(),
	}
	_ = m.AddTool(mt)

	// Case: PreHook Error (via Policy)
	callPolicy := configv1.CallPolicy_builder{
		DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
		Rules: []*configv1.CallPolicyRule{
			configv1.CallPolicyRule_builder{
				NameRegex: proto.String(".*mock-tool$"),
				Action:    configv1.CallPolicy_DENY.Enum(),
			}.Build(),
		},
	}.Build()

	// Setup ServiceInfo with Config
	svcConfig := configv1.UpstreamServiceConfig_builder{
		Id:           proto.String("s1"),
		Name:         proto.String("s1"),
		CallPolicies: []*configv1.CallPolicy{callPolicy},
	}.Build()

	m.AddServiceInfo("s1", &ServiceInfo{Config: svcConfig})

	_, err := m.ExecuteTool(context.Background(), &ExecutionRequest{ToolName: "s1.mock-tool"})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "denied by policy rule")
	}
}

func TestManager_ClearToolsForService_Coverage(t *testing.T) {
	t.Parallel()

	b, _ := bus.NewProvider(nil)
	m := NewManager(b)

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	t1 := &mockToolSimple{serviceID: "s1", toolDef: routerv1.Tool_builder{Name: proto.String("t1"), ServiceId: proto.String("s1"), InputSchema: inputSchema}.Build()}
	t2 := &mockToolSimple{serviceID: "s2", toolDef: routerv1.Tool_builder{Name: proto.String("t2"), ServiceId: proto.String("s2"), InputSchema: inputSchema}.Build()}
	t3 := &mockToolSimple{serviceID: "s1", toolDef: routerv1.Tool_builder{Name: proto.String("t3"), ServiceId: proto.String("s1"), InputSchema: inputSchema}.Build()}

	_ = m.AddTool(t1)
	_ = m.AddTool(t2)
	_ = m.AddTool(t3)

	assert.Len(t, m.ListTools(), 3)

	m.ClearToolsForService("s1")

	tools := m.ListTools()
	assert.Len(t, tools, 1)
	assert.Equal(t, "t2", tools[0].Tool().GetName())
}

// Mock MCP Server
type mockMCPServerProvider struct {
	server *mcp.Server
}

func (m *mockMCPServerProvider) Server() *mcp.Server {
	return m.server
}

func TestManager_AddTool_WithMCPServer_Coverage(t *testing.T) {
	t.Parallel()

	b, _ := bus.NewProvider(nil)
	m := NewManager(b)

	// Setup Mock MCP Server
	impl := &mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	mcpServer := mcp.NewServer(impl, &mcp.ServerOptions{})

	mp := &mockMCPServerProvider{server: mcpServer}
	m.SetMCPServer(mp)

	mt := &mockToolSimple{
		serviceID: "s1",
		toolDef: routerv1.Tool_builder{
			Name:      proto.String("mock-tool"),
			ServiceId: proto.String("s1"),
			InputSchema: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"type": {Kind: &structpb.Value_StringValue{StringValue: "object"}},
					"properties": {Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"arg": {Kind: &structpb.Value_StringValue{StringValue: "val"}},
						},
					}}},
				},
			},
		}.Build(),
	}

	err := m.AddTool(mt)
	assert.NoError(t, err)

	// Coverage: OutputSchema
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	mt.toolDef = routerv1.Tool_builder{
		Name:      proto.String("mock-tool"),
		ServiceId: proto.String("s1"),
		InputSchema: inputSchema,
		OutputSchema: &structpb.Struct{Fields: map[string]*structpb.Value{
			"type": {Kind: &structpb.Value_StringValue{StringValue: "object"}},
		}},
	}.Build()

	err = m.AddTool(mt)
	assert.NoError(t, err)

	// Use reflection to get the handler and call it!
	srvVal := reflect.ValueOf(mcpServer).Elem()
	toolsFieldSet := srvVal.FieldByName("tools")
	if !toolsFieldSet.IsValid() {
		t.Fatalf("Could not find tools field in mcp.Server")
	}
	toolsMap := toolsFieldSet.Elem().FieldByName("features")
	if !toolsMap.IsValid() {
		t.Fatalf("Could not find features field in featureSet")
	}

	toolHandlerVal := toolsMap.MapIndex(reflect.ValueOf("s1.mock-tool"))
	if !toolHandlerVal.IsValid() {
		t.Fatalf("Could not find tool handler for s1.mock-tool")
	}
	handlerField := toolHandlerVal.Elem().FieldByName("handler")
	if !handlerField.IsValid() {
		t.Fatalf("Could not find handler field in serverTool")
	}

	handlerAccessible := reflect.NewAt(handlerField.Type(), unsafe.Pointer(handlerField.UnsafeAddr())).Elem() //nolint:gosec

	// Prepare request
	callToolReq := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "s1.mock-tool",
			Arguments: json.RawMessage(`{"arg":"val"}`),
		},
	}

	// Subscribe to bus to verify publication
	reqBus, err := bus.GetBus[*bus.ToolExecutionRequest](b, "tool_execution_requests")
	assert.NoError(t, err)
	reqChan := make(chan *bus.ToolExecutionRequest, 1)
	reqBus.SubscribeOnce(context.Background(), "request", func(req *bus.ToolExecutionRequest) {
		reqChan <- req
	})

	// Invoke handler
	go func() {
		args := []reflect.Value{
			reflect.ValueOf(context.Background()),
			reflect.ValueOf(callToolReq),
		}
		_ = handlerAccessible.Call(args)
	}()

	select {
	case req := <-reqChan:
		assert.Equal(t, "s1.mock-tool", req.ToolName)
		resJSON, _ := json.Marshal(map[string]any{"result": "ok"})

		resBus, err := bus.GetBus[*bus.ToolExecutionResult](b, "tool_execution_results")
		assert.NoError(t, err)
		_ = resBus.Publish(context.Background(), req.CorrelationID(), &bus.ToolExecutionResult{
			Result: json.RawMessage(resJSON),
		})
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for execution request on bus")
	}
}
