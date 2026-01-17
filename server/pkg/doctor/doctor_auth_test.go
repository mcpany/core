// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package doctor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRunChecks_Auth(t *testing.T) {
	// Start a mock HTTP server that requires authentication
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for API Key
		if r.Header.Get("X-API-Key") == "valid-secret" {
			w.WriteHeader(http.StatusOK)
			return
		}
		// Check for Bearer Token
		if r.Header.Get("Authorization") == "Bearer valid-token" {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	// Set env var for secret resolution
	os.Setenv("TEST_API_KEY", "valid-secret")
	os.Setenv("TEST_TOKEN", "valid-token")
	defer os.Unsetenv("TEST_API_KEY")
	defer os.Unsetenv("TEST_TOKEN")

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("auth-http-api-key"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtr(ts.URL),
					},
				},
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_ApiKey{
						ApiKey: &configv1.APIKeyAuth{
							ParamName: proto.String("X-API-Key"),
							Value: &configv1.SecretValue{
								Value: &configv1.SecretValue_EnvironmentVariable{
									EnvironmentVariable: "TEST_API_KEY",
								},
							},
						},
					},
				},
			},
			{
				Name: strPtr("auth-http-bearer"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtr(ts.URL),
					},
				},
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_BearerToken{
						BearerToken: &configv1.BearerTokenAuth{
							Token: &configv1.SecretValue{
								Value: &configv1.SecretValue_EnvironmentVariable{
									EnvironmentVariable: "TEST_TOKEN",
								},
							},
						},
					},
				},
			},
			{
				Name: strPtr("no-auth-fail"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtr(ts.URL),
					},
				},
				// No auth provided, should fail (or warn 401)
			},
		},
	}

	results := RunChecks(context.Background(), config)

	assert.Len(t, results, 3)

	// Check API Key service
	assert.Equal(t, "auth-http-api-key", results[0].ServiceName)
	assert.Equal(t, StatusOk, results[0].Status, "Expected OK for valid API Key")

	// Check Bearer Token service
	assert.Equal(t, "auth-http-bearer", results[1].ServiceName)
	assert.Equal(t, StatusOk, results[1].Status, "Expected OK for valid Bearer Token")

	// Check no-auth service
	assert.Equal(t, "no-auth-fail", results[2].ServiceName)
	// Currently it returns Warning for 4xx.
	assert.Equal(t, StatusWarning, results[2].Status, "Expected Warning for missing auth")
	assert.Contains(t, results[2].Message, "401")
}
