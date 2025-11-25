// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCommand(t *testing.T) {
	// Helper function to create a temporary config file
	createTempConfigFile := func(t *testing.T, content string) string {
		t.Helper()
		tmpfile, err := os.CreateTemp("", "config-*.yaml")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tmpfile.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
		if err := tmpfile.Close(); err != nil {
			t.Fatal(err)
		}
		return tmpfile.Name()
	}

	testCases := []struct {
		name          string
		configContent string
		expectedErr   string
		expectSuccess bool
	}{
		{
			name: "Valid config",
			configContent: `
global_settings:
  mcp_listen_address: ":8080"
upstream_services:
  - name: "my-http-service"
    http_service:
      address: "http://localhost:8081"
`,
			expectSuccess: true,
		},
		{
			name: "Missing MCP listen address",
			configContent: `
global_settings: {}
`,
			expectedErr: "global_settings.mcp_listen_address is not set",
		},
		{
			name: "Missing upstream service name",
			configContent: `
global_settings:
  mcp_listen_address: ":8080"
upstream_services:
  - http_service:
      address: "http://localhost:8081"
`,
			expectedErr: "upstream_services[0].name is not set",
		},
		{
			name: "Missing upstream service address",
			configContent: `
global_settings:
  mcp_listen_address: ":8080"
upstream_services:
  - name: "my-http-service"
    http_service: {}
`,
			expectedErr: "upstream_services[0].http_service.address is not set",
		},
		{
			name: "Nil service config",
			configContent: `
global_settings:
  mcp_listen_address: ":8080"
upstream_services:
  - name: "my-service"
`,
			expectedErr: "upstream_services[0].service_config is not set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configFile := createTempConfigFile(t, tc.configContent)
			defer os.Remove(configFile)

			rootCmd := newRootCmd()
			validateCmd, _, err := rootCmd.Find([]string{"validate"})
			assert.NoError(t, err)

			// Redirect stdout and stderr
			var outBuf, errBuf bytes.Buffer
			validateCmd.SetOut(&outBuf)
			validateCmd.SetErr(&errBuf)

			// Set the config path
			validateCmd.Flags().Set("config-paths", configFile)

			// Execute the command
			err = validateCmd.Execute()

			if tc.expectSuccess {
				assert.NoError(t, err)
				assert.Contains(t, outBuf.String(), "Configuration is valid.")
			} else {
				assert.Error(t, err)
				assert.Contains(t, errBuf.String(), tc.expectedErr)
			}
		})
	}
}
