// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestValidate_Security_VolumeMounts(t *testing.T) {
	// This test reproduces a security vulnerability where insecure volume mounts
	// (using ".." traversal) are allowed in the container environment configuration.
	// We first assert that it IS allowed (proving the issue), then we will fix it
	// and update the assertion.

	jsonConfig := `{
		"upstream_services": [
			{
				"name": "malicious-cmd-svc",
				"command_line_service": {
					"command": "echo hacked",
					"container_environment": {
						"image": "ubuntu",
						"volumes": {
							"../../../etc/passwd": "/target"
						}
					}
				}
			}
		]
	}`

	cfg := &configv1.McpAnyServerConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(jsonConfig), cfg))

	validationErrors := Validate(context.Background(), cfg, Server)

	// We expect validation errors now because the vulnerability is fixed
	require.NotEmpty(t, validationErrors, "Expected validation errors for insecure volume mount")
	assert.Contains(t, validationErrors[0].Error(), "is not a secure path")
	assert.Contains(t, validationErrors[0].Error(), "container environment volume host path")
}

func TestValidate_Security_URLSchemes(t *testing.T) {
	testCases := []struct {
		name          string
		configJSON    string
		expectedError string
	}{
		{
			name: "mcp_http_connection_ftp_scheme",
			configJSON: `{
				"upstream_services": [{
					"name": "bad-mcp-http",
					"mcp_service": {
						"http_connection": {
							"http_address": "ftp://example.com/file"
						}
					}
				}]
			}`,
			expectedError: "mcp service with http_connection has invalid http_address scheme: ftp",
		},
		{
			name: "openapi_spec_url_ftp_scheme",
			configJSON: `{
				"upstream_services": [{
					"name": "bad-openapi",
					"openapi_service": {
						"spec_url": "ftp://example.com/spec"
					}
				}]
			}`,
			expectedError: "invalid openapi spec_url scheme: ftp",
		},
		{
			name: "audit_webhook_url_ftp_scheme",
			configJSON: `{
				"global_settings": {
					"audit": {
						"enabled": true,
						"storage_type": "STORAGE_TYPE_WEBHOOK",
						"webhook_url": "ftp://example.com/webhook"
					}
				}
			}`,
			expectedError: "invalid webhook_url scheme: ftp",
		},
		{
			name: "oauth2_token_url_ftp_scheme",
			configJSON: `{
				"upstream_services": [{
					"name": "bad-oauth",
					"http_service": {
						"address": "http://example.com"
					},
					"upstream_auth": {
						"oauth2": {
							"token_url": "ftp://example.com/token"
						}
					}
				}]
			}`,
			expectedError: "invalid oauth2 token_url scheme: ftp",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &configv1.McpAnyServerConfig{}
			require.NoError(t, protojson.Unmarshal([]byte(tc.configJSON), cfg))

			validationErrors := Validate(context.Background(), cfg, Server)
			require.NotEmpty(t, validationErrors, "Expected validation errors")

			found := false
			for _, err := range validationErrors {
				if err.Err != nil && assert.Contains(t, err.Err.Error(), tc.expectedError) {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected error %q not found in %v", tc.expectedError, validationErrors)
		})
	}
}
