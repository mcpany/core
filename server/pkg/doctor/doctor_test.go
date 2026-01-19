// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package doctor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func TestRunChecks_Http(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	// Start a mock HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("valid-http"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtr(ts.URL),
					},
				},
			},
			{
				Name: strPtr("invalid-http"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtr("http://localhost:12345/nonexistent"),
					},
				},
			},
			{
				Name:    strPtr("disabled-service"),
				Disable: boolPtr(true),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtr(ts.URL),
					},
				},
			},
		},
	}

	results := RunChecks(context.Background(), config)

	assert.Len(t, results, 3)

	// Check valid service
	assert.Equal(t, "valid-http", results[0].ServiceName)
	assert.Equal(t, StatusOk, results[0].Status)

	// Check invalid service
	assert.Equal(t, "invalid-http", results[1].ServiceName)
	assert.Equal(t, StatusError, results[1].Status)

	// Check disabled service
	assert.Equal(t, "disabled-service", results[2].ServiceName)
	assert.Equal(t, StatusSkipped, results[2].Status)
}

func TestRunChecks_Grpc(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	// We can cheat and use the HTTP listener address for TCP check
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	// Extract host:port from ts.URL (http://127.0.0.1:xxxxx)
	addr := ts.Listener.Addr().String()

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("valid-grpc"),
				ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
					GrpcService: &configv1.GrpcUpstreamService{
						Address: strPtr(addr),
					},
				},
			},
			{
				Name: strPtr("invalid-grpc"),
				ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
					GrpcService: &configv1.GrpcUpstreamService{
						Address: strPtr("localhost:1"), // Unlikely port
					},
				},
			},
		},
	}

	results := RunChecks(context.Background(), config)

	assert.Len(t, results, 2)
	assert.Equal(t, StatusOk, results[0].Status)
	assert.Equal(t, StatusError, results[1].Status)
}

func TestRunChecks_OpenAPI(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("valid-openapi"),
				ServiceConfig: &configv1.UpstreamServiceConfig_OpenapiService{
					OpenapiService: &configv1.OpenapiUpstreamService{
						SpecSource: &configv1.OpenapiUpstreamService_SpecUrl{
							SpecUrl: ts.URL,
						},
					},
				},
			},
			{
				Name: strPtr("invalid-openapi"),
				ServiceConfig: &configv1.UpstreamServiceConfig_OpenapiService{
					OpenapiService: &configv1.OpenapiUpstreamService{
						SpecSource: &configv1.OpenapiUpstreamService_SpecUrl{
							SpecUrl: "http://localhost:12345/nonexistent",
						},
					},
				},
			},
		},
	}

	results := RunChecks(context.Background(), config)

	assert.Len(t, results, 2)
	assert.Equal(t, StatusOk, results[0].Status)
	assert.Equal(t, StatusError, results[1].Status)
}

func TestRunChecks_Authentication_OAuth2(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	// Mock OAuth2 Token Endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// Simulate different behaviors based on path
		switch r.URL.Path {
		case "/token-ok":
			w.WriteHeader(http.StatusOK)
		case "/token-400":
			w.WriteHeader(http.StatusBadRequest)
		case "/token-401":
			w.WriteHeader(http.StatusUnauthorized)
		case "/token-404":
			w.WriteHeader(http.StatusNotFound)
		case "/token-500":
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("auth-ok"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{Address: strPtr(ts.URL)}, // Dummy service URL
				},
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_Oauth2{
						Oauth2: &configv1.OAuth2Auth{
							TokenUrl: strPtr(ts.URL + "/token-ok"),
						},
					},
				},
			},
			{
				Name: strPtr("auth-400"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{Address: strPtr(ts.URL)},
				},
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_Oauth2{
						Oauth2: &configv1.OAuth2Auth{
							TokenUrl: strPtr(ts.URL + "/token-400"),
						},
					},
				},
			},
			{
				Name: strPtr("auth-401"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{Address: strPtr(ts.URL)},
				},
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_Oauth2{
						Oauth2: &configv1.OAuth2Auth{
							TokenUrl: strPtr(ts.URL + "/token-401"),
						},
					},
				},
			},
			{
				Name: strPtr("auth-404"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{Address: strPtr(ts.URL)},
				},
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_Oauth2{
						Oauth2: &configv1.OAuth2Auth{
							TokenUrl: strPtr(ts.URL + "/token-404"),
						},
					},
				},
			},
			{
				Name: strPtr("auth-500"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{Address: strPtr(ts.URL)},
				},
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_Oauth2{
						Oauth2: &configv1.OAuth2Auth{
							TokenUrl: strPtr(ts.URL + "/token-500"),
						},
					},
				},
			},
		},
	}

	results := RunChecks(context.Background(), config)

	assert.Len(t, results, 5)

	// Auth OK -> OK
	assert.Equal(t, "auth-ok", results[0].ServiceName)
	assert.Equal(t, StatusOk, results[0].Status)
	assert.Contains(t, results[0].Message, "OAuth2 Reachable (200)")

	// Auth 400 -> OK (Expected for empty POST)
	assert.Equal(t, "auth-400", results[1].ServiceName)
	assert.Equal(t, StatusOk, results[1].Status)
	assert.Contains(t, results[1].Message, "OAuth2 Reachable (400)")

	// Auth 401 -> OK
	assert.Equal(t, "auth-401", results[2].ServiceName)
	assert.Equal(t, StatusOk, results[2].Status)
	assert.Contains(t, results[2].Message, "OAuth2 Reachable (401)")

	// Auth 404 -> Error
	assert.Equal(t, "auth-404", results[3].ServiceName)
	assert.Equal(t, StatusError, results[3].Status)
	assert.Contains(t, results[3].Message, "not found (404)")

	// Auth 500 -> Error
	assert.Equal(t, "auth-500", results[4].ServiceName)
	assert.Equal(t, StatusError, results[4].Status)
	assert.Contains(t, results[4].Message, "server error")
}

