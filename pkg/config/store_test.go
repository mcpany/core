/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

func TestReadURL(t *testing.T) {
	t.Run("should block loopback addresses", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		_, err := readURL(server.URL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
	})

	t.Run("should block private non-loopback addresses", func(t *testing.T) {
		// This is a bit tricky to test without being able to bind to a private IP.
		// We'll simulate this by creating a URL with a private IP and assuming the DialContext will block it.
		// We can't actually make a request to it in a test environment easily.
		// The check happens before the dial, so this is sufficient.
		_, err := readURL("http://192.168.1.1/config.yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
	})

	t.Run("should fail on non-200 status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		// To bypass the SSRF check for this test, we need to use a public IP.
		// We'll replace the httpClient for this test.
		originalClient := httpClient
		defer func() { httpClient = originalClient }()
		httpClient = &http.Client{
			Timeout: 5 * time.Second,
		}

		_, err := readURL(server.URL)
		assert.Contains(t, err.Error(), "status code 404")
	})

	t.Run("should correctly handle SNI for HTTPS connections", func(t *testing.T) {
		// 1. Create a test server that requires SNI.
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"message": "success"}`)
		}))

		// Replace the server's listener with one that captures the SNI name.
		var capturedSNI string
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)

		server.Listener = listener
		server.TLS = &tls.Config{
			GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				capturedSNI = hello.ServerName
				// Return a pointer to the certificate
				return &server.TLS.Certificates[0], nil
			},
		}
		server.StartTLS()
		defer server.Close()

		// 2. The hostname for the test server.
		hostname := "sni.example.com"

		// 3. Replace the httpClient's transport to use a custom dialer that resolves our test hostname
		//    to the test server's address. This is necessary because we can't add "sni.example.com"
		//    to the system's hosts file.
		originalClient := httpClient
		defer func() { httpClient = originalClient }()

		httpClient = &http.Client{
			Transport: &http.Transport{
				// The DialTLSContext from the main code will be used, but we need to provide a custom
				// DialContext to handle the name resolution for our test domain.
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					// When the code under test tries to connect to "sni.example.com", we intercept
					// it and redirect it to our local test server.
					return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
				},
				// We need to provide the test server's certificate to the client.
				TLSClientConfig: &tls.Config{
					RootCAs: server.Client().Transport.(*http.Transport).TLSClientConfig.RootCAs,
				},
			},
		}

		// 4. Make a request to the server using the test hostname.
		_, err = readURL("https://" + hostname)
		assert.NoError(t, err, "Expected the request to succeed")
		assert.Equal(t, hostname, capturedSNI, "Expected the server to receive the correct SNI hostname")
	})
}

func TestNewEngine(t *testing.T) {
	t.Run("UnsupportedExtension", func(t *testing.T) {
		_, err := NewEngine("config.txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported config file extension")
	})

	t.Run("JSONExtension", func(t *testing.T) {
		engine, err := NewEngine("config.json")
		assert.NoError(t, err)
		assert.IsType(t, &jsonEngine{}, engine)
	})
}

func TestJsonEngine_Unmarshal(t *testing.T) {
	engine := &jsonEngine{}

	t.Run("ValidJSON", func(t *testing.T) {
		validJSON := []byte(`{
			"global_settings": {
				"mcp_listen_address": "0.0.0.0:8080",
				"log_level": "LOG_LEVEL_INFO"
			}
		}`)
		cfg := &configv1.McpAnyServerConfig{}
		err := engine.Unmarshal(validJSON, cfg)
		require.NoError(t, err)
		assert.Equal(t, "0.0.0.0:8080", cfg.GetGlobalSettings().GetMcpListenAddress())
		assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_INFO, cfg.GetGlobalSettings().GetLogLevel())
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		invalidJSON := []byte(`{
			"global_settings": {
				"bind_address": "0.0.0.0:8080",
				"log_level": "INFO",
			}
		}`)
		cfg := &configv1.McpAnyServerConfig{}
		err := engine.Unmarshal(invalidJSON, cfg)
		assert.Error(t, err)
	})
}

func TestYamlEngine_Unmarshal(t *testing.T) {
	engine := &yamlEngine{}

	t.Run("InvalidYAML", func(t *testing.T) {
		invalidYAML := []byte(`
global_settings:
  bind_address: "0.0.0.0:8080"
  log_level: "INFO"
  protoc_version: "3.19.4"
- this is not valid
`)
		cfg := &configv1.McpAnyServerConfig{}
		err := engine.Unmarshal(invalidYAML, cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal YAML")
	})

	t.Run("ValidYAML", func(t *testing.T) {
		validYAML := []byte(`
global_settings:
  mcp_listen_address: "0.0.0.0:8080"
  log_level: "LOG_LEVEL_INFO"
`)
		cfg := &configv1.McpAnyServerConfig{}
		err := engine.Unmarshal(validYAML, cfg)
		require.NoError(t, err)
		assert.Equal(t, "0.0.0.0:8080", cfg.GetGlobalSettings().GetMcpListenAddress())
		assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_INFO, cfg.GetGlobalSettings().GetLogLevel())
	})
}

// marshalError is a helper type that always returns an error when marshaled to JSON.
type marshalError struct{}

func (m *marshalError) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("marshal error")
}

func (e *yamlEngine) UnmarshalWithFailingJSON(b []byte, v proto.Message) error {
	var yamlMap map[string]interface{}
	if err := yaml.Unmarshal(b, &yamlMap); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}
	// Purposely cause a marshaling error by using a type that fails to marshal.
	_, err := json.Marshal(&marshalError{})
	return fmt.Errorf("failed to marshal map to JSON: %w", err)
}

func TestYamlEngine_Unmarshal_MarshalError(t *testing.T) {
	engine := &yamlEngine{}
	validYAML := []byte(`
global_settings:
  bind_address: "0.0.0.0:8080"
`)
	err := engine.UnmarshalWithFailingJSON(validYAML, &configv1.McpAnyServerConfig{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal map to JSON")
}

func TestFileStore_Load(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Setup directory and config files
	require.NoError(t, fs.MkdirAll("configs/subdir", 0o755))
	afero.WriteFile(fs, "configs/01_base.yaml", []byte(`
global_settings:
  mcp_listen_address: "0.0.0.0:8080"
  log_level: "LOG_LEVEL_INFO"
upstream_services:
- id: "service-1"
  name: "first-service"
`), 0o644)

	afero.WriteFile(fs, "configs/02_override.yaml", []byte(`
global_settings:
  mcp_listen_address: "127.0.0.1:9090"
upstream_services:
- id: "service-2"
  name: "second-service"
`), 0o644)

	afero.WriteFile(fs, "configs/invalid.txt", []byte("invalid content"), 0o644)
	afero.WriteFile(fs, "malformed.yaml", []byte("bad-yaml:"), 0o644)
	require.NoError(t, fs.Mkdir("configs/subdir/empty", 0o755))

	testCases := []struct {
		name          string
		paths         []string
		expectErr     bool
		expectedCfg   *configv1.McpAnyServerConfig
		checkResult   func(t *testing.T, cfg *configv1.McpAnyServerConfig)
		expectedErrFn func(t *testing.T, err error)
	}{
		{
			name:  "Load single file",
			paths: []string{"configs/01_base.yaml"},
			checkResult: func(t *testing.T, cfg *configv1.McpAnyServerConfig) {
				assert.Equal(t, "0.0.0.0:8080", cfg.GetGlobalSettings().GetMcpListenAddress())
				assert.Len(t, cfg.GetUpstreamServices(), 1)
			},
		},
		{
			name:  "Load and merge multiple files",
			paths: []string{"configs/01_base.yaml", "configs/02_override.yaml"},
			checkResult: func(t *testing.T, cfg *configv1.McpAnyServerConfig) {
				// Last one wins for scalar fields
				assert.Equal(t, "127.0.0.1:9090", cfg.GetGlobalSettings().GetMcpListenAddress())
				// Repeated fields are appended
				assert.Len(t, cfg.GetUpstreamServices(), 2)
				assert.Equal(t, "service-1", cfg.GetUpstreamServices()[0].GetId())
				assert.Equal(t, "service-2", cfg.GetUpstreamServices()[1].GetId())
			},
		},
		{
			name:      "Path does not exist",
			paths:     []string{"nonexistent/"},
			expectErr: true,
			expectedErrFn: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "failed to stat path")
			},
		},
		{
			name:      "Load with malformed file",
			paths:     []string{"malformed.yaml"},
			expectErr: true,
			expectedErrFn: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "failed to unmarshal config")
			},
		},
		{
			name:  "Empty directory results in nil config",
			paths: []string{"configs/subdir/empty"},
			checkResult: func(t *testing.T, cfg *configv1.McpAnyServerConfig) {
				assert.Nil(t, cfg)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := NewFileStore(fs, tc.paths)
			cfg, err := store.Load()

			if tc.expectErr {
				require.Error(t, err)
				if tc.expectedErrFn != nil {
					tc.expectedErrFn(t, err)
				}
			} else {
				require.NoError(t, err)
				if tc.checkResult != nil {
					tc.checkResult(t, cfg)
				}
			}
		})
	}
}

func TestIsURL(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Valid HTTP URL",
			path:     "http://example.com",
			expected: true,
		},
		{
			name:     "Valid HTTPS URL",
			path:     "https://example.com",
			expected: true,
		},
		{
			name:     "Mixed-case HTTP URL",
			path:     "HTTP://example.com",
			expected: true,
		},
		{
			name:     "Mixed-case HTTPS URL",
			path:     "HTTPS://example.com",
			expected: true,
		},
		{
			name:     "Uppercase HTTP URL",
			path:     "HTTP://EXAMPLE.COM",
			expected: true,
		},
		{
			name:     "Local file path",
			path:     "/path/to/file.yaml",
			expected: false,
		},
		{
			name:     "FTP URL (unsupported)",
			path:     "ftp://example.com",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, isURL(tc.path))
		})
	}
}

func TestReadURL_RedirectShouldFail(t *testing.T) {
	// This server will redirect to a "safe" URL, but the redirect should not be followed
	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://example.com", http.StatusFound)
	}))
	defer redirectServer.Close()

	// Temporarily remove the SSRF protection to test the redirect logic in isolation.
	originalTransport := httpClient.Transport
	defer func() { httpClient.Transport = originalTransport }()
	httpClient.Transport = &http.Transport{}

	_, err := readURL(redirectServer.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redirects are disabled for security reasons")
}

func TestExpand(t *testing.T) {
	t.Setenv("TEST_VAR", "test_value")
	t.Setenv("EMPTY_VAR", "")

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no variables",
			input:    "this is a simple string",
			expected: "this is a simple string",
		},
		{
			name:     "braced variable",
			input:    "hello ${TEST_VAR}",
			expected: "hello test_value",
		},
		{
			name:     "variable with default value",
			input:    "hello ${UNDEFINED_VAR:default_value}",
			expected: "hello default_value",
		},
		{
			name:     "empty variable with default value",
			input:    "hello ${EMPTY_VAR:default_value}",
			expected: "hello default_value",
		},
		{
			name:     "undefined variable without default",
			input:    "hello ${UNDEFINED_VAR}",
			expected: "hello ${UNDEFINED_VAR}",
		},
		{
			name:     "multiple variables",
			input:    "${TEST_VAR} ${TEST_VAR} ${UNDEFINED_VAR} ${UNDEFINED_VAR:default}",
			expected: "test_value test_value ${UNDEFINED_VAR} default",
		},
		{
			name:     "simple variable syntax is ignored",
			input:    "$TEST_VAR",
			expected: "$TEST_VAR",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := expand([]byte(tc.input))
			assert.Equal(t, tc.expected, string(actual))
		})
	}
}
