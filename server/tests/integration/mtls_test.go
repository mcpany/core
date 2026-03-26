// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// generateMTLSCerts creates a self-signed CA and signs server + client
// certificates, writing all PEM files into dir.
func generateMTLSCerts(t *testing.T, dir string) {
	t.Helper()

	writePEM := func(path, pemType string, data []byte) {
		t.Helper()
		f, err := os.Create(path) //nolint:gosec
		require.NoError(t, err)
		defer f.Close()
		require.NoError(t, pem.Encode(f, &pem.Block{Type: pemType, Bytes: data}))
	}

	// --- CA ---
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	require.NoError(t, err)
	caCert, err := x509.ParseCertificate(caBytes)
	require.NoError(t, err)
	writePEM(filepath.Join(dir, "ca.crt"), "CERTIFICATE", caBytes)

	// --- Server cert ---
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "localhost"},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	serverBytes, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	require.NoError(t, err)
	writePEM(filepath.Join(dir, "server.crt"), "CERTIFICATE", serverBytes)
	writePEM(filepath.Join(dir, "server.key"), "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(serverKey))

	// --- Client cert ---
	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject:      pkix.Name{CommonName: "test-client"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientBytes, err := x509.CreateCertificate(rand.Reader, clientTemplate, caCert, &clientKey.PublicKey, caKey)
	require.NoError(t, err)
	writePEM(filepath.Join(dir, "client.crt"), "CERTIFICATE", clientBytes)
	writePEM(filepath.Join(dir, "client.key"), "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(clientKey))
}

func TestMTLSAuthentication(t *testing.T) {
	// Generate TLS certificates programmatically so the test is hermetic and
	// works in all environments (including Bazel sandboxes where cert symlinks
	// resolve outside the server working-directory security check).
	certDir := t.TempDir()
	generateMTLSCerts(t, certDir)

	// Create a mock upstream server that requires mTLS
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode("ok")
	}))

	// Configure the server with mTLS
	caCert, err := os.ReadFile(filepath.Join(certDir, "ca.crt")) //nolint:gosec
	require.NoError(t, err)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	serverCert, err := tls.LoadX509KeyPair(filepath.Join(certDir, "server.crt"), filepath.Join(certDir, "server.key")) //nolint:gosec
	require.NoError(t, err)

	server.TLS = &tls.Config{ //nolint:gosec
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{serverCert},
	}
	server.StartTLS()
	defer server.Close()

	// Configure the gateway to use mTLS for the upstream.
	// Cert paths are absolute; certDir is added to allowed_file_paths so the
	// server's path-validation logic accepts them regardless of CWD.
	config := `
global_settings:
  allowed_file_paths:
    - ` + certDir + `
upstream_services:
  - name: my-upstream
    upstream_auth:
      mtls:
        client_cert_path: "` + filepath.Join(certDir, "client.crt") + `"
        client_key_path: "` + filepath.Join(certDir, "client.key") + `"
        ca_cert_path: "` + filepath.Join(certDir, "ca.crt") + `"
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
