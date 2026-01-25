// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package upstream

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/tests/integration"
	configv1 "github.com/mcpany/core/proto/config/v1"
	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHTTPInvalidQueryOverrideBug(t *testing.T) {
	// 1. Start Mock Server
	var capturedURL string
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer mockServer.Close()

	// 2. Start MCP Any Server
	serverInfo := integration.StartMCPANYServer(t, "InvalidQueryBug")
	defer serverInfo.CleanupFunc()

	// 3. Register Service
    // Base URL has invalid query param: ?q=invalid%
	baseURL := mockServer.URL + "?q=invalid%"
    // Endpoint URL overrides it: ?q=valid
	endpointPath := "/test?q=valid"

    toolName := "test-tool"
    callID := "call-test"

    // Construct config manually
    toolDef := configv1.ToolDefinition_builder{
		Name: proto.String(toolName),
        CallId: proto.String(callID),
	}.Build()

    callDef := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String(endpointPath),
		Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
	}.Build()

    serviceName := "bug-service"
	upstreamServiceConfigBuilder := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(serviceName),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(baseURL),
			Tools:   []*configv1.ToolDefinition{toolDef},
			Calls:   map[string]*configv1.HttpCallDefinition{callID: callDef},
		}.Build(),
	}
    config := upstreamServiceConfigBuilder.Build()
    req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

    integration.RegisterServiceViaAPI(t, serverInfo.RegistrationClient, req)

	// 4. Initialize Client
    err := serverInfo.Initialize(context.Background())
    require.NoError(t, err)

	// 5. Call Tool
    // Let's verify tool list first
    tools, err := serverInfo.ListTools(context.Background())
    require.NoError(t, err)

    toolFound := false
    fullToolName := serviceName + "." + toolName
    for _, tool := range tools.Tools {
        if tool.Name == fullToolName {
            toolFound = true
            break
        }
    }
    require.True(t, toolFound, "Tool not found in list")

    // Call it
    // The tool name is prefixed with service name
    _, err = serverInfo.CallTool(context.Background(), &mcp.CallToolParams{
        Name: fullToolName,
        Arguments: map[string]interface{}{},
    })
    require.NoError(t, err)

    // 6. Verify Mock Server received correct URL
    // Expected: /test?q=valid
    // If bug exists: /test?q=invalid%&q=valid
    require.Equal(t, "/test?q=valid", capturedURL)
}
