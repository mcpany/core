// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/app"
	"github.com/mcpany/core/server/pkg/appconsts"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Allow loopback and private network resources for tests, as many tests
	// use localhost which might resolve to loopback IPs like ::1.
	os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", util.TrueStr)
}

// mockRunner is a mock implementation of the app.Runner interface for testing.
type mockRunner struct {
	called                   bool
	capturedStdio            bool
	capturedMcpListenAddress string
	capturedGrpcPort         string
	capturedConfigPaths      []string
	capturedShutdownTimeout  time.Duration
}

func (m *mockRunner) Run(opts app.RunOptions) error {
	m.called = true
	m.capturedStdio = opts.Stdio
	m.capturedMcpListenAddress = opts.JSONRPCPort
	m.capturedGrpcPort = opts.GRPCPort
	m.capturedConfigPaths = opts.ConfigPaths
	m.capturedShutdownTimeout = opts.ShutdownTimeout
	return nil
}

func (m *mockRunner) ReloadConfig(_ context.Context, _ afero.Fs, _ []string) error {
	return nil
}

func (m *mockRunner) RunHealthServer(_ string) error {
	return nil
}

func TestHealthCmd(t *testing.T) {
	viper.Reset()
	// Start a mock HTTP server on a random port using httptest
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	port := fmt.Sprintf("%d", ts.Listener.Addr().(*net.TCPAddr).Port)

	// Wait for the server to start (httptest starts immediately, but sleep doesn't hurt)
	time.Sleep(100 * time.Millisecond)

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"health", "--mcp-listen-address", port})
	err := rootCmd.Execute()

	assert.NoError(t, err, "Health check should pass when server is running on custom port")
}

func TestHealthCmdWithCustomPort(t *testing.T) {
	viper.Reset()
	// Start a mock HTTP server on a random port using httptest
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	port := fmt.Sprintf("%d", ts.Listener.Addr().(*net.TCPAddr).Port)

	// Wait for the server to start
	time.Sleep(100 * time.Millisecond)

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"health", "--mcp-listen-address", port})
	err := rootCmd.Execute()

	assert.NoError(t, err, "Health check should pass when server is running on custom port")
}

func TestRootCmd(t *testing.T) {
	viper.Reset()
	mock := &mockRunner{}
	originalRunner := appRunner
	appRunner = mock
	defer func() { appRunner = originalRunner }()

	// Create temp config files/dirs that actually exist, otherwise fsnotify watcher fails
	tmpDir := t.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "config*.yaml")
	assert.NoError(t, err)
	tmpFilePath := tmpFile.Name()
	_ = tmpFile.Close()

	port := findFreePort(t)
	grpcPort := findFreePort(t)

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{
		"run",
		"--stdio",
		"--mcp-listen-address", fmt.Sprintf("127.0.0.1:%d", port),
		"--grpc-port", fmt.Sprintf("%d", grpcPort),
		"--config-path", fmt.Sprintf("%s,%s", tmpFilePath, tmpDir),
		"--shutdown-timeout", "10s",
	})
	err = rootCmd.Execute()
	assert.NoError(t, err)

	assert.True(t, mock.called, "app.Run should have been called")
	assert.True(t, mock.capturedStdio, "stdio flag should be true")
	assert.Equal(t, fmt.Sprintf("127.0.0.1:%d", port), mock.capturedMcpListenAddress, "mcp-listen-address should be captured")
	assert.Equal(t, fmt.Sprintf("%d", grpcPort), mock.capturedGrpcPort, "grpc-port should be captured")
	assert.Equal(t, []string{tmpFilePath, tmpDir}, mock.capturedConfigPaths, "config-path should be captured")
	assert.Equal(t, 10*time.Second, mock.capturedShutdownTimeout, "shutdown-timeout should be captured")
}

func TestVersionCmd(t *testing.T) {
	viper.Reset()
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"version"})

	var outBuf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&outBuf, r)
		close(done)
	}()

	err := rootCmd.Execute()
	assert.NoError(t, err)

	_ = w.Close()
	<-done
	os.Stdout = originalStdout
	out := outBuf.Bytes()

	expectedVersion := appconsts.Version
	if expectedVersion == "" {
		expectedVersion = "dev"
	}
	expectedOutput := appconsts.Name + " version " + expectedVersion + "\n"
	assert.Equal(t, expectedOutput, string(out))
}

