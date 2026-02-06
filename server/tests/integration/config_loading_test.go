package integration

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	v1 "github.com/mcpany/core/proto/api/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestConfigLoading(t *testing.T) {
	testCases := []struct {
		name               string
		configFile         string
		expectedToolName   string
		toolShouldBeLoaded bool
	}{
		{
			name:               "json config",
			configFile:         "testdata/config.json",
			expectedToolName:   "http-echo-from-json",
			toolShouldBeLoaded: true,
		},
		{
			name:               "yaml config",
			configFile:         "testdata/config.yaml",
			expectedToolName:   "http-echo-from-yaml",
			toolShouldBeLoaded: true,
		},
		{
			name:               "textproto config",
			configFile:         "testdata/config.textproto",
			expectedToolName:   "http-echo-from-textproto",
			toolShouldBeLoaded: true,
		},
		{
			name:               "disabled config",
			configFile:         "testdata/disabled_config.yaml",
			expectedToolName:   "disabled-service",
			toolShouldBeLoaded: false,
		},
	}

	root, err := GetProjectRoot()
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("MCPANY_BINARY_PATH", filepath.Join(root, "../build/bin/server"))
			absConfigFile := filepath.Join(root, "tests", "integration", tc.configFile)


			mcpAny := StartMCPANYServer(t, "config-loading-"+tc.name, "--config-path", absConfigFile)
			defer mcpAny.CleanupFunc()

			conn, err := grpc.NewClient(mcpAny.GrpcRegistrationEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
			require.NoError(t, err)
			defer func() { _ = conn.Close() }()

			client := v1.NewRegistrationServiceClient(conn)

			require.Eventually(t, func() bool {
				resp, err := client.ListServices(context.Background(), v1.ListServicesRequest_builder{}.Build())
				require.NoError(t, err)

				var serviceFound bool
				for _, service := range resp.GetServices() {
					if service.GetName() == tc.expectedToolName {
						serviceFound = true
						break
					}
				}
				return serviceFound == tc.toolShouldBeLoaded
			}, 10*time.Second, 500*time.Millisecond, "service loading status mismatch")
		})
	}
}

func TestDisabledHierarchyConfig(t *testing.T) {
	root, err := GetProjectRoot()
	require.NoError(t, err)

	t.Setenv("MCPANY_BINARY_PATH", filepath.Join(root, "../build/bin/server"))
	absConfigFile := filepath.Join(root, "tests", "integration", "testdata", "disabled_hierarchy_config.yaml")

	mcpAny := StartMCPANYServer(t, "config-loading-hierarchy", "--config-path", absConfigFile)
	defer mcpAny.CleanupFunc()

	conn, err := grpc.NewClient(mcpAny.GrpcRegistrationEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	client := v1.NewRegistrationServiceClient(conn)

	// Verify Services
	require.Eventually(t, func() bool {
		resp, err := client.ListServices(context.Background(), v1.ListServicesRequest_builder{}.Build())
		require.NoError(t, err)

		services := make(map[string]bool)
		for _, service := range resp.GetServices() {
			services[service.GetName()] = true
		}

		// 1. Service Disabled
		if services["disabled-service"] {
			t.Log("disabled-service found but should be disabled")
			return false
		}
		// 2. Tool Disabled - Service should be present
		if !services["tool-disabled-service"] {
			t.Log("tool-disabled-service not found")
			return false
		}
		// 3. Call Disabled - Service should be present
		if !services["call-disabled-service"] {
			t.Log("call-disabled-service not found")
			return false
		}
		return true
	}, 10*time.Second, 500*time.Millisecond, "Service list mismatch")

	// Verify via Logs
	// We wait a bit for logs to flush (StartMCPANYServer waits for health, so logs should be mostly there,
	// but registration is async).
	// We already waited for "Service list mismatch" which confirms services are loaded,
	// but registration worker might still be processing.
	// Actually ListServices showed they are there.

	// Helper to check logs
	require.Eventually(t, func() bool {
		logs := mcpAny.Process.StdoutString()

		// 1. Service Disabled
		if !strings.Contains(logs, `Service disabled by profile override, skipping`) || !strings.Contains(logs, `service_name=disabled-service`) {
			return false
		}

		// 2. Call Disabled
		if !strings.Contains(logs, `Skipping blocked tool/call" toolName=disabled-call callID=disabled-call`) {
			return false
		}
		// call-disabled-service should have 0 tools
		if !strings.Contains(logs, `Registered HTTP service" serviceID=call-disabled-service toolsAdded=0`) {
			return false
		}

		// 3. Tool Disabled
		// tool-disabled-service should have 1 tool (enabled-tool)
		if !strings.Contains(logs, `Registered HTTP service" serviceID=tool-disabled-service toolsAdded=1`) {
			return false
		}
		// disabled-tool should be skipped
		if !strings.Contains(logs, `Skipping blocked tool/call" toolName=disabled-tool`) {
			return false
		}

		return true
	}, 5*time.Second, 100*time.Millisecond, "Logs did not contain expected disable confirmations.\nLogs:\n%s", mcpAny.Process.StdoutString())
}
