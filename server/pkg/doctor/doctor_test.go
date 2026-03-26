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
	"google.golang.org/protobuf/proto"
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

	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("valid-http"),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: strPtr(ts.URL),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("invalid-http"),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: strPtr("http://127.0.0.1:12345/nonexistent"),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name:    strPtr("disabled-service"),
				Disable: boolPtr(true),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: strPtr(ts.URL),
				}.Build(),
			}.Build(),
		},
	}.Build()

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

	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("valid-grpc"),
				GrpcService: configv1.GrpcUpstreamService_builder{
					Address: strPtr(addr),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("invalid-grpc"),
				GrpcService: configv1.GrpcUpstreamService_builder{
					Address: strPtr("127.0.0.1:1"), // Unlikely port
				}.Build(),
			}.Build(),
		},
	}.Build()

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

	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("valid-openapi"),
				OpenapiService: configv1.OpenapiUpstreamService_builder{
					SpecUrl: proto.String(ts.URL),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("invalid-openapi"),
				OpenapiService: configv1.OpenapiUpstreamService_builder{
					SpecUrl: proto.String("http://127.0.0.1:12345/nonexistent"),
				}.Build(),
			}.Build(),
		},
	}.Build()

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

	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("auth-ok"),
				HttpService: configv1.HttpUpstreamService_builder{Address: strPtr(ts.URL)}.Build(), // Dummy service URL
				UpstreamAuth: configv1.Authentication_builder{
					Oauth2: configv1.OAuth2Auth_builder{
						TokenUrl: strPtr(ts.URL + "/token-ok"),
					}.Build(),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("auth-400"),
				HttpService: configv1.HttpUpstreamService_builder{Address: strPtr(ts.URL)}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					Oauth2: configv1.OAuth2Auth_builder{
						TokenUrl: strPtr(ts.URL + "/token-400"),
					}.Build(),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("auth-401"),
				HttpService: configv1.HttpUpstreamService_builder{Address: strPtr(ts.URL)}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					Oauth2: configv1.OAuth2Auth_builder{
						TokenUrl: strPtr(ts.URL + "/token-401"),
					}.Build(),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("auth-404"),
				HttpService: configv1.HttpUpstreamService_builder{Address: strPtr(ts.URL)}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					Oauth2: configv1.OAuth2Auth_builder{
						TokenUrl: strPtr(ts.URL + "/token-404"),
					}.Build(),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("auth-500"),
				HttpService: configv1.HttpUpstreamService_builder{Address: strPtr(ts.URL)}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					Oauth2: configv1.OAuth2Auth_builder{
						TokenUrl: strPtr(ts.URL + "/token-500"),
					}.Build(),
				}.Build(),
			}.Build(),
		},
	}.Build()

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

	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("valid-fs"),
				FilesystemService: configv1.FilesystemUpstreamService_builder{
					// RootPaths is a map, direct set or use builder if it has map field. Usually yes.
					// Repeated/Map fields in builder?
					// Usually exposed as field.
					RootPaths: map[string]string{
						"/data": tmpDir,
					},
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("invalid-fs"),
				FilesystemService: configv1.FilesystemUpstreamService_builder{
					RootPaths: map[string]string{
						"/data": filepath.Join(tmpDir, "nonexistent"),
					},
				}.Build(),
			}.Build(),
		},
	}.Build()

	results := RunChecks(context.Background(), config)

	assert.Len(t, results, 2)
	assert.Equal(t, StatusOk, results[0].Status)
	assert.Equal(t, StatusError, results[1].Status)
}

func TestRunChecks_CommandLine(t *testing.T) {
	// Assume "ls" (or "dir" on windows) exists. Docker env is linux.
	cmd := "ls"
	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("valid-cmd"),
				CommandLineService: configv1.CommandLineUpstreamService_builder{
					Command: strPtr(cmd),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("invalid-cmd"),
				CommandLineService: configv1.CommandLineUpstreamService_builder{
					Command: strPtr("nonexistentcommand12345"),
				}.Build(),
			}.Build(),
		},
	}.Build()

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

	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("valid-mcp-http"),
				McpService: configv1.McpUpstreamService_builder{
					HttpConnection: configv1.McpStreamableHttpConnection_builder{
						HttpAddress: strPtr(ts.URL),
					}.Build(),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("valid-mcp-stdio"),
				McpService: configv1.McpUpstreamService_builder{
					StdioConnection: configv1.McpStdioConnection_builder{
						Command: strPtr("ls"),
					}.Build(),
				}.Build(),
			}.Build(),
		},
	}.Build()

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

	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("valid-ws"),
				WebsocketService: configv1.WebsocketUpstreamService_builder{
					Address: strPtr(wsURL),
				}.Build(),
			}.Build(),
		},
	}.Build()

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

	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("oidc-service"),
				HttpService: configv1.HttpUpstreamService_builder{Address: strPtr(ts.URL)}.Build(),
				UpstreamAuth: configv1.Authentication_builder{
					Oidc: configv1.OIDCAuth_builder{
						Issuer: strPtr(ts.URL),
					}.Build(),
				}.Build(),
			}.Build(),
		},
	}.Build()

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

	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("graphql-service"),
				GraphqlService: configv1.GraphQLUpstreamService_builder{
					Address: strPtr(ts.URL),
				}.Build(),
			}.Build(),
		},
	}.Build()

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

	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("webrtc-service"),
				WebrtcService: configv1.WebrtcUpstreamService_builder{
					Address: strPtr(ts.URL),
				}.Build(),
			}.Build(),
		},
	}.Build()

	results := RunChecks(context.Background(), config)
	assert.Len(t, results, 1)
	assert.Equal(t, StatusOk, results[0].Status)
}

func TestRunChecks_SQL(t *testing.T) {
	// DSN: ":memory:" for sqlite
	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("sql-service"),
				SqlService: configv1.SqlUpstreamService_builder{
					Driver: strPtr("sqlite"),
					Dsn:    strPtr(":memory:"),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name: strPtr("invalid-sql-driver"),
				SqlService: configv1.SqlUpstreamService_builder{
					Driver: strPtr("invalid-driver"),
					Dsn:    strPtr(":memory:"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	results := RunChecks(context.Background(), config)
	assert.Len(t, results, 2)
	assert.Equal(t, StatusOk, results[0].Status)
	assert.Equal(t, StatusError, results[1].Status)
}
