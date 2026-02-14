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

func TestValidate_Security_StdioArgs_SSRF_Repro(t *testing.T) {
	// This test reproduces an SSRF vulnerability where local/private URLs
	// are allowed as arguments for interpreters like Deno/Bun.
	// We expect this to FAIL (pass validation) initially, proving the vulnerability.

	jsonConfig := `{
		"upstream_services": [
			{
				"name": "ssrf-deno-service",
				"mcp_service": {
					"stdio_connection": {
						"command": "deno",
						"args": ["run", "http://127.0.0.1:8080/internal-script.ts"]
					}
				}
			}
		]
	}`

	cfg := configv1.McpAnyServerConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(jsonConfig), cfg))

	// Mock execLookPath to ensure command validation passes
	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		return "/usr/bin/deno", nil
	}

	// 4. Validate the config.
	validationErrors := Validate(context.Background(), cfg, Server)

	require.NotEmpty(t, validationErrors, "Expected validation errors for unsafe URL")
	if len(validationErrors) > 0 {
		assert.Contains(t, validationErrors[0].Error(), "unsafe url", "Expected SSRF protection")
	}
}

func TestValidate_Security_StdioArgs_SSRF_Allowed(t *testing.T) {
	// This test ensures that safe public URLs are allowed.
	// We use a public IP address to avoid DNS resolution dependency (flakiness).
	// 8.8.8.8 is Google Public DNS, definitely not private/local.

	jsonConfig := `{
		"upstream_services": [
			{
				"name": "valid-deno-service",
				"mcp_service": {
					"stdio_connection": {
						"command": "deno",
						"args": ["run", "http://8.8.8.8/script.ts"]
					}
				}
			}
		]
	}`

	cfg := configv1.McpAnyServerConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(jsonConfig), cfg))

	// Mock execLookPath
	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		return "/usr/bin/deno", nil
	}

	validationErrors := Validate(context.Background(), cfg, Server)
	assert.Empty(t, validationErrors, "Expected no validation errors for safe public URL")
}
