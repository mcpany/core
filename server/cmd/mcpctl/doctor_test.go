// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/health"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoctorRunner_Run_HappyPath(t *testing.T) {
	// 1. Setup Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/api/v1/doctor" {
			w.WriteHeader(http.StatusOK)
			report := health.DoctorReport{
				Status: "healthy",
				Checks: map[string]health.CheckResult{
					"database": {Status: "ok", Latency: "1ms"},
					"redis":    {Status: "ok", Latency: "2ms"},
				},
			}
			json.NewEncoder(w).Encode(report)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// 2. Setup Mock Filesystem and Config
	fs := afero.NewMemMapFs()
	// Extract port from server.URL
	_, port, _ := javaLikeSplitHostPort(server.URL)
	configContent := fmt.Sprintf(`
mcp-listen-address: ":%s"
`, port)
	err := afero.WriteFile(fs, "config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	// 3. Setup Runner
	var out bytes.Buffer
	runner := &DoctorRunner{
		Out:        &out,
		Fs:         fs,
		HTTPClient: server.Client(),
	}

	// 4. Setup Command
	cmd := &cobra.Command{}
	cmd.Flags().StringSlice("config-path", []string{"."}, "config path")
	cmd.Flags().Parse([]string{"--config-path", "."})

	// Force address via viper
	viper.Set("mcp-listen-address", ":"+port)
	defer viper.Set("mcp-listen-address", "")

	// Dummy service config
	err = fs.MkdirAll("server", 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, "server/service.yaml", []byte(`
apiVersion: mcpany.org/v1alpha1
kind: Service
metadata:
  name: test-service
spec:
  mcp:
    command: echo
    args: ["hello"]
`), 0644)
	require.NoError(t, err)

	// Run
	err = runner.Run(cmd, nil)
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "[ ] Checking Configuration... OK")
	assert.Contains(t, output, fmt.Sprintf("[ ] Checking Server Connectivity (http://localhost:%s)... OK", port))
	assert.Contains(t, output, "[ ] Checking System Health... OK")
	assert.Contains(t, output, "database: OK")
	assert.Contains(t, output, "redis: OK")
}

func TestDoctorRunner_Run_ServerDown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close() // Close immediately

	_, port, _ := javaLikeSplitHostPort(server.URL)
	viper.Set("mcp-listen-address", ":"+port)
	defer viper.Set("mcp-listen-address", "")

	var out bytes.Buffer
	runner := &DoctorRunner{
		Out:        &out,
		Fs:         afero.NewMemMapFs(),
		HTTPClient: server.Client(),
	}

	cmd := &cobra.Command{}
	err := runner.Run(cmd, nil)
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Checking Server Connectivity")
	assert.Contains(t, output, "FAILED")
	assert.Contains(t, output, "Could not connect to server")
}

func TestDoctorRunner_Run_DoctorEndpointFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/api/v1/doctor" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	_, port, _ := javaLikeSplitHostPort(server.URL)
	viper.Set("mcp-listen-address", ":"+port)
	defer viper.Set("mcp-listen-address", "")

	var out bytes.Buffer
	runner := &DoctorRunner{
		Out:        &out,
		Fs:         afero.NewMemMapFs(),
		HTTPClient: server.Client(),
	}

	cmd := &cobra.Command{}
	err := runner.Run(cmd, nil)
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Checking Server Connectivity")
	assert.Contains(t, output, "OK")
	assert.Contains(t, output, "Checking System Health")
	assert.Contains(t, output, "WARNING")
	assert.Contains(t, output, "Doctor endpoint returned status: 500 Internal Server Error")
}

func TestDoctorRunner_Run_DoctorDegraded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/api/v1/doctor" {
			w.WriteHeader(http.StatusOK)
			report := health.DoctorReport{
				Status: "degraded",
				Checks: map[string]health.CheckResult{
					"database": {Status: "fail", Message: "connection refused"},
				},
			}
			json.NewEncoder(w).Encode(report)
			return
		}
	}))
	defer server.Close()

	_, port, _ := javaLikeSplitHostPort(server.URL)
	viper.Set("mcp-listen-address", ":"+port)
	defer viper.Set("mcp-listen-address", "")

	var out bytes.Buffer
	runner := &DoctorRunner{
		Out:        &out,
		Fs:         afero.NewMemMapFs(),
		HTTPClient: server.Client(),
	}

	cmd := &cobra.Command{}
	err := runner.Run(cmd, nil)
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "DEGRADED")
	assert.Contains(t, output, "database: FAIL")
	assert.Contains(t, output, "connection refused")
}

// Helper to handle httptest.Server URL format which is http://127.0.0.1:PORT
func javaLikeSplitHostPort(urlStr string) (string, string, error) {
	// Strip scheme
	if len(urlStr) > 7 && urlStr[:7] == "http://" {
		urlStr = urlStr[7:]
	} else if len(urlStr) > 8 && urlStr[:8] == "https://" {
		urlStr = urlStr[8:]
	}
	// Using net.SplitHostPort
	host, port, err := net.SplitHostPort(urlStr)
	if err != nil {
		return "", "", err
	}
	return host, port, nil
}
