// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/tests/integration"
	"github.com/stretchr/testify/require"
)

func TestE2ECaching(t *testing.T) {
	RunE2ETest(t, &E2ETestCase{
		Name:                "caching",
		UpstreamServiceType: "http",
		BuildUpstream:       BuildCachingServer,
		RegisterUpstream:    RegisterCachingService,
		ValidateMiddlewares: func(t *testing.T, mcpanyEndpoint, upstreamEndpoint string) {
			ValidateCaching(t, mcpanyEndpoint, upstreamEndpoint)
		},
		InvokeAIClient:      func(t *testing.T, mcpanyEndpoint string) {},
		RegistrationMethods: []RegistrationMethod{GRPCRegistration},
	})
}

func BuildCachingServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "http_caching_server", filepath.Join(root, "build/test/bin/http_caching_server"), []string{"--port", fmt.Sprintf("%d", port)}, nil)
	proc.Port = port
	return proc
}

func RegisterCachingService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_caching_server"
	integration.RegisterHTTPService(t, registrationClient, serviceID, upstreamEndpoint, "get_data", "/", "GET", nil)
}

func callTool(t *testing.T, mcpanyEndpoint, toolName string) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": toolName,
		},
		"id": "1",
	})
	require.NoError(t, err)

	resp, err := http.Post(mcpanyEndpoint, "application/json", bytes.NewBuffer(requestBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func ValidateCaching(t *testing.T, mcpanyEndpoint, upstreamEndpoint string) {
	// 1. Reset the upstream server's counter.
	resp, err := http.Post(fmt.Sprintf("http://%s/reset", upstreamEndpoint), "", nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// 2. Make a request to the tool and check that the upstream service was called.
	callTool(t, mcpanyEndpoint, "e2e_caching_server.get_data")

	metrics := getUpstreamMetrics(t, upstreamEndpoint)
	require.Equal(t, int64(1), metrics["counter"])

	// 3. Make another request to the tool and check that the upstream service was NOT called.
	callTool(t, mcpanyEndpoint, "e2e_caching_server.get_data")

	metrics = getUpstreamMetrics(t, upstreamEndpoint)
	require.Equal(t, int64(1), metrics["counter"])

	// 4. Advance the fake clock to expire the cache.
	time.Sleep(6 * time.Second)

	// 5. Make another request to the tool and check that the upstream service was called.
	callTool(t, mcpanyEndpoint, "e2e_caching_server.get_data")

	metrics = getUpstreamMetrics(t, upstreamEndpoint)
	require.Equal(t, int64(2), metrics["counter"])
}

func getUpstreamMetrics(t *testing.T, upstreamEndpoint string) map[string]int64 {
	resp, err := http.Get(fmt.Sprintf("http://%s/metrics", upstreamEndpoint))
	require.NoError(t, err)
	defer resp.Body.Close()

	var metrics map[string]int64
	err = json.NewDecoder(resp.Body).Decode(&metrics)
	require.NoError(t, err)

	return metrics
}
