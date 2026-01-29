//go:build e2e

package all_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestLoadAllPopularServices(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for All Popular Services...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EAllPopularServicesTest", "--config-path", "../../../examples/popular_services")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)

	// --- 3. Assert Response ---
	// Find all the config files
	configs, err := filepath.Glob("../../../examples/popular_services/*/config.yaml")
	require.NoError(t, err)
	require.Greater(t, len(configs), 0, "No popular service configs found")

	// Create a new FileStore
	fs := afero.NewOsFs()
	store := config.NewFileStore(fs, configs)

	// Load the config
	cfg, err := config.LoadServices(context.Background(), store, "server")
	require.NoError(t, err)

	// Get the expected tool names
	var expectedToolNames []string
	for _, service := range cfg.GetUpstreamServices() {
		sanitizedServiceName, _ := util.SanitizeServiceName(service.GetName())
		for _, tool := range service.GetHttpService().GetTools() {
			sanitizedToolName, _ := util.SanitizeToolName(tool.GetName())
			expectedToolNames = append(expectedToolNames, sanitizedServiceName+"."+sanitizedToolName)
		}
	}

	// Get the actual tool names
	var actualToolNames []string
	for _, tool := range listToolsResult.Tools {
		actualToolNames = append(actualToolNames, tool.Name)
	}

	// Assert that the tool names match
	require.ElementsMatch(t, expectedToolNames, actualToolNames, "The discovered tools do not match the expected tools")

	for _, tool := range listToolsResult.Tools {
		t.Logf("Discovered tool from MCPANY: %s", tool.Name)
	}

	t.Log("INFO: E2E Test Scenario for All Popular Services Completed Successfully!")
}
