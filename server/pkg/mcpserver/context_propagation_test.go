// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/consts"
	"github.com/mcpany/core/pkg/mcpserver"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream/factory"
	"github.com/mcpany/core/pkg/util"
	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestContextPropagation(t *testing.T) {
	// Setup Mock Upstream
	headerKey := "X-Custom-Header"
	headerVal := "test-value"

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, headerVal, r.Header.Get(headerKey))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer upstream.Close()

	// Setup MCP Server components
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	// Register Upstream Service with Context Propagation
	method := configv1.HttpCallDefinition_HTTP_METHOD_GET
	serviceConfig := &configv1.UpstreamServiceConfig{
		Name:           proto.String("test-service"),
		ConnectionPool: &configv1.ConnectionPoolConfig{},
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String(upstream.URL),
				Calls: map[string]*configv1.HttpCallDefinition{
					"my-call": {
						Method:       &method,
						EndpointPath: proto.String("/"),
					},
				},
			},
		},
		ContextPropagation: &configv1.ContextPropagationConfig{
			Headers: []string{headerKey},
		},
		AutoDiscoverTool: proto.Bool(true),
	}

	// Register service via factory/upstream logic
	upstreamImpl, err := factory.NewUpstream(serviceConfig)
	require.NoError(t, err)
	// Register returns tool definitions, but also adds them to toolManager
	_, _, _, err = upstreamImpl.Register(ctx, serviceConfig, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)

	// Inject middleware to simulate incoming HTTP request with headers
	injectMiddleware := func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			mockReq, _ := http.NewRequest("POST", "http://mock", nil)
			mockReq.Header.Set(headerKey, headerVal)
			// Inject http.request using the convention
			ctx = context.WithValue(ctx, consts.ContextKeyHTTPRequest, mockReq)
			return next(ctx, method, req)
		}
	}
	server.Server().AddReceivingMiddleware(injectMiddleware)

	// Connect Client
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// We need to connect the server session first
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "client"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Call tool
	sanitizedToolName, _ := util.SanitizeToolName("my-call")
	toolName := "test-service." + sanitizedToolName

	// Ensure tool exists
	listRes, err := clientSession.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	found := false
	for _, t := range listRes.Tools {
		if t.Name == toolName {
			found = true
			break
		}
	}
	require.True(t, found, "tool %s not found", toolName)

	_, err = clientSession.CallTool(ctx, &mcp.CallToolParams{Name: toolName})
	require.NoError(t, err)
}
