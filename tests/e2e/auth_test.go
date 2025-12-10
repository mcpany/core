package e2e

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/app"
	"github.com/mcpany/core/pkg/auth"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2EAuth(t *testing.T) {
	const apiKey = "test-api-key-12345"

	// Create a temporary config file
	configFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(configFile.Name())

	configContent := fmt.Sprintf(`
global_settings:
  api_key: "%s"
`, apiKey)
	_, err = configFile.WriteString(configContent)
	require.NoError(t, err)
	configFile.Close()

	// Get free ports for JSON-RPC and gRPC
	jsonrpcPort, err := getFreePort()
	require.NoError(t, err)
	grpcPort, err := getFreePort()
	require.NoError(t, err)

	// Run the server in a separate goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fs := afero.NewOsFs()
	application := app.NewApplication()

	go func() {
		err := application.Run(
			ctx,
			fs,
			false,
			fmt.Sprintf(":%d", jsonrpcPort),
			fmt.Sprintf(":%d", grpcPort),
			[]string{configFile.Name()},
			5*time.Second,
		)
		if err != nil && err != context.Canceled {
			t.Logf("Server returned an error: %v", err)
		}
	}()

	// Wait for the server to be ready
	require.Eventually(t, func() bool {
		err := app.HealthCheckWithContext(ctx, io.Discard, fmt.Sprintf("localhost:%d", jsonrpcPort))
		return err == nil
	}, 5*time.Second, 100*time.Millisecond, "Server did not start in time")

	t.Run("Valid API Key", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d", jsonrpcPort), strings.NewReader(`{"jsonrpc": "2.0", "method": "tools/list", "id": 1}`))
		require.NoError(t, err)
		req.Header.Set(auth.APIKeyHeader, apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Missing API Key", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d", jsonrpcPort), strings.NewReader(`{"jsonrpc": "2.0", "method": "tools/list", "id": 1}`))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Invalid API Key", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d", jsonrpcPort), strings.NewReader(`{"jsonrpc": "2.0", "method": "tools/list", "id": 1}`))
		require.NoError(t, err)
		req.Header.Set(auth.APIKeyHeader, "invalid-key")
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Health Check without API Key", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/healthz", jsonrpcPort), nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
