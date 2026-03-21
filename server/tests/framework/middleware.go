// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"context"

	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

// TestE2ECaching tests the end-to-end caching functionality.
//
// t is the t.
func TestE2ECaching(t *testing.T) {
	t.Parallel()
	RunE2ETest(t, &E2ETestCase{
		Name:                "caching",
		UpstreamServiceType: "http",
		BuildUpstream:       BuildCachingServer,
		RegisterUpstream:    RegisterCachingService,
		ValidateMiddlewares: func(t *testing.T, mcpanyEndpoint, upstreamEndpoint string) {
			ValidateCaching(t, mcpanyEndpoint, upstreamEndpoint)
		},
		InvokeAIClient:      func(_ *testing.T, _ string) {},
		RegistrationMethods: []RegistrationMethod{GRPCRegistration},
	})
}

// BuildCachingServer builds and starts a caching server for testing.
//
// t is the t.
//
// Returns the result.
func BuildCachingServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	proc := integration.NewManagedProcess(t, "http_caching_server", integration.MockBinary(t, "http_caching_server"), []string{"--port", fmt.Sprintf("%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterCachingService registers the caching service with the MCP server.
//
// t is the t.
// registrationClient is the registrationClient.
// upstreamEndpoint is the upstreamEndpoint.
func RegisterCachingService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	serviceID := "e2e_caching_server"
	operationID := "get_data"
	callID := "call-" + operationID
	method := configv1.HttpCallDefinition_HTTP_METHOD_GET
	cacheEnabled := true
	ttl := durationpb.New(5 * time.Second)

	req := apiv1.RegisterServiceRequest_builder{
		Config: configv1.UpstreamServiceConfig_builder{
			Name: &serviceID,
			HttpService: configv1.HttpUpstreamService_builder{
				Address: &upstreamEndpoint,
				Tools: []*configv1.ToolDefinition{
					configv1.ToolDefinition_builder{
						Name:   &operationID,
						CallId: &callID,
					}.Build(),
				},
				Calls: map[string]*configv1.HttpCallDefinition{
					callID: configv1.HttpCallDefinition_builder{
						Id:           &callID,
						EndpointPath: protoString("/"),
						Method:       &method,
						Cache: configv1.CacheConfig_builder{
							IsEnabled: &cacheEnabled,
							Ttl:       ttl,
						}.Build(),
					}.Build(),
				},
			}.Build(),
		}.Build(),
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationClient, req)
}

func protoString(value string) *string {
	return &value
}

// NoOpMiddleware is a middleware that does nothing and calls the next handler.
//
// _ is an unused parameter.
// next is the next.
//
// Returns the result.
func NoOpMiddleware(_ *testing.T, next http.Handler) http.Handler {
	return next
}

func connectMCP(t *testing.T, mcpanyEndpoint string) *mcp.ClientSession {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "framework-test-client"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = session.Close()
	})

	return session
}

func callTool(t *testing.T, session *mcp.ClientSession, toolName string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	_, err := session.CallTool(ctx, &mcp.CallToolParams{Name: toolName})
	require.NoError(t, err)
}

// ValidateCaching validates that caching is working correctly.
//
// t is the t.
// mcpanyEndpoint is the mcpanyEndpoint.
// upstreamEndpoint is the upstreamEndpoint.
func ValidateCaching(t *testing.T, mcpanyEndpoint, upstreamEndpoint string) {
	session := connectMCP(t, mcpanyEndpoint)

	baseURL := upstreamEndpoint
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}

	// 1. Reset the upstream server's counter.
	req, err := http.NewRequestWithContext(context.Background(), "POST", baseURL+"/reset", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// 2. Make a request to the tool and check that the upstream service was called.
	callTool(t, session, "e2e_caching_server.get_data")

	metrics := getUpstreamMetrics(t, baseURL)
	require.Equal(t, int64(1), metrics["counter"])

	// 3. Make another request to the tool and check that the upstream service was NOT called.
	callTool(t, session, "e2e_caching_server.get_data")

	metrics = getUpstreamMetrics(t, baseURL)
	require.Equal(t, int64(1), metrics["counter"])

	// 4. Advance the fake clock to expire the cache.
	time.Sleep(6 * time.Second)

	// 5. Make another request to the tool and check that the upstream service was called.
	callTool(t, session, "e2e_caching_server.get_data")

	metrics = getUpstreamMetrics(t, baseURL)
	require.Equal(t, int64(2), metrics["counter"])
}

func getUpstreamMetrics(t *testing.T, upstreamEndpoint string) map[string]int64 {
	req, err := http.NewRequestWithContext(context.Background(), "GET", upstreamEndpoint+"/metrics", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var metrics map[string]int64
	err = json.NewDecoder(resp.Body).Decode(&metrics)
	require.NoError(t, err)

	return metrics
}