// This test is for the main function, which is not easily testable.
// We can, however, test the command execution.
func TestMainExecution(t *testing.T) {
	viper.Reset()
	// This is a bit of a meta-test. We're just making sure that calling main()
	// doesn't panic. We can't really inspect the output without more refactoring.
	// We will rely on the other tests to validate the behavior of the commands.
	assert.NotPanics(t, func() {
		// We can't actually run main because it will block.
		// Instead, we test the command directly.
		cmd := newRootCmd()
		cmd.SetArgs([]string{"--help"})
		err := cmd.Execute()
		assert.NoError(t, err)
	})
}

func TestHealthCmdFlagPrecedence(t *testing.T) {
	viper.Reset()
	// Start a mock HTTP server on a random port using httptest
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	port := fmt.Sprintf("%d", ts.Listener.Addr().(*net.TCPAddr).Port)

	// Wait for the server to start
	time.Sleep(100 * time.Millisecond)

	// Create a temporary config file with a different port
	dir, err := os.MkdirTemp("", "test-config")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(dir) }()

	configFile := dir + "/config.yaml"
	err = os.WriteFile(configFile, []byte(`
global_settings:
  bind_address: "127.0.0.1:9090"
`), 0o600)
	assert.NoError(t, err)

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"health", "--config-path", configFile, "--mcp-listen-address", port})
	err = rootCmd.Execute()

	assert.NoError(t, err, "Health check should pass because the --mcp-listen-address flag should take precedence over the config file")
}

func findFreePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func TestGracefulShutdown(t *testing.T) {
	if os.Getenv("GO_TEST_GRACEFUL_SHUTDOWN") == "1" {
		// Create a temporary log file to capture output and find the port
		logFile, err := os.CreateTemp("", "mcpany-shutdown-test-*.log")
		assert.NoError(t, err)
		defer os.Remove(logFile.Name())
		logFileName := logFile.Name()
		_ = logFile.Close()

		cmd := newRootCmd()
		// Use port 0 to let the OS pick a free port
		cmd.SetArgs([]string{
			"run",
			"--mcp-listen-address", "127.0.0.1:0",
			"--metrics-listen-address", "127.0.0.1:0",
			"--config-path", "",
			"--logfile", logFileName,
		})
		done := make(chan struct{})
		go func() {
			err := cmd.Execute()
			assert.NoError(t, err)
			close(done)
		}()

		var realPort int
		// Wait for the server to start by parsing the log file for the port
		assert.Eventually(t, func() bool {
			content, err := os.ReadFile(logFileName)
			if err != nil {
				return false
			}
			// Scan line by line to find the listener log
			scanner := bufio.NewScanner(bytes.NewReader(content))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "HTTP server listening") {
					// Regex to match port regardless of position
					// port=[^:]+:(\d+) matches port=IP:PORT
					re := regexp.MustCompile(`port=[^:]+:(\d+)`)
					matches := re.FindStringSubmatch(line)
					if len(matches) > 1 {
						fmt.Sscanf(matches[1], "%d", &realPort)
						return true
					}
				}
			}
			return false
		}, 15*time.Second, 100*time.Millisecond, "Failed to find port in logs")

		if realPort == 0 {
			return
		}

		// Wait for the server to be healthy
		assert.Eventually(t, func() bool {
			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/healthz", realPort))
			if err != nil {
				return false
			}
			defer func() { _ = resp.Body.Close() }()
			return resp.StatusCode == http.StatusOK
		}, 10*time.Second, 100*time.Millisecond)

		// Signal parent that we are ready
		fmt.Println("READY")

		// Wait for server to finish
		<-done
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestGracefulShutdown$") //nolint:gosec
	cmd.Env = append(os.Environ(), "GO_TEST_GRACEFUL_SHUTDOWN=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	stdout, err := cmd.StdoutPipe()
	assert.NoError(t, err)

	err = cmd.Start()
	assert.NoError(t, err)

	// Wait for "READY" signal from child
	scanner := bufio.NewScanner(stdout)
	ready := false
	timer := time.NewTimer(60 * time.Second)
	done := make(chan struct{})

	go func() {
		for scanner.Scan() {
			if scanner.Text() == "READY" {
				close(done)
				return
			}
		}
	}()

	select {
	case <-done:
		ready = true
	case <-timer.C:
		t.Log("Timed out waiting for child process to be ready")
	}
	timer.Stop()

	if !ready {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		t.Fatal("Child process did not become ready in time")
	}

	// Send the interrupt signal to the child process group.
	err = syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
	assert.NoError(t, err)

	err = cmd.Wait()
	assert.NoError(t, err)
}

