// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package e2e

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateCert(t *testing.T, isCA bool, commonName string, parent *x509.Certificate, parentKey *rsa.PrivateKey) (*x509.Certificate, *rsa.PrivateKey, []byte, []byte) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	notBefore := time.Now()
	notAfter := notBefore.Add(1 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
			CommonName:   commonName,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	} else {
		template.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
	}

	parentCert := &template
	if parent != nil {
		parentCert = parent
	}
	parentK := priv
	if parentKey != nil {
		parentK = parentKey
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, parentCert, &priv.PublicKey, parentK)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	cert, err := x509.ParseCertificate(derBytes)
	require.NoError(t, err)

	return cert, priv, certPEM, keyPEM
}

func TestStartup_mTLS(t *testing.T) {
	tmpDir := t.TempDir()

	// Generate CA
	caCert, caKey, caCertPEM, _ := generateCert(t, true, "Test CA", nil, nil)
	caPath := filepath.Join(tmpDir, "ca.crt")
	require.NoError(t, os.WriteFile(caPath, caCertPEM, 0644))

	// Generate Server Cert (signed by CA)
	_, _, serverCertPEM, serverKeyPEM := generateCert(t, false, "server", caCert, caKey)
	serverCertPath := filepath.Join(tmpDir, "server.crt")
	serverKeyPath := filepath.Join(tmpDir, "server.key")
	require.NoError(t, os.WriteFile(serverCertPath, serverCertPEM, 0644))
	require.NoError(t, os.WriteFile(serverKeyPath, serverKeyPEM, 0600))

	// Generate Client Cert (signed by CA)
	_, _, clientCertPEM, clientKeyPEM := generateCert(t, false, "client", caCert, caKey)
	clientCertPath := filepath.Join(tmpDir, "client.crt")
	clientKeyPath := filepath.Join(tmpDir, "client.key")
	require.NoError(t, os.WriteFile(clientCertPath, clientCertPEM, 0644))
	require.NoError(t, os.WriteFile(clientKeyPath, clientKeyPEM, 0600))

	// Config file
	configPath := filepath.Join(tmpDir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("global_settings:\n  log_level: DEBUG"), 0644))

	// Get a free port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	l.Close()
	_, _, err = net.SplitHostPort(addr)
	require.NoError(t, err)

	// Build server binary
	binPath := filepath.Join(tmpDir, "server")
	buildCmd := exec.Command("go", "build", "-o", binPath, "../../cmd/server/main.go")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	require.NoError(t, buildCmd.Run(), "Failed to build server binary")

	// Run server
	cmd := exec.Command(binPath, "run",
		"--config-path", configPath,
		"--mcp-listen-address", addr,
		"--grpc-port", "127.0.0.1:0", // using random port for grpc to avoid collision
		"--tls-cert", serverCertPath,
		"--tls-key", serverKeyPath,
		"--tls-client-ca", caPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "MCPANY_ENABLE_FILE_CONFIG=true")

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
	}()

	// Wait for server to be ready
	baseURL := "https://" + addr
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: func() *x509.CertPool {
					pool := x509.NewCertPool()
					pool.AppendCertsFromPEM(caCertPEM)
					return pool
				}(),
				Certificates: []tls.Certificate{
					func() tls.Certificate {
						cert, err := tls.X509KeyPair(clientCertPEM, clientKeyPEM)
						require.NoError(t, err)
						return cert
					}(),
				},
			},
		},
		Timeout: 5 * time.Second,
	}

	require.Eventually(t, func() bool {
		req, _ := http.NewRequest("GET", baseURL+"/health", nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Client Do error: %v", err)
			return false
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Status: %d, Body: %s", resp.StatusCode, body)
			return false
		}
		return true
	}, 10*time.Second, 500*time.Millisecond, "Server failed to start or accept mTLS connection")

	t.Run("Client with valid cert should succeed", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL+"/health", nil)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Client without cert should fail", func(t *testing.T) {
		noCertClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // We want to verify that server REJECTS us, not that we reject server
				},
			},
		}
		_, err := noCertClient.Get(baseURL + "/health")
		// It might be a connection error (alert) or handshake failure
		assert.Error(t, err)
	})
}
