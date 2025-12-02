package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/app"
	"github.com/spf13/afero"
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

	// Create a new application
	application := app.NewApplication()

	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the application in a separate goroutine
	go func() {
		// This Run call will block until the context is canceled.
		_ = application.Run(ctx, afero.NewOsFs(), false, "0", "0", []string{configPath}, 5*time.Second)
	}()

	// Wait for the server to start and load the initial configuration
	require.Eventually(t, func() bool {
		tools := application.ToolManager.ListTools()
		return len(tools) == 1
	}, 5*time.Second, 100*time.Millisecond, "server did not load initial tool in time")

	// Check that the initial tool is present
	tools := application.ToolManager.ListTools()
	require.Len(t, tools, 1)
	require.Equal(t, "my-http-service.get_user", tools[0].Tool().GetName())

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

	// Reload the configuration
	err = application.ReloadConfig(afero.NewOsFs(), []string{configPath})
	require.NoError(t, err)

	// Check that the new tool is present
	require.Eventually(t, func() bool {
		tools = application.ToolManager.ListTools()
		return len(tools) == 2
	}, 2*time.Second, 100*time.Millisecond, "server did not reload to two tools")

	require.Len(t, tools, 2)

	// Note: The order of tools is not guaranteed, so we check for presence instead of order.
	toolNames := []string{tools[0].Tool().GetName(), tools[1].Tool().GetName()}
	require.Contains(t, toolNames, "my-http-service.get_user")
	require.Contains(t, toolNames, "my-http-service.create_user")
}
