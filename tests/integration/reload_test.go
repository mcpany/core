package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestConfigReload(t *testing.T) {
	// Create a temporary directory for the test files
	tmpDir, err := os.MkdirTemp("", "reload-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create an initial configuration file
	configPath := filepath.Join(tmpDir, "config.yaml")
	initialConfig := `
upstreamServices:
  - name: "my-http-service"
    httpService:
      address: "https://api.example.com"
      calls:
        - operationId: "get_user"
          description: "Get user by ID"
          method: GET
          endpointPath: "/users/{userId}"
`
	err = os.WriteFile(configPath, []byte(initialConfig), 0644)
	require.NoError(t, err)

	// Start the MCPANY server
	mcpanyTestServerInfo := StartMCPANYServerWithConfig(t, "reload-test", initialConfig)
	defer func() {
		t.Logf("Server logs:\n%s", mcpanyTestServerInfo.Process.StderrString())
		mcpanyTestServerInfo.CleanupFunc()
	}()

	// Check that the initial tool is present
	require.Eventually(t, func() bool {
		tools, err := listTools(mcpanyTestServerInfo.HTTPEndpoint)
		if err != nil {
			return false
		}
		if len(tools) != 1 {
			return false
		}
		return tools[0].Name == "my-http-service.get_user"
	}, 10*time.Second, 500*time.Millisecond, "Initial tool not found")

	// Create a new configuration file
	newConfig := `
upstreamServices:
  - name: "my-http-service"
    httpService:
      address: "https://api.example.com"
      calls:
        - operationId: "get_user"
          description: "Get user by ID"
          method: GET
          endpointPath: "/users/{userId}"
        - operationId: "create_user"
          description: "Create a new user"
          method: POST
          endpointPath: "/users"
`
	err = os.WriteFile(configPath, []byte(newConfig), 0644)
	require.NoError(t, err)

	// Check that the new tool is present
	var tools []*mcp.Tool
	require.Eventually(t, func() bool {
		var err error
		tools, err = listTools(mcpanyTestServerInfo.HTTPEndpoint)
		if err != nil {
			return false
		}
		return len(tools) == 2
	}, 10*time.Second, 500*time.Millisecond, "Reloaded tools not found")

	toolNames := []string{tools[0].Name, tools[1].Name}
	require.ElementsMatch(t, toolNames, []string{"my-http-service.get_user", "my-http-service.create_user"})
}

func listTools(endpoint string) ([]*mcp.Tool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "e2e-test-client"}, nil)

	transport := &mcp.StreamableClientTransport{
		Endpoint: endpoint,
	}

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	tools, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, err
	}
	return tools.Tools, nil
}
