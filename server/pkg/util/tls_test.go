package util //nolint:revive

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create temporary cert and key files for testing
func generateTestCerts(t *testing.T, tempDir string) (certPath, keyPath string) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Co"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	require.NoError(t, err)

	certFile, err := os.Create(filepath.Join(tempDir, "cert.pem")) //nolint:gosec // Test file
	require.NoError(t, err)
	_ = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	_ = certFile.Close()

	keyFile, err := os.Create(filepath.Join(tempDir, "key.pem")) //nolint:gosec // Test file
	require.NoError(t, err)
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(t, err)
	_ = pem.Encode(keyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	_ = keyFile.Close()

	return certFile.Name(), keyFile.Name()
}

func TestNewHTTPClientWithTLS(t *testing.T) {
	t.Parallel()
	t.Run("nil config returns client with safe dialer", func(t *testing.T) {
		client, err := NewHTTPClientWithTLS(nil)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotEqual(t, http.DefaultClient, client)

		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)
		assert.NotNil(t, transport.DialContext)

		// Check if DialContext blocks LinkLocal (169.254.169.254)
		// Use a short timeout to fail fast if connection hangs
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		_, err = transport.DialContext(ctx, "tcp", "169.254.169.254:80")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
	})

	t.Run("empty config returns a valid client", func(t *testing.T) {
		client, err := NewHTTPClientWithTLS(&configv1.TLSConfig{})
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotEqual(t, http.DefaultClient, client)
		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)
		assert.NotNil(t, transport.TLSClientConfig)
		assert.False(t, transport.TLSClientConfig.InsecureSkipVerify)
		// Check SafeDialer presence
		assert.NotNil(t, transport.DialContext)
	})

	t.Run("insecure skip verify is set correctly", func(t *testing.T) {
		config := &configv1.TLSConfig{}
		config.SetInsecureSkipVerify(true)
		client, err := NewHTTPClientWithTLS(config)
		require.NoError(t, err)
		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)
		assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
	})

	t.Run("with CA cert", func(t *testing.T) {
		tempDir := t.TempDir()
		caCertPath, _ := generateTestCerts(t, tempDir)
		config := &configv1.TLSConfig{}
		config.SetCaCertPath(caCertPath)
		client, err := NewHTTPClientWithTLS(config)
		require.NoError(t, err)
		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)
		assert.NotNil(t, transport.TLSClientConfig.RootCAs)
	})

	t.Run("with invalid CA cert path", func(t *testing.T) {
		config := &configv1.TLSConfig{}
		config.SetCaCertPath("/path/to/nonexistent/cert.pem")
		_, err := NewHTTPClientWithTLS(config)
		require.Error(t, err)
	})

	t.Run("with malformed CA cert", func(t *testing.T) {
		tempDir := t.TempDir()
		malformedCertPath := filepath.Join(tempDir, "malformed.pem")
		err := os.WriteFile(malformedCertPath, []byte("not a cert"), 0o600)
		require.NoError(t, err)

		config := &configv1.TLSConfig{}
		config.SetCaCertPath(malformedCertPath)
		_, err = NewHTTPClientWithTLS(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to append CA certs from PEM")
	})

	t.Run("with client certs", func(t *testing.T) {
		tempDir := t.TempDir()
		clientCertPath, clientKeyPath := generateTestCerts(t, tempDir)
		config := &configv1.TLSConfig{}
		config.SetClientCertPath(clientCertPath)
		config.SetClientKeyPath(clientKeyPath)
		client, err := NewHTTPClientWithTLS(config)
		require.NoError(t, err)
		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)
		assert.Len(t, transport.TLSClientConfig.Certificates, 1)
	})

	t.Run("with invalid client cert path", func(t *testing.T) {
		tempDir := t.TempDir()
		_, clientKeyPath := generateTestCerts(t, tempDir)
		config := &configv1.TLSConfig{}
		config.SetClientCertPath("/path/to/nonexistent/cert.pem")
		config.SetClientKeyPath(clientKeyPath)
		_, err := NewHTTPClientWithTLS(config)
		require.Error(t, err)
	})

	t.Run("with all settings configured", func(t *testing.T) {
		tempDir := t.TempDir()
		caCertPath, _ := generateTestCerts(t, tempDir)
		clientCertPath, clientKeyPath := generateTestCerts(t, tempDir)

		config := &configv1.TLSConfig{}
		config.SetServerName("test.example.com")
		config.SetInsecureSkipVerify(true)
		config.SetCaCertPath(caCertPath)
		config.SetClientCertPath(clientCertPath)
		config.SetClientKeyPath(clientKeyPath)

		client, err := NewHTTPClientWithTLS(config)
		require.NoError(t, err)
		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)

		assert.Equal(t, "test.example.com", transport.TLSClientConfig.ServerName)
		assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
		assert.NotNil(t, transport.TLSClientConfig.RootCAs)
		assert.Len(t, transport.TLSClientConfig.Certificates, 1)
	})
}

func TestNewHTTPClientWithTLS_Env(t *testing.T) {
	t.Run("with MCPANY_DENY_PRIVATE_UPSTREAM", func(t *testing.T) {
		t.Setenv("MCPANY_DENY_PRIVATE_UPSTREAM", "true")
		client, err := NewHTTPClientWithTLS(nil)
		require.NoError(t, err)
		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)

		// Should block private
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		_, err = transport.DialContext(ctx, "tcp", "192.168.1.1:80")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
	})
}