func TestRunChecks_Filesystem(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "doctor-fs")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("valid-fs"),
				ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
					FilesystemService: &configv1.FilesystemUpstreamService{
						RootPaths: map[string]string{
							"/data": tmpDir,
						},
					},
				},
			},
			{
				Name: strPtr("invalid-fs"),
				ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
					FilesystemService: &configv1.FilesystemUpstreamService{
						RootPaths: map[string]string{
							"/data": filepath.Join(tmpDir, "nonexistent"),
						},
					},
				},
			},
		},
	}

	results := RunChecks(context.Background(), config)

	assert.Len(t, results, 2)
	assert.Equal(t, StatusOk, results[0].Status)
	assert.Equal(t, StatusError, results[1].Status)
}

func TestRunChecks_CommandLine(t *testing.T) {
	// Assume "ls" (or "dir" on windows) exists. Docker env is linux.
	cmd := "ls"
	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("valid-cmd"),
				ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
					CommandLineService: &configv1.CommandLineUpstreamService{
						Command: strPtr(cmd),
					},
				},
			},
			{
				Name: strPtr("invalid-cmd"),
				ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
					CommandLineService: &configv1.CommandLineUpstreamService{
						Command: strPtr("nonexistentcommand12345"),
					},
				},
			},
		},
	}

	results := RunChecks(context.Background(), config)

	assert.Len(t, results, 2)
	assert.Equal(t, StatusOk, results[0].Status)
	assert.Equal(t, StatusError, results[1].Status)
}

func TestRunChecks_MCP(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("valid-mcp-http"),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_HttpConnection{
							HttpConnection: &configv1.McpStreamableHttpConnection{
								HttpAddress: strPtr(ts.URL),
							},
						},
					},
				},
			},
			{
				Name: strPtr("valid-mcp-stdio"),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_StdioConnection{
							StdioConnection: &configv1.McpStdioConnection{
								Command: strPtr("ls"),
							},
						},
					},
				},
			},
		},
	}

	results := RunChecks(context.Background(), config)

	assert.Len(t, results, 2)
	assert.Equal(t, StatusOk, results[0].Status)
	assert.Equal(t, StatusOk, results[1].Status)
}

func TestRunChecks_WebSocket(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Convert http:// to ws://
	wsURL := "ws" + ts.URL[4:]

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("valid-ws"),
				ServiceConfig: &configv1.UpstreamServiceConfig_WebsocketService{
					WebsocketService: &configv1.WebsocketUpstreamService{
						Address: strPtr(wsURL),
					},
				},
			},
		},
	}

	results := RunChecks(context.Background(), config)

	assert.Len(t, results, 1)
	assert.Equal(t, StatusOk, results[0].Status)
}

func TestRunChecks_OIDC(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("oidc-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{Address: strPtr(ts.URL)},
				},
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_Oidc{
						Oidc: &configv1.OIDCAuth{
							Issuer: strPtr(ts.URL),
						},
					},
				},
			},
		},
	}

	results := RunChecks(context.Background(), config)
	assert.Len(t, results, 1)
	assert.Equal(t, StatusOk, results[0].Status)
}

func TestRunChecks_GraphQL(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("graphql-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_GraphqlService{
					GraphqlService: &configv1.GraphQLUpstreamService{
						Address: strPtr(ts.URL),
					},
				},
			},
		},
	}

	results := RunChecks(context.Background(), config)
	assert.Len(t, results, 1)
	assert.Equal(t, StatusOk, results[0].Status)
}

func TestRunChecks_WebRTC(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("webrtc-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_WebrtcService{
					WebrtcService: &configv1.WebrtcUpstreamService{
						Address: strPtr(ts.URL),
					},
				},
			},
		},
	}

	results := RunChecks(context.Background(), config)
	assert.Len(t, results, 1)
	assert.Equal(t, StatusOk, results[0].Status)
}

func TestRunChecks_SQL(t *testing.T) {
	// DSN: ":memory:" for sqlite
	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("sql-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_SqlService{
					SqlService: &configv1.SqlUpstreamService{
						Driver: strPtr("sqlite"),
						Dsn:    strPtr(":memory:"),
					},
				},
			},
			{
				Name: strPtr("invalid-sql-driver"),
				ServiceConfig: &configv1.UpstreamServiceConfig_SqlService{
					SqlService: &configv1.SqlUpstreamService{
						Driver: strPtr("invalid-driver"),
						Dsn:    strPtr(":memory:"),
					},
				},
			},
		},
	}

	results := RunChecks(context.Background(), config)
	assert.Len(t, results, 2)
	assert.Equal(t, StatusOk, results[0].Status)
	assert.Equal(t, StatusError, results[1].Status)
}
