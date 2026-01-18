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
	// Start a mock HTTP server that enforces authentication
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "secret123" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer ts.Close()

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("auth-success"),
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
				Name: strPtr("auth-failure-wrong-key"),
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
				Name: strPtr("auth-missing-but-required"),
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

	assert.Len(t, results, 3)

	// 1. Success
	assert.Equal(t, "auth-success", results[0].ServiceName)
	assert.Equal(t, StatusOk, results[0].Status)
	assert.Contains(t, results[0].Message, "Service reachable")

	// 2. Failure (Auth Provided but Rejected)
	assert.Equal(t, "auth-failure-wrong-key", results[1].ServiceName)
	assert.Equal(t, StatusError, results[1].Status)
	assert.Contains(t, results[1].Message, "Authentication failed (401 Unauthorized)")

	// 3. Warning (No Auth Provided, Server says 401)
	assert.Equal(t, "auth-missing-but-required", results[2].ServiceName)
	assert.Equal(t, StatusWarning, results[2].Status)
	// Message should indicate it is reachable but returned 401
	assert.Contains(t, results[2].Message, "Service reachable but returned: 401 Unauthorized")
}