func TestConfigValidateCmd(t *testing.T) {
	viper.Reset()
	port := findFreePort(t)
	// Create a temporary valid config file
	validConfigFile, err := os.CreateTemp("", "valid-config-*.yaml")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(validConfigFile.Name()) }()
	_, err = validConfigFile.WriteString(fmt.Sprintf(`
global_settings:
  mcp_listen_address: "127.0.0.1:%d"
`, port))
	assert.NoError(t, err)
	_ = validConfigFile.Close()

	// Create a temporary invalid config file
	invalidConfigFile, err := os.CreateTemp("", "invalid-config-*.yaml")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(invalidConfigFile.Name()) }()
	_, err = invalidConfigFile.WriteString(`
invalid-yaml
`)
	assert.NoError(t, err)
	_ = invalidConfigFile.Close()

	invalidLogicConfigFile, err := os.CreateTemp("", "invalid-logic-config-*.yaml")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(invalidLogicConfigFile.Name()) }()
	_, err = invalidLogicConfigFile.WriteString(`
upstream_services:
  - name: "my-service"
    http_service:
      address: "::invalid::"

`)
	assert.NoError(t, err)
	_ = invalidLogicConfigFile.Close()

	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectedOutput string
	}{
		{
			name:           "valid config",
			args:           []string{"config", "validate", "--config-path", validConfigFile.Name()},
			expectError:    false,
			expectedOutput: "Configuration is valid.",
		},
		{
			name:        "invalid config",
			args:        []string{"config", "validate", "--config-path", invalidConfigFile.Name()},
			expectError: true,
		},
		{
			name:        "invalid logic config",
			args:        []string{"config", "validate", "--config-path", invalidLogicConfigFile.Name()},
			expectError: true,
		},
		{
			name:           "no config file",
			args:           []string{"config", "validate"},
			expectError:    false,
			expectedOutput: "Configuration is valid.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			rootCmd := newRootCmd()
			rootCmd.SetArgs(tt.args)

			var outBuf bytes.Buffer
			done := make(chan struct{})
			go func() {
				_, _ = io.Copy(&outBuf, r)
				close(done)
			}()

			err := rootCmd.Execute()

			_ = w.Close()
			<-done
			os.Stdout = originalStdout
			out := outBuf.Bytes()

			if tt.expectError {
				if err == nil {
					t.Logf("Expected error for case %q but got nil. Output: %s", tt.name, string(out))
				}
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, string(out), tt.expectedOutput)
			}
		})
	}
}

func TestConfigGenerateCmd(t *testing.T) {
	viper.Reset()
	originalStdout := os.Stdout
	originalStdin := os.Stdin
	defer func() {
		os.Stdout = originalStdout
		os.Stdin = originalStdin
	}()

	rOut, wOut, _ := os.Pipe()
	rIn, wIn, _ := os.Pipe()
	os.Stdout = wOut
	os.Stdin = rIn

	// Inputs for HTTP service generation
	inputs := []string{
		"http",
		"my-service",
		"http://127.0.0.1:8080", // Keep this one as it's just input data, not actual binding

		"get-data",
		"Get some data",
		"HTTP_METHOD_GET",
		"/data",
	}

	go func() {
		defer func() { _ = wIn.Close() }()
		for _, input := range inputs {
			_, _ = fmt.Fprintln(wIn, input)
		}
	}()

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"config", "generate"})

	var outBuf bytes.Buffer
	outDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(&outBuf, rOut)
		close(outDone)
	}()

	err := rootCmd.Execute()

	_ = wOut.Close()
	<-outDone
	out := outBuf.Bytes()

	assert.NoError(t, err)
	output := string(out)
	assert.Contains(t, output, "MCP Any CLI: Configuration Generator")
	assert.Contains(t, output, "Generated configuration:")
	assert.Contains(t, output, "upstreamServices:")
	assert.Contains(t, output, "name: \"my-service\"")
}

