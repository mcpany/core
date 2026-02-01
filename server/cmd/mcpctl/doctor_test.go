// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/health"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoctor_HappyPath(t *testing.T) {
	// 1. Setup Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/doctor" {
			w.WriteHeader(http.StatusOK)
			report := health.DoctorReport{
				Status: "healthy",
				Checks: map[string]health.CheckResult{
					"database": {Status: "ok", Latency: "1ms"},
				},
			}
			json.NewEncoder(w).Encode(report)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// 2. Setup Config
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	validConfig := `
upstream_services:
  - name: "test-service"
    http_service:
      address: "http://example.com"
`
	err := os.WriteFile(configFile, []byte(validConfig), 0644)
	require.NoError(t, err)

	// 3. Setup Env
	_, port, err := net.SplitHostPort(server.Listener.Addr().String())
	require.NoError(t, err)

	// We point the CLI to the mock server using the env var which settings.go respects
	t.Setenv("MCPANY_MCP_LISTEN_ADDRESS", "localhost:"+port)

	// 4. Run Command
	cmd := newRootCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&outBuf)

	// Pass the config path flag
	cmd.SetArgs([]string{"doctor", "--config-path", configFile})

	err = cmd.Execute()
	require.NoError(t, err)

	output := outBuf.String()
	assert.Contains(t, output, "[ ] Checking Configuration... OK")
	assert.Contains(t, output, "[ ] Checking Server Connectivity")
	assert.Contains(t, output, "OK")
	assert.Contains(t, output, "[ ] Checking System Health... OK")
	assert.Contains(t, output, "- database: OK")

	// Clean up viper to avoid side effects
	viper.Reset()
}

func TestDoctor_ConfigSyntaxError(t *testing.T) {
	// Setup invalid YAML syntax
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	err := os.WriteFile(configFile, []byte("invalid: yaml: content: ["), 0644)
	require.NoError(t, err)

	cmd := newRootCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&outBuf)
	cmd.SetArgs([]string{"doctor", "--config-path", configFile})

	err = cmd.Execute()
	require.NoError(t, err) // Doctor continues even with load error
	output := outBuf.String()
	assert.Contains(t, output, "WARNING")
	assert.Contains(t, output, "Failed to load services")

	viper.Reset()
}

func TestDoctor_ConfigValidationError(t *testing.T) {
	// Setup valid YAML but invalid config (invalid URL)
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	invalidConfig := `
upstream_services:
  - name: "test-service"
    http_service:
      address: "::invalid::"
`
	err := os.WriteFile(configFile, []byte(invalidConfig), 0644)
	require.NoError(t, err)

	cmd := newRootCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&outBuf)
	cmd.SetArgs([]string{"doctor", "--config-path", configFile})

	err = cmd.Execute()
	require.NoError(t, err)

	output := outBuf.String()
	// config.LoadServices performs validation and returns error if invalid.
	// doctor.go prints "WARNING" and the error message.
	assert.Contains(t, output, "WARNING")
	assert.Contains(t, output, "Failed to load services")
	assert.Contains(t, output, "Configuration Validation Failed")

	viper.Reset()
}
