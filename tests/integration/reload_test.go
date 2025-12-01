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

	// Print the content of the configuration file
	t.Logf("Initial config file:\n%s", initialConfig)

	// Create a new application
	application := app.NewApplication()

	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a channel to signal when the server is ready
	ready := make(chan struct{})

	// Start the application in a separate goroutine
	go func() {
		close(ready)
		err := application.Run(ctx, afero.NewOsFs(), false, "0", "0", []string{configPath}, 5*time.Second)
		require.NoError(t, err)
	}()

	// Wait for the server to start
	<-ready
	time.Sleep(2 * time.Second)

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
	tools = application.ToolManager.ListTools()
	require.Len(t, tools, 2)
	require.Equal(t, "my-http-service.get_user", tools[0].Tool().GetName())
	require.Equal(t, "my-http-service.create_user", tools[1].Tool().GetName())
}
