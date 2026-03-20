// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"testing"
	"time"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// SeedStandardData populates the server with a standard set of data for E2E testing.
// It registers a "Core" service (mocked) and a "Tools" service (mocked).
// It also seeds some traffic history if supported.
func SeedStandardData(t *testing.T, serverInfo *MCPANYTestServerInfo) {
	t.Helper()

	// 1. Start a mock server for our seeded services
	// We need this mock server to stay alive as long as the test server is running.
	// But `StartMockServer` returns a server that needs to be closed.
	// We can attach the cleanup to the test.
	mockHandler := DefaultMockHandler(t, map[string]string{
		"/core/status": `{"status": "ok", "version": "1.0.0"}`,
		"/tools/list":  `{"tools": [{"name": "calculator", "description": "Basic calculator"}]}`,
	})
	mockServer := StartMockServer(t, mockHandler)
	t.Cleanup(mockServer.Close)

	// 2. Register "Core" Service
	coreConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("seed-core"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(mockServer.URL),
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name: proto.String("status"),
					CallId: proto.String("status"),
				}.Build(),
			},
			Calls: map[string]*configv1.HttpCallDefinition{
				"status": configv1.HttpCallDefinition_builder{
					Id:           proto.String("status"),
					EndpointPath: proto.String("/core/status"),
					Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
				}.Build(),
			},
		}.Build(),
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: coreConfig,
	}.Build()

	RegisterServiceViaAPI(t, serverInfo.RegistrationClient, req)
	t.Log("Seeded 'seed-core' service.")

	// 3. Register "Tools" Service
	toolsConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("seed-tools"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(mockServer.URL),
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name: proto.String("calculator"),
					CallId: proto.String("calculator"),
				}.Build(),
			},
			Calls: map[string]*configv1.HttpCallDefinition{
				"calculator": configv1.HttpCallDefinition_builder{
					Id:           proto.String("calculator"),
					EndpointPath: proto.String("/tools/list"), // Just dummy endpoint
					Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
				}.Build(),
			},
		}.Build(),
	}.Build()

	reqTools := apiv1.RegisterServiceRequest_builder{
		Config: toolsConfig,
	}.Build()

	RegisterServiceViaAPI(t, serverInfo.RegistrationClient, reqTools)
	t.Log("Seeded 'seed-tools' service.")

	// 4. Seed Traffic/History (Optional, via Debug Endpoint if needed)
	// Example traffic data seeding
	trafficData := []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-1 * time.Hour).UnixMilli(),
			"serviceId": "seed-core",
			"requests":  100,
			"errors":    2,
			"latency":   50,
		},
		{
			"timestamp": time.Now().UnixMilli(),
			"serviceId": "seed-core",
			"requests":  120,
			"errors":    0,
			"latency":   45,
		},
	}

	// Assuming /api/v1/debug/seed_traffic exists and accepts JSON list
	// We need to implement SeedTraffic in e2e_helpers or here.
	// serverInfo.HTTPClient...
	// For now, I'll skip traffic seeding until I verify the endpoint exists.
	// But I can add a placeholder.
	_ = trafficData

	t.Log("Standard data seeding complete.")
}
