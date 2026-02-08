package mcp

import (
	"context"
	"net/http"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithMCPClientSession(t *testing.T) {
	ctx := context.Background()

	// Mock Connect function
	originalConnect := connectForTesting
	defer func() { connectForTesting = originalConnect }()

	mockSession := new(MockClientSession)
	mockSession.On("Close").Return(nil)

	connectForTesting = func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
		return mockSession, nil
	}

	// We need an mcp.Client instance for the connection
	client := mcp.NewClient(&mcp.Implementation{}, nil)

	t.Run("HTTP", func(t *testing.T) {
		conn := &mcpConnection{
			client:      client,
			httpAddress: "http://localhost:8080",
			httpClient:  http.DefaultClient,
		}

		err := conn.withMCPClientSession(ctx, func(cs ClientSession) error {
			assert.Equal(t, mockSession, cs)
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("Bundle", func(t *testing.T) {
		conn := &mcpConnection{
			client:          client,
			bundleTransport: &BundleDockerTransport{},
		}
		err := conn.withMCPClientSession(ctx, func(cs ClientSession) error {
			assert.Equal(t, mockSession, cs)
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("NoTransport", func(t *testing.T) {
		conn := &mcpConnection{
			client: client,
		}
		err := conn.withMCPClientSession(ctx, func(cs ClientSession) error {
			return nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transport is not configured")
	})

	t.Run("Stdio_Command", func(t *testing.T) {
	stdioConfig := configv1.McpStdioConnection_builder{}.Build()
	stdioConfig.SetCommand("echo")
	stdioConfig.SetArgs([]string{"hello"})

	conn := &mcpConnection{
		client: client,
		stdioConfig: stdioConfig,
	}
		err := conn.withMCPClientSession(ctx, func(cs ClientSession) error {
			assert.Equal(t, mockSession, cs)
			return nil
		})
		require.NoError(t, err)
	})
}
