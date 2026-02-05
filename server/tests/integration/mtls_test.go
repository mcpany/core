// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestMTLSAuthentication(t *testing.T) {
	// Create a mock upstream server that requires mTLS
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode("ok")
	}))

	// Configure the server with mTLS
	caCert, err := os.ReadFile("../tls/ca.crt")
	require.NoError(t, err)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	serverCert, err := tls.LoadX509KeyPair("../tls/server.crt", "../tls/server.key")
	require.NoError(t, err)

	server.TLS = &tls.Config{ //nolint:gosec
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{serverCert},
	}
	server.StartTLS()
	defer server.Close()

	// Configure the gateway to use mTLS for the upstream.
	// The paths must be relative to the project root, where the server binary runs.
	config := `
upstream_services:
  - name: my-upstream
    upstream_auth:
      mtls:
        client_cert_path: "tests/tls/client.crt"
        client_key_path: "tests/tls/client.key"
        ca_cert_path: "tests/tls/ca.crt"
    http_service:
      address: "` + server.URL + `"
      tools:
        - name: "my-tool"
          call_id: "get-root"
      calls:
        get-root:
          endpoint_path: "/"
          method: "HTTP_METHOD_GET"
`
	// Start the gateway
	serverInfo := StartMCPANYServerWithConfig(t, "mtls-test", config)
	defer serverInfo.CleanupFunc()

	// Create a new MCP client.
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)

	// Create a new streamable HTTP transport.
	transport := &mcp.StreamableClientTransport{Endpoint: serverInfo.HTTPEndpoint}

	// Connect to the server.
	session, err := client.Connect(context.Background(), transport, nil)
	require.NoError(t, err, "failed to connect to mcp server")
	defer func() { _ = session.Close() }()

	expectedToolName := "my-upstream.my-tool"

	// List the tools and check for the expected tool.
	listResult, err := session.ListTools(context.Background(), &mcp.ListToolsParams{})
	require.NoError(t, err, "failed to list tools")

	found := false
	for _, tool := range listResult.Tools {
		if tool.Name == expectedToolName {
			found = true
			break
		}
	}
	require.True(t, found, "Tool '%s' not found in the list of available tools", expectedToolName)

	// Call the tool, which should use mTLS
	params := &mcp.CallToolParams{
		Name: expectedToolName,
	}
	callResult, err := session.CallTool(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, callResult, "result should not be nil")
	require.Len(t, callResult.Content, 1)
	textContent, ok := callResult.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected text content")

	var result string
	err = json.Unmarshal([]byte(textContent.Text), &result)
	require.NoError(t, err)
	require.Equal(t, "ok", result)
}