func TestDocCmd(t *testing.T) {
	viper.Reset()
	// Create a temporary valid config file
	// Create a temporary valid config file
	configFile, err := os.CreateTemp("", "doc-config-*.yaml")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(configFile.Name()) }()

	// Use a free port to avoid conflicts if health check is attempted
	port := findFreePort(t)
	_, err = configFile.WriteString(fmt.Sprintf(`
upstream_services:
  - name: "my-service"
    http_service:
      address: "http://127.0.0.1:%d"
      tools:
        - name: "my-tool"
          call_id: "my-call"
      calls:
        my-call:
          id: "my-call"
          method: HTTP_METHOD_GET
          endpoint_path: "/test"
`, port))
	assert.NoError(t, err)
	_ = configFile.Close()

	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"config", "doc", "--config-path", configFile.Name()})

	var outBuf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&outBuf, r)
		close(done)
	}()

	err = rootCmd.Execute()

	_ = w.Close()
	<-done
	os.Stdout = originalStdout
	out := outBuf.Bytes()

	assert.NoError(t, err)
	output := string(out)
	assert.Contains(t, output, "# Available Tools")
	assert.Contains(t, output, "my-service.my-tool")
}

func TestUpdateCmd(t *testing.T) {
	viper.Reset()
	// Mock GitHub API
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/mcpany/core/releases/latest" {
			w.Header().Set("Content-Type", "application/json")
			// Return a release with same version as current
			fmt.Fprintf(w, `{"tag_name": "%s"}`, Version)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	t.Setenv("GITHUB_API_URL", ts.URL)

	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"update"})

	var outBuf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&outBuf, r)
		close(done)
	}()

	err := rootCmd.Execute()

	_ = w.Close()
	<-done
	os.Stdout = originalStdout
	out := outBuf.Bytes()

	assert.NoError(t, err)
	assert.Contains(t, string(out), "You are already running the latest version.")
}

func TestConfigSchemaCmd(t *testing.T) {
	viper.Reset()

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"config", "schema"})

	buf := new(strings.Builder)
	rootCmd.SetOut(buf)

	err := rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "$schema")
	assert.Contains(t, output, "mcpany.config.v1.McpAnyServerConfig")
}

func TestConfigCheckCmd(t *testing.T) {
	viper.Reset()
	// Create a temporary valid config file
	validConfigFile, err := os.CreateTemp("", "valid-config-*.yaml")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(validConfigFile.Name()) }()
	_, err = validConfigFile.WriteString(`
global_settings:
  mcp_listen_address: "127.0.0.1:8080"
`)
	assert.NoError(t, err)
	_ = validConfigFile.Close()

	// Create a temporary invalid schema config file
	invalidConfigFile, err := os.CreateTemp("", "invalid-schema-config-*.yaml")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(invalidConfigFile.Name()) }()
	_, err = invalidConfigFile.WriteString(`
global_settings:
  mcp_listen_address: 12345
`)
	assert.NoError(t, err)
	_ = invalidConfigFile.Close()

	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectedOutput string
	}{
		{
			name:           "valid config",
			args:           []string{"config", "check", validConfigFile.Name()},
			expectError:    false,
			expectedOutput: "Configuration schema is valid.",
		},
		{
			name:        "invalid config",
			args:        []string{"config", "check", invalidConfigFile.Name()},
			expectError: true,
		},
		{
			name:        "file not found",
			args:        []string{"config", "check", "nonexistent.yaml"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			rootCmd := newRootCmd()
			rootCmd.SetArgs(tt.args)

			var outBuf bytes.Buffer
			done := make(chan struct{})
			go func() {
				_, _ = io.Copy(&outBuf, r)
				close(done)
			}()

			err := rootCmd.Execute()

			_ = w.Close()
			<-done
			os.Stdout = originalStdout
			out := outBuf.Bytes()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, string(out), tt.expectedOutput)
			}
		})
	}
}
