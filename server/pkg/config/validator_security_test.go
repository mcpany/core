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
	"google.golang.org/protobuf/proto"
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

func TestValidate_Security_SSRF(t *testing.T) {
	testCases := []struct {
		name        string
		address     string
		expectError bool
		errorMsg    string
	}{
		{"localhost", "http://localhost:8080", true, "localhost is not allowed"},
		{"private_ip", "http://10.0.0.1:8080", true, "private IP \"10.0.0.1\" is not allowed"},
		{"cloud_metadata", "http://169.254.169.254", true, "link-local IP \"169.254.169.254\" is not allowed"},
		{"public_ip", "http://8.8.8.8", false, ""},
		{"google", "https://google.com", false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("ssrf-test"),
						ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
							HttpService: &configv1.HttpUpstreamService{
								Address: proto.String(tc.address),
							},
						},
					},
				},
			}

			errs := Validate(context.Background(), cfg, Server)
			if tc.expectError {
				require.NotEmpty(t, errs, "Expected validation errors")
				assert.Contains(t, errs[0].Error(), tc.errorMsg)
			} else {
				assert.Empty(t, errs, "Expected no validation errors")
			}
		})
	}
}
