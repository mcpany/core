// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestMTLSAuthentication(t *testing.T) {
	// Resolve the TLS test certificate directory.
	tlsDir := filepath.Join(ProjectRoot(t), "tests", "tls")
	if _, err := os.Stat(tlsDir); err != nil {
		t.Skipf("TLS test certificates not found at %s, skipping", tlsDir)
	}

	// Under Bazel the sandbox uses symlinks; the server's path validation rejects those.
	// Copy certs to a real temp directory so absolute paths pass EvalSymlinks checks.
	realTLSDir := t.TempDir()
	for _, f := range []string{"ca.crt", "server.crt", "server.key", "client.crt", "client.key"} {
		src := filepath.Join(tlsDir, f)
		dst := filepath.Join(realTLSDir, f)
		data, err := os.ReadFile(src) //nolint:gosec
		if err != nil {
			t.Skipf("Cannot read TLS cert file %s: %v", src, err)
		}
		if err := os.WriteFile(dst, data, 0600); err != nil { //nolint:gosec
			t.Fatalf("Cannot write TLS cert file %s: %v", dst, err)
		}
	}

	// Create a mock upstream server that requires mTLS
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode("ok")
	}))

	// Configure the server with mTLS
	caCert, err := os.ReadFile(filepath.Join(realTLSDir, "ca.crt")) //nolint:gosec
	require.NoError(t, err)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	serverCert, err := tls.LoadX509KeyPair(filepath.Join(realTLSDir, "server.crt"), filepath.Join(realTLSDir, "server.key")) //nolint:gosec
	require.NoError(t, err)

	server.TLS = &tls.Config{ //nolint:gosec
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{serverCert},
	}
	server.StartTLS()
	defer server.Close()

	// Configure the gateway to use mTLS for the upstream.
	// Use absolute paths to the temp cert directory so the server's path validation
	// (which resolves symlinks) accepts them in any environment including Bazel.
	config := fmt.Sprintf(`
upstream_services:
  - name: my-upstream
    upstream_auth:
      mtls:
        client_cert_path: %s/client.crt
        client_key_path: %s/client.key
        ca_cert_path: %s/ca.crt
    http_service:
      address: "%s"
      tools:
        - name: "my-tool"
          call_id: "get-root"
      calls:
        get-root:
          endpoint_path: "/"
          method: "HTTP_METHOD_GET"
`, realTLSDir, realTLSDir, realTLSDir, server.URL)
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
