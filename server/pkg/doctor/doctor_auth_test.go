// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package doctor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestRunChecks_Authentication_ActiveVerification(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	// Start a mock HTTP server that enforces authentication
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check API Key
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "secret123" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Check Bearer Token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "Bearer valid-token" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Check Basic Auth
		user, pass, ok := r.BasicAuth()
		if ok && user == "user" && pass == "pass" {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("api-key-success"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtr(ts.URL),
					},
				},
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_ApiKey{
						ApiKey: &configv1.APIKeyAuth{
							ParamName: strPtr("X-API-Key"),
							Value: &configv1.SecretValue{
								Value: &configv1.SecretValue_PlainText{
									PlainText: "secret123",
								},
							},
						},
					},
				},
			},
			{
				Name: strPtr("api-key-fail"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtr(ts.URL),
					},
				},
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_ApiKey{
						ApiKey: &configv1.APIKeyAuth{
							ParamName: strPtr("X-API-Key"),
							Value: &configv1.SecretValue{
								Value: &configv1.SecretValue_PlainText{
									PlainText: "wrong-key",
								},
							},
						},
					},
				},
			},
			{
				Name: strPtr("bearer-success"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtr(ts.URL),
					},
				},
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_BearerToken{
						BearerToken: &configv1.BearerTokenAuth{
							Token: &configv1.SecretValue{
								Value: &configv1.SecretValue_PlainText{
									PlainText: "valid-token",
								},
							},
						},
					},
				},
			},
			{
				Name: strPtr("bearer-fail"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtr(ts.URL),
					},
				},
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_BearerToken{
						BearerToken: &configv1.BearerTokenAuth{
							Token: &configv1.SecretValue{
								Value: &configv1.SecretValue_PlainText{
									PlainText: "invalid-token",
								},
							},
						},
					},
				},
			},
			{
				Name: strPtr("basic-success"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtr(ts.URL),
					},
				},
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_BasicAuth{
						BasicAuth: &configv1.BasicAuth{
							Username: strPtr("user"),
							Password: &configv1.SecretValue{
								Value: &configv1.SecretValue_PlainText{
									PlainText: "pass",
								},
							},
						},
					},
				},
			},
			{
				Name: strPtr("basic-fail"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtr(ts.URL),
					},
				},
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_BasicAuth{
						BasicAuth: &configv1.BasicAuth{
							Username: strPtr("user"),
							Password: &configv1.SecretValue{
								Value: &configv1.SecretValue_PlainText{
									PlainText: "wrongpass",
								},
							},
						},
					},
				},
			},
			{
				Name: strPtr("no-auth-401"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtr(ts.URL),
					},
				},
				// No Auth configured
			},
		},
	}

	results := RunChecks(context.Background(), config)

	assert.Len(t, results, 7)

	// API Key
	assert.Equal(t, "api-key-success", results[0].ServiceName)
	assert.Equal(t, StatusOk, results[0].Status)
	assert.Equal(t, "api-key-fail", results[1].ServiceName)
	assert.Equal(t, StatusError, results[1].Status)
	assert.Contains(t, results[1].Message, "Authentication failed (401 Unauthorized)")

	// Bearer Token
	assert.Equal(t, "bearer-success", results[2].ServiceName)
	assert.Equal(t, StatusOk, results[2].Status)
	assert.Equal(t, "bearer-fail", results[3].ServiceName)
	assert.Equal(t, StatusError, results[3].Status)

	// Basic Auth
	assert.Equal(t, "basic-success", results[4].ServiceName)
	assert.Equal(t, StatusOk, results[4].Status)
	assert.Equal(t, "basic-fail", results[5].ServiceName)
	assert.Equal(t, StatusError, results[5].Status)

	// No Auth
	assert.Equal(t, "no-auth-401", results[6].ServiceName)
	assert.Equal(t, StatusWarning, results[6].Status)
	assert.Contains(t, results[6].Message, "Service reachable but returned: 401 Unauthorized")
}
