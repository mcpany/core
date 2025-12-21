// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package integration contains integration tests and helpers
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	apiv1 "github.com/mcpany/core/proto/api/v1"
	bus "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/mcpany/core/pkg/app"
	"github.com/spf13/afero"
)

// CreateTempConfigFile creates a temporary configuration file for the configured upstream service.
func CreateTempConfigFile(t *testing.T, config *configv1.UpstreamServiceConfig) string {
	t.Helper()

	// Build the configuration
	mcpanyConfig := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{config},
	}.Build()

	// Use protojson to ensure correct field names and handle opaque structs
	data, err := protojson.Marshal(mcpanyConfig)
	require.NoError(t, err)

	tempFile, err := os.CreateTemp(t.TempDir(), "mcpany-config-*.json")
	require.NoError(t, err)

	_, err = tempFile.Write(data)
	require.NoError(t, err)
	require.NoError(t, tempFile.Close())

	return tempFile.Name()
}

// CreateTempNatsConfigFile creates a temporary configuration file for NATS.
func CreateTempNatsConfigFile(t *testing.T) string {
	t.Helper()

	natsURL := "${NATS_URL}"
	// Build the configuration
	mcpanyConfig := configv1.McpAnyServerConfig_builder{
		GlobalSettings: configv1.GlobalSettings_builder{
			MessageBus: bus.MessageBus_builder{
				Nats: bus.NatsBus_builder{
					ServerUrl: &natsURL,
				}.Build(),
			}.Build(),
		}.Build(),
	}.Build()

	// Use protojson to ensure correct field names (snake_case/camelCase as expected by proto)
	data, err := protojson.Marshal(mcpanyConfig)
	require.NoError(t, err)

	tmpFile, err := os.CreateTemp(t.TempDir(), "nats-config-*.json")
	require.NoError(t, err)
	_, err = tmpFile.Write(data)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)
	return tmpFile.Name()
}

// A thread-safe buffer for capturing process output concurrently.
type threadSafeBuffer struct {
	b  bytes.Buffer
	mu sync.RWMutex
}

// Write appends the contents of p to the buffer, growing the buffer as needed.
func (b *threadSafeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}

// String returns the contents of the unread portion of the buffer
// as a string. If the Buffer is a nil pointer, it returns "<nil>".
func (b *threadSafeBuffer) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.b.String()
}

// ProjectRoot returns the absolute path to the project root.
func ProjectRoot(t *testing.T) string {
	t.Helper()
	root, err := GetProjectRoot()
	require.NoError(t, err)
	return root
}

const (
	// McpAnyServerStartupTimeout is the timeout for the server to start.
	McpAnyServerStartupTimeout = 120 * time.Second
	// ServiceStartupTimeout is the timeout for services to start up.
	ServiceStartupTimeout = 60 * time.Second
	// TestWaitTimeShort is a short wait time for tests.
	TestWaitTimeShort = 120 * time.Second
	// TestWaitTimeMedium is the default timeout for medium duration tests.
	TestWaitTimeMedium = 240 * time.Second
	// TestWaitTimeLong is the default timeout for long duration tests.
	TestWaitTimeLong = 5 * time.Minute
	// RetryInterval is the interval between retries.
	RetryInterval           = 250 * time.Millisecond
	localHeaderMcpSessionID = "Mcp-Session-Id"
	dockerCmd               = "docker"
	sudoCmd                 = "sudo"
)

var (
	dockerCommand string
	dockerArgs    []string
	dockerOnce    sync.Once
)

// getDockerCommand returns the command and base arguments for running Docker,
// checking for direct access, then trying passwordless sudo. The result is
// cached for subsequent calls.
func getDockerCommand() (string, []string) {
	dockerOnce.Do(func() {
		// Environment variable overrides detection.
		if val := os.Getenv("USE_SUDO_FOR_DOCKER"); val == "true" || val == "1" {
			dockerCommand = sudoCmd
			dockerArgs = []string{dockerCmd}
			return
		}

		// First, try running docker directly.
		if _, err := exec.LookPath("docker"); err == nil {
			cmd := exec.Command(dockerCmd, "info")
			if err := cmd.Run(); err == nil {
				dockerCommand = dockerCmd
				dockerArgs = []string{}
				return
			}
		}

		// If direct access fails, check for passwordless sudo.
		if _, err := exec.LookPath(sudoCmd); err == nil {
			cmd := exec.Command(sudoCmd, "-n", dockerCmd, "info")
			if err := cmd.Run(); err == nil {
				dockerCommand = sudoCmd
				dockerArgs = []string{dockerCmd}
				return
			}
		}

		// Fallback to plain docker if all else fails.
		dockerCommand = dockerCmd
		dockerArgs = []string{}
	})
	return dockerCommand, dockerArgs
}

// --- Binary Paths ---

var (
	projectRoot  string
	findRootOnce sync.Once
)

// GetProjectRoot returns the absolute path to the project root.
func GetProjectRoot() (string, error) {
	var err error
	findRootOnce.Do(func() {
		// Find the project root by looking for the go.mod file
		var dir string
		dir, err = os.Getwd()
		if err != nil {
			return
		}
		for {
			if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
				projectRoot = dir
				return
			}
			if dir == filepath.Dir(dir) {
				err = fmt.Errorf("go.mod not found")
				return
			}
			dir = filepath.Dir(dir)
		}
	})
	if err != nil {
		return "", err
	}
	return filepath.Abs(projectRoot)
}

// --- Helper: Find Free Port ---
var portMutex sync.Mutex

// FindFreePort finds a free TCP port on localhost.
func FindFreePort(t *testing.T) int {
	portMutex.Lock()
	defer portMutex.Unlock()
	t.Helper()
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	l, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err)
	defer func() {
		err := l.Close()
		if err != nil {
			t.Logf("Error closing listener for free port check: %v", err)
		}
	}()
	return l.Addr().(*net.TCPAddr).Port
}

// ManagedProcess represents an external process managed by the test framework.
// --- Process Management for External Services ---
// ManagedProcess manages an external process for testing.
type ManagedProcess struct {
	cmd                 *exec.Cmd
	t                   *testing.T
	wg                  sync.WaitGroup
	stdout              threadSafeBuffer
	stderr              threadSafeBuffer
	waitDone            chan struct{}
	label               string
	IgnoreExitStatusOne bool
	Port                int
	Dir                 string
}

// NewManagedProcess creates a new ManagedProcess instance.
func NewManagedProcess(t *testing.T, label, command string, args []string, env []string) *ManagedProcess {
	t.Helper()
	cmd := exec.Command(command, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}
	if command == "sudo" && len(args) > 0 && args[0] == "docker" {
		cmd.Stdin = os.Stdin
	}

	mp := &ManagedProcess{
		cmd:      cmd,
		t:        t,
		label:    label,
		waitDone: make(chan struct{}),
	}
	cmd.Stdout = &mp.stdout
	cmd.Stderr = &mp.stderr
	return mp
}

// Cmd returns the underlying exec.Cmd.
func (mp *ManagedProcess) Cmd() *exec.Cmd {
	return mp.cmd
}

// Start starts the process.
func (mp *ManagedProcess) Start() error {
	if mp.Dir != "" {
		mp.cmd.Dir = mp.Dir
	}
	mp.t.Logf("[%s] Starting process: %s %v", mp.label, mp.cmd.Path, mp.cmd.Args)
	if err := mp.cmd.Start(); err != nil {
		// Give a tiny moment for stderr to populate on immediate start failure.
		time.Sleep(50 * time.Millisecond)
		return fmt.Errorf("[%s] failed to start process %s: %w. Stderr: %s, Stdout: %s", mp.label, mp.cmd.Path, err, mp.stderr.String(), mp.stdout.String())
	}
	mp.wg.Add(1)
	go func() {
		defer mp.wg.Done()
		err := mp.cmd.Wait()
		close(mp.waitDone)
		// Log output regardless of error, can be useful for debugging successful exits too
		mp.t.Logf("[%s] Process %s finished. Stdout:\n%s\nStderr:\n%s", mp.label, mp.cmd.Path, mp.stdout.String(), mp.stderr.String())
		if err != nil {
			errStr := err.Error()
			switch {
			case mp.IgnoreExitStatusOne && errStr == "exit status 1":
				mp.t.Logf("[%s] Process %s exited with status 1, which is being ignored as requested.", mp.label, mp.cmd.Path)
			case !strings.Contains(errStr, "killed") && !strings.Contains(errStr, "signal: interrupt") && !strings.Contains(errStr, "exit status -1"):
				mp.t.Logf("[%s] Process %s exited with unexpected error: %v", mp.label, mp.cmd.Path, err)
			default:
				mp.t.Logf("[%s] Process %s exited as expected (killed/interrupted).", mp.label, mp.cmd.Path)
			}
		} else {
			mp.t.Logf("[%s] Process %s exited successfully.", mp.label, mp.cmd.Path)
		}
	}()
	return nil
}

// Allow patching for testing
var syscallKill = syscall.Kill

// Stop stops the process, attempting graceful shutdown then force kill.
func (mp *ManagedProcess) Stop() {
	select {
	case <-mp.waitDone:
		mp.t.Logf("[%s] Process %s already exited.", mp.label, mp.cmd.Path)
		mp.wg.Wait() // ensure Wait goroutine has fully finished
		return
	default:
		// Not exited yet, proceed to stop it.
	}

	if mp.cmd == nil || mp.cmd.Process == nil {
		mp.t.Logf("[%s] Process %s not running or already stopped.", mp.label, mp.cmd.Path)
		mp.wg.Wait() // ensure Wait goroutine finishes if process exited itself
		return
	}
	mp.t.Logf("[%s] Stopping process: %s (PID: %d)", mp.label, mp.cmd.Path, mp.cmd.Process.Pid)

	pgid, err := syscall.Getpgid(mp.cmd.Process.Pid)
	sentSignal := false
	if err == nil {
		// Try to kill the whole process group
		if errSignal := syscallKill(-pgid, syscall.SIGINT); errSignal == nil {
			sentSignal = true
			mp.t.Logf("[%s] Sent SIGINT to process group %d for %s.", mp.label, pgid, mp.cmd.Path)
		} else {
			mp.t.Logf("[%s] Failed to send SIGINT to process group %d for %s: %v. Will try single process.", mp.label, pgid, mp.cmd.Path, errSignal)
		}
	} else {
		mp.t.Logf("[%s] Failed to get PGID for %s (PID: %d): %v. Attempting SIGINT to single process.", mp.label, mp.cmd.Path, mp.cmd.Process.Pid, err)
	}

	// Fallback to single process kill if group kill failed or wasn't attempted
	if !sentSignal {
		if errKill := mp.cmd.Process.Signal(syscall.SIGINT); errKill == nil {
			sentSignal = true
			mp.t.Logf("[%s] Sent SIGINT to single process %s (PID: %d).", mp.label, mp.cmd.Path, mp.cmd.Process.Pid)
		} else {
			mp.t.Logf("[%s] Failed to send SIGINT to single process %s (PID: %d): %v. Will try SIGKILL.", mp.label, mp.cmd.Path, mp.cmd.Process.Pid, errKill)
		}
	}

	done := make(chan struct{})
	go func() {
		mp.wg.Wait()
		close(done)
	}()

	if sentSignal {
		select {
		case <-done:
			mp.t.Logf("[%s] Process %s stopped gracefully after SIGINT.", mp.label, mp.cmd.Path)
			return
		case <-time.After(15 * time.Second):
			mp.t.Logf("[%s] Process %s did not stop gracefully after SIGINT and 15s, attempting SIGKILL.", mp.label, mp.cmd.Path)
		}
	}

	// Force kill if not stopped or SIGINT failed
	if pgid != 0 {
		if errKillGroup := syscallKill(-pgid, syscall.SIGKILL); errKillGroup != nil {
			mp.t.Logf("[%s] Failed to send SIGKILL to process group %d for %s: %v. Will try single process.", mp.label, pgid, mp.cmd.Path, errKillGroup)
			if mp.cmd.Process != nil {
				if errKillHard := mp.cmd.Process.Kill(); errKillHard != nil {
					mp.t.Logf("[%s] Failed to send SIGKILL to single process %s (PID: %d): %v", mp.label, mp.cmd.Path, mp.cmd.Process.Pid, errKillHard)
				}
			}
		}
	} else if mp.cmd.Process != nil {
		if errKillHard := mp.cmd.Process.Kill(); errKillHard != nil {
			mp.t.Logf("[%s] Failed to send SIGKILL to single process %s (PID: %d): %v", mp.label, mp.cmd.Path, mp.cmd.Process.Pid, errKillHard)
		}
	}
	<-done
	mp.t.Logf("[%s] Process %s stopped (SIGKILL or already exited).", mp.label, mp.cmd.Path)
}

// StdoutString returns the captured stdout as a string.
func (mp *ManagedProcess) StdoutString() string { return mp.stdout.String() }

// StderrString returns the captured stderr as a string.
func (mp *ManagedProcess) StderrString() string { return mp.stderr.String() }

// WaitForText waits for specific text to appear in the process's stdout.
func (mp *ManagedProcess) WaitForText(t *testing.T, text string, timeout time.Duration) {
	t.Helper()
	require.Eventually(t, func() bool {
		return strings.Contains(mp.StdoutString(), text)
	}, timeout, RetryInterval, "Text '%s' not found in stdout for process '%s' in time.\nStdout: %s\nStderr: %s", text, mp.label, mp.StdoutString(), mp.StderrString())
}

// WaitForTCPPort waits for a TCP port to become open and accepting connections.
func WaitForTCPPort(t *testing.T, port int, timeout time.Duration) {
	t.Helper()
	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 100*time.Millisecond)
		if err != nil {
			return false // Port is not open yet
		}
		_ = conn.Close()
		return true // Port is open
	}, timeout, 250*time.Millisecond, "Port %d did not become available in time", port)
}

// WaitForGRPCReady waits for a gRPC server to become ready by attempting to connect.
func WaitForGRPCReady(t *testing.T, grpcAddress string, timeout time.Duration) {
	t.Helper()
	require.Eventually(t, func() bool {
		// This context is for a single connection attempt.
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		// grpc.NewClient is non-blocking.
		conn, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Logf("gRPC server at %s not ready, error creating client: %v", grpcAddress, err)
			return false
		}
		defer func() { _ = conn.Close() }()

		// Wait for the connection to be ready.
		for {
			s := conn.GetState()
			if s == connectivity.Ready {
				return true
			}
			if !conn.WaitForStateChange(ctx, s) {
				// Context expired, so this attempt failed.
				return false
			}
		}
	}, timeout, RetryInterval, "gRPC server at %s did not become ready in time", grpcAddress)
}

// WaitForWebsocketReady waits for a websocket server to become ready by attempting to connect.
func WaitForWebsocketReady(t *testing.T, url string, timeout time.Duration) {
	t.Helper()
	require.Eventually(t, func() bool {
		dialer := websocket.Dialer{
			HandshakeTimeout: 2 * time.Second,
		}
		conn, resp, err := dialer.Dial(url, nil)
		if err != nil {
			t.Logf("Websocket server at %s not ready: %v", url, err)
			return false
		}
		defer func() { _ = resp.Body.Close() }()
		_ = conn.Close()
		return true
	}, timeout, RetryInterval, "Websocket server at %s did not become ready in time", url)
}

// WaitForHTTPHealth waits for an HTTP endpoint to return a 200 OK status.
func WaitForHTTPHealth(t *testing.T, url string, timeout time.Duration) {
	t.Helper()
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	require.Eventually(t, func() bool {
		resp, err := client.Get(url)
		if err != nil {
			return false
		}
		defer func() { _ = resp.Body.Close() }()
		return resp.StatusCode == http.StatusOK
	}, timeout, 250*time.Millisecond, "URL %s did not become healthy in time", url)
}

// IsDockerSocketAccessible checks if the Docker daemon is accessible.
func IsDockerSocketAccessible() bool {
	dockerExe, dockerArgs := getDockerCommand()

	cmd := exec.Command(dockerExe, append(dockerArgs, "info")...) //nolint:gosec // Test helper
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// --- Mock Service Start Helpers (External Processes) ---

// StartDockerContainer starts a docker container with the given image and args.
func StartDockerContainer(t *testing.T, imageName, containerName string, runArgs []string, command ...string) (cleanupFunc func()) {
	t.Helper()
	dockerExe, dockerBaseArgs := getDockerCommand()

	// buildArgs safely creates a new slice for command arguments.
	buildArgs := func(cmdArgs ...string) []string {
		// Create a new slice with enough capacity
		finalArgs := make([]string, 0, len(dockerBaseArgs)+len(cmdArgs))
		// Append the base arguments first
		finalArgs = append(finalArgs, dockerBaseArgs...)
		// Append the command-specific arguments
		finalArgs = append(finalArgs, cmdArgs...)
		return finalArgs
	}

	// Ensure the container is not already running from a previous failed run

	stopCmd := exec.Command(dockerExe, buildArgs("stop", containerName)...) //nolint:gosec // Test helper
	_ = stopCmd.Run()                                                       // Ignore error, it might not be running

	rmCmd := exec.Command(dockerExe, buildArgs("rm", containerName)...) //nolint:gosec // Test helper
	_ = rmCmd.Run()                                                     // Ignore error, it might not exist

	dockerRunArgs := []string{"run", "--name", containerName, "--rm"}
	dockerRunArgs = append(dockerRunArgs, runArgs...)
	dockerRunArgs = append(dockerRunArgs, imageName)
	dockerRunArgs = append(dockerRunArgs, command...)

	startCmd := exec.Command(dockerExe, buildArgs(dockerRunArgs...)...) //nolint:gosec // Test helper
	// Capture stderr for better error reporting
	var stderr bytes.Buffer
	startCmd.Stderr = &stderr

	// Use Run instead of Start for 'docker run -d' to ensure the command completes
	// and the container is running before proceeding.
	err := startCmd.Run()
	require.NoError(t, err, "failed to start docker container %s. Stderr: %s", imageName, stderr.String())

	cleanupFunc = func() {
		t.Logf("Stopping and removing docker container: %s", containerName)

		stopCleanupCmd := exec.Command(dockerExe, buildArgs("stop", containerName)...) //nolint:gosec // Test helper
		err := stopCleanupCmd.Run()
		if err != nil {
			// Log as error, but don't fail the test, as cleanup failure is secondary.
			t.Errorf("Failed to stop docker container %s: %v", containerName, err)
		}
	}

	// Give the container a moment to initialize. This is still a good idea even
	// after using Run, as the service inside the container needs time to start up.
	time.Sleep(3 * time.Second)

	return cleanupFunc
}

// MCPANYTestServerInfo contains information about a running MCPANY test server.
// --- MCPANY Server Helper (External Process) ---
// MCPANYTestServerInfo contains information about a running MCP Any server instance for testing.
type MCPANYTestServerInfo struct {
	Process                  *ManagedProcess
	JSONRPCEndpoint          string
	HTTPEndpoint             string
	GrpcRegistrationEndpoint string
	NatsURL                  string
	SessionID                string
	HTTPClient               *http.Client
	GRPCRegConn              *grpc.ClientConn
	RegistrationClient       apiv1.RegistrationServiceClient
	CleanupFunc              func()
	T                        *testing.T
}

// WebsocketEchoServerInfo contains information about a running Websocket echo server.
// --- Websocket Echo Server Helper ---
// WebsocketEchoServerInfo contains information about a running mock WebSocket echo server.
type WebsocketEchoServerInfo struct {
	URL         string
	CleanupFunc func()
}

// StartWebsocketEchoServer starts a mock WebSocket echo server.
func StartWebsocketEchoServer(t *testing.T) *WebsocketEchoServerInfo {
	t.Helper()

	port := FindFreePort(t)
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	upgrader := websocket.Upgrader{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("Websocket upgrade error: %v", err)
			return
		}
		defer func() { _ = c.Close() }()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				// Don't log expected closure errors
				if !websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					t.Logf("Websocket read error: %v", err)
				}
				break
			}
			err = c.WriteMessage(mt, message)
			if err != nil {
				t.Logf("Websocket write error: %v", err)
				break
			}
		}
	})

	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			t.Logf("Websocket echo server ListenAndServe error: %v", err)
		}
	}()

	// Wait for the server to be ready
	WaitForTCPPort(t, port, 5*time.Second)

	return &WebsocketEchoServerInfo{
		URL: "ws://" + addr,
		CleanupFunc: func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				t.Logf("Websocket echo server shutdown error: %v", err)
			}
		},
	}
}

// StartMCPANYServerWithConfig starts the MCP Any server with a provided config content.
func StartMCPANYServerWithConfig(t *testing.T, testName, configContent string) *MCPANYTestServerInfo {
	t.Helper()
	tmpFile, err := os.CreateTemp(t.TempDir(), "mcpany-config-*.yaml")
	require.NoError(t, err)
	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)
	return StartMCPANYServer(t, testName, "--config-path", tmpFile.Name())
}

// StartMCPANYServer starts the MCP Any server with default settings.
func StartMCPANYServer(t *testing.T, testName string, extraArgs ...string) *MCPANYTestServerInfo {
	return StartMCPANYServerWithClock(t, testName, true, extraArgs...)
}

// StartMCPANYServerWithNoHealthCheck starts the MCP Any server but skips the health check.
func StartMCPANYServerWithNoHealthCheck(t *testing.T, testName string, extraArgs ...string) *MCPANYTestServerInfo {
	return StartMCPANYServerWithClock(t, testName, false, extraArgs...)
}

// StartInProcessMCPANYServer starts an in-process MCP Any server for testing.
func StartInProcessMCPANYServer(t *testing.T, _ string) *MCPANYTestServerInfo {
	t.Helper()

	_, err := GetProjectRoot()
	require.NoError(t, err, "Failed to get project root")

	jsonrpcPort := FindFreePort(t)
	grpcRegPort := FindFreePort(t)
	for grpcRegPort == jsonrpcPort {
		grpcRegPort = FindFreePort(t)
	}

	jsonrpcAddress := fmt.Sprintf(":%d", jsonrpcPort)
	grpcRegAddress := fmt.Sprintf(":%d", grpcRegPort)

	jsonrpcEndpoint := fmt.Sprintf("http://127.0.0.1:%d", jsonrpcPort)
	grpcRegEndpoint := fmt.Sprintf("127.0.0.1:%d", grpcRegPort)
	mcpRequestURL := jsonrpcEndpoint + "/mcp"

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		appRunner := app.NewApplication()
		err := appRunner.Run(ctx, afero.NewOsFs(), false, jsonrpcAddress, grpcRegAddress, []string{}, 5*time.Second)
		require.NoError(t, err)
	}()

	WaitForHTTPHealth(t, fmt.Sprintf("http://127.0.0.1:%d/healthz", jsonrpcPort), McpAnyServerStartupTimeout)

	var grpcRegConn *grpc.ClientConn
	require.Eventually(t, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var errDial error
		grpcRegConn, errDial = grpc.NewClient(grpcRegEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if errDial != nil {
			t.Logf("MCPANY gRPC registration endpoint at %s not ready, error creating client: %v", grpcRegEndpoint, errDial)
			return false
		}
		state := grpcRegConn.GetState()
		if state == connectivity.Ready || state == connectivity.Idle {
			t.Logf("Successfully connected to MCPANY gRPC registration endpoint at %s with state %s", grpcRegEndpoint, state)
			return true
		}
		if !grpcRegConn.WaitForStateChange(ctx, state) {
			t.Logf("MCPANY gRPC registration endpoint at %s did not transition from %s", grpcRegEndpoint, state)
			_ = grpcRegConn.Close()
			return false
		}
		t.Logf("Successfully connected to MCPANY gRPC registration endpoint at %s", grpcRegEndpoint)
		return true
	}, McpAnyServerStartupTimeout, RetryInterval, "MCPANY gRPC registration endpoint at %s did not become healthy in time", grpcRegEndpoint)

	registrationClient := apiv1.NewRegistrationServiceClient(grpcRegConn)

	return &MCPANYTestServerInfo{
		JSONRPCEndpoint:          jsonrpcEndpoint,
		HTTPEndpoint:             mcpRequestURL,
		GrpcRegistrationEndpoint: grpcRegEndpoint,
		HTTPClient:               &http.Client{Timeout: 2 * time.Second},
		GRPCRegConn:              grpcRegConn,
		RegistrationClient:       registrationClient,
		CleanupFunc: func() {
			cancel()
			if grpcRegConn != nil {
				_ = grpcRegConn.Close()
			}
		},
		T: t,
	}
}

// StartNatsServer starts a NATS server for testing.
func StartNatsServer(t *testing.T) (string, func()) {
	t.Helper()

	var natsServerBin string
	// Try to find nats-server in PATH first
	pathBin, err := exec.LookPath("nats-server")
	if err == nil {
		natsServerBin = pathBin
	} else {
		// Check /tools/nats-server (Docker)
		if _, err := os.Stat("/tools/nats-server"); err == nil {
			natsServerBin = "/tools/nats-server"
		} else {
			root, err := GetProjectRoot()
			require.NoError(t, err)
			natsServerBin = filepath.Join(root, "../build/env/bin/nats-server")
			if info, err := os.Stat(natsServerBin); err == nil {
				t.Logf("DEBUG: Using nats-server binary at: %s (ModTime: %s)", natsServerBin, info.ModTime())
			} else {
				t.Logf("DEBUG: nats-server binary not found at: %s", natsServerBin)
			}
			_, err = os.Stat(natsServerBin)
			require.NoError(t, err, "nats-server binary not found at %s or /tools/nats-server. Run 'make prepare'.", natsServerBin)
		}
	}

	natsPort := FindFreePort(t)
	natsURL := fmt.Sprintf("nats://127.0.0.1:%d", natsPort)

	cmd := exec.Command(natsServerBin, "-p", fmt.Sprintf("%d", natsPort)) //nolint:gosec // Test helper
	err = cmd.Start()
	require.NoError(t, err)
	WaitForTCPPort(t, natsPort, 10*time.Second) // Wait for NATS server to be ready
	cleanup := func() {
		_ = cmd.Process.Kill()
	}
	return natsURL, cleanup
}

// StartRedisContainer starts a Redis container for testing.
func StartRedisContainer(t *testing.T) (redisAddr string, cleanupFunc func()) {
	t.Helper()
	require.True(t, IsDockerSocketAccessible(), "Docker is not running or accessible. Please start Docker to run this test.")

	containerName := fmt.Sprintf("mcpany-redis-test-%d", time.Now().UnixNano())
	redisPort := FindFreePort(t)
	redisAddr = fmt.Sprintf("127.0.0.1:%d", redisPort)

	runArgs := []string{
		"-d", // detached mode
		"-p", fmt.Sprintf("%d:6379", redisPort),
	}

	command := []string{
		"redis-server",
		"--bind", "0.0.0.0",
	}

	cleanup := StartDockerContainer(t, "mirror.gcr.io/library/redis:latest", containerName, runArgs, command...)

	// Wait for Redis to be ready
	require.Eventually(t, func() bool {
		// Use redis-cli to ping the server
		dockerExe, dockerBaseArgs := getDockerCommand()
		pingArgs := append(dockerBaseArgs, "exec", containerName, "redis-cli", "ping") //nolint:gocritic // Helper
		cmd := exec.Command(dockerExe, pingArgs...)                                    //nolint:gosec // Test helper
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("redis-cli ping failed: %v, output: %s", err, string(output))
			return false
		}
		return strings.Contains(string(output), "PONG")
	}, 15*time.Second, 500*time.Millisecond, "Redis container did not become ready in time")

	return redisAddr, cleanup
}

// StartMCPANYServerWithClock starts the MCP Any server, optionally waiting for health.
func StartMCPANYServerWithClock(t *testing.T, testName string, healthCheck bool, extraArgs ...string) *MCPANYTestServerInfo {
	t.Helper()

	root, err := GetProjectRoot()
	require.NoError(t, err, "Failed to get project root")
	mcpanyBinary := filepath.Join(root, "../build/bin/server")

	fmt.Printf("DEBUG: Using MCPANY binary from: %s\n", mcpanyBinary)

	// Use port 0 to let the OS assign free ports
	jsonrpcPortArg := "127.0.0.1:0"
	grpcRegPortArg := "127.0.0.1:0"
	natsURL, natsCleanup := StartNatsServer(t)

	args := []string{
		"run",
		"--mcp-listen-address", jsonrpcPortArg,
		"--grpc-port", grpcRegPortArg,
	}
	args = append(args, extraArgs...)
	env := []string{"MCPANY_LOG_LEVEL=debug", "NATS_URL=" + natsURL}
	if sudo, ok := os.LookupEnv("USE_SUDO_FOR_DOCKER"); ok {
		env = append(env, "USE_SUDO_FOR_DOCKER="+sudo)
	}
	if metricsAddr, ok := os.LookupEnv("MCPANY_METRICS_LISTEN_ADDRESS"); ok {
		t.Logf("Found MCPANY_METRICS_LISTEN_ADDRESS=%s in env, creating server with it", metricsAddr)
		env = append(env, "MCPANY_METRICS_LISTEN_ADDRESS="+metricsAddr)
	} else {
		// Fallback for tests if env var propagation fails
		t.Logf("MCPANY_METRICS_LISTEN_ADDRESS not found in env, using default localhost:19091")
		env = append(env, "MCPANY_METRICS_LISTEN_ADDRESS=localhost:19091")
	}

	absMcpAnyBinaryPath, err := filepath.Abs(mcpanyBinary)
	require.NoError(t, err, "Failed to get absolute path for MCPANY binary: %s", mcpanyBinary)
	_, err = os.Stat(absMcpAnyBinaryPath)
	require.NoError(t, err, "MCPANY binary not found at %s. Run 'make build'.", absMcpAnyBinaryPath)

	mcpProcess := NewManagedProcess(t, "MCPANYServer-"+testName, absMcpAnyBinaryPath, args, env)
	mcpProcess.cmd.Dir = root
	err = mcpProcess.Start()
	require.NoError(t, err, "Failed to start MCPANY server. Stderr: %s", mcpProcess.StderrString())

	// Wait for ports to be assigned and logged
	var jsonrpcPort, grpcRegPort int

	// Regex patterns to extract ports from logs.
	// Matches: msg="HTTP server listening" ... port=127.0.0.1:12345
	httpPortRegex := regexp.MustCompile(`msg="HTTP server listening".*?port=[^:]+:(\d+)`)
	// Matches: msg="gRPC server listening" ... port=127.0.0.1:12345
	// OR: INFO grpc_weather_server: Listening on port port=43523
	grpcPortRegex := regexp.MustCompile(`(?:msg="gRPC server listening".*?port=[^:]+:(\d+))|(?:Listening on port port=(\d+))`)

	require.Eventually(t, func() bool {
		stdout := mcpProcess.StdoutString()
		if jsonrpcPort == 0 {
			matches := httpPortRegex.FindStringSubmatch(stdout)
			if len(matches) >= 2 {
				if _, err := fmt.Sscanf(matches[1], "%d", &jsonrpcPort); err != nil {
					t.Logf("failed to parse jsonrpc port: %v", err)
				}
			}
		}
		if grpcRegPort == 0 {
			matches := grpcPortRegex.FindStringSubmatch(stdout)
			if len(matches) >= 2 {
				if _, err := fmt.Sscanf(matches[1], "%d", &grpcRegPort); err != nil {
					t.Logf("failed to parse grpc port: %v", err)
				}
			}
		}
		// If we are stdio mode, we might not get HTTP port if listen address is not set?
		// But we set --mcp-listen-address explicitly to 0. So it SHOULD listen.
		// NOTE: if stdio mode is used, we still pass network flags, and the server runs both stdio and http usually?
		// Or maybe not? runServerMode vs runStdioMode.
		// If --stdio is passed, `Run` calls `runStdioModeFunc` which DOES NOT start HTTP server for MCP usually?
		// But in `runServerMode`, we start HTTP.
		// Let's check `Run` in `server.go`:
		// if stdio { return a.runStdioModeFunc(...) }
		// `runStdioMode` only runs stdio transport. NO HTTP server.
		// So if --stdio is extracted from extraArgs, we won't get HTTP port log.

		isStdio := false
		for _, arg := range extraArgs {
			if arg == "--stdio" {
				isStdio = true
				break
			}
		}
		if isStdio {
			// In stdio mode, we don't expect HTTP/gRPC ports to be logged or relevant for *connecting* via network (except maybe metrics?).
			return true
		}

		return jsonrpcPort != 0 && grpcRegPort != 0
	}, McpAnyServerStartupTimeout, RetryInterval, "Failed to discover bound ports from logs.\nStdout: %s\nStderr: %s", mcpProcess.StdoutString(), mcpProcess.StderrString())

	// If stdio, we might not have ports.
	jsonrpcEndpoint := ""
	grpcRegEndpoint := ""
	mcpRequestURL := ""

	if jsonrpcPort != 0 {
		jsonrpcEndpoint = fmt.Sprintf("http://127.0.0.1:%d", jsonrpcPort)
		mcpRequestURL = jsonrpcEndpoint + "/mcp"
	}
	if grpcRegPort != 0 {
		grpcRegEndpoint = fmt.Sprintf("127.0.0.1:%d", grpcRegPort)
	}

	httpClient := &http.Client{Timeout: 2 * time.Second}
	var grpcRegConn *grpc.ClientConn
	var registrationClient apiv1.RegistrationServiceClient

	if healthCheck && jsonrpcPort != 0 { // Only check health if we have a port
		t.Logf("MCPANY server health check target URL: %s", mcpRequestURL)

		// Wait for gRPC readiness
		require.Eventually(t, func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			var errDial error
			grpcRegConn, errDial = grpc.NewClient(grpcRegEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if errDial != nil {
				return false
			}
			state := grpcRegConn.GetState()
			if state == connectivity.Ready || state == connectivity.Idle {
				return true
			}
			if !grpcRegConn.WaitForStateChange(ctx, state) {
				_ = grpcRegConn.Close()
				return false
			}
			return true
		}, McpAnyServerStartupTimeout, RetryInterval, "MCPANY gRPC endpoint %s not healthy.\nStdout: %s", grpcRegEndpoint, mcpProcess.StdoutString())

		registrationClient = apiv1.NewRegistrationServiceClient(grpcRegConn)

		// Wait for HTTP readiness
		require.Eventually(t, func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, "GET", mcpRequestURL, nil)
			if err != nil {
				return false
			}
			resp, err := httpClient.Do(req)
			if err != nil {
				return false
			}
			defer func() { _ = resp.Body.Close() }()
			return true
		}, McpAnyServerStartupTimeout, RetryInterval, "MCPANY HTTP endpoint %s not healthy.", mcpRequestURL)
	} else if healthCheck {
		// stdio mode health check
		mcpProcess.WaitForText(t, "MCPANY server is ready", McpAnyServerStartupTimeout) // Assumption or skipped?
	}

	t.Logf("MCPANY Server process started. MCP Endpoint Base: %s, gRPC Reg: %s", jsonrpcEndpoint, grpcRegEndpoint)

	return &MCPANYTestServerInfo{
		Process:                  mcpProcess,
		JSONRPCEndpoint:          jsonrpcEndpoint,
		HTTPEndpoint:             mcpRequestURL,
		GrpcRegistrationEndpoint: grpcRegEndpoint,
		HTTPClient:               httpClient,
		GRPCRegConn:              grpcRegConn,
		RegistrationClient:       registrationClient,
		NatsURL:                  natsURL,
		CleanupFunc: func() {
			t.Logf("Cleaning up MCPANYTestServerInfo for %s...", testName)
			if grpcRegConn != nil {
				_ = grpcRegConn.Close()
			}
			mcpProcess.Stop()
			natsCleanup()
		},
		T: t,
	}
}

// RegisterServiceViaAPI registers a service using the gRPC API.
func RegisterServiceViaAPI(t *testing.T, regClient apiv1.RegistrationServiceClient, req *apiv1.RegisterServiceRequest) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	resp, err := regClient.RegisterService(ctx, req)
	require.NoError(t, err, "Failed to register service %s via API.", req.GetConfig().GetName())
	require.NotNil(t, resp, "Nil response from RegisterService API for %s", req.GetConfig().GetName())
	t.Logf("Service %s registered via API successfully. Message: %s, Discovered tools:\n%v", req.GetConfig().GetName(), resp.GetMessage(), resp.GetDiscoveredTools())
}

// RegisterHTTPService registers a simple HTTP service.
func RegisterHTTPService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, baseURL, operationID, endpointPath, httpMethod string, authConfig *configv1.UpstreamAuthentication) {
	t.Helper()
	toolDef := configv1.ToolDefinition_builder{
		Name: &operationID,
	}.Build()
	RegisterHTTPServiceWithParams(t, regClient, serviceID, baseURL, toolDef, endpointPath, httpMethod, nil, authConfig)
}

// RegisterHTTPServiceWithParams registers an HTTP service with parameters.
func RegisterHTTPServiceWithParams(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, baseURL string, toolDef *configv1.ToolDefinition, endpointPath, httpMethod string, params []*configv1.HttpParameterMapping, authConfig *configv1.UpstreamAuthentication) {
	t.Helper()
	t.Logf("Registering HTTP service '%s' with endpoint path: %s", serviceID, endpointPath)

	httpMethodEnumName := "HTTP_METHOD_" + strings.ToUpper(httpMethod)
	if _, ok := configv1.HttpCallDefinition_HttpMethod_value[httpMethodEnumName]; !ok {
		t.Fatalf("Invalid HTTP method provided: %s", httpMethod)
	}
	method := configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value[httpMethodEnumName])

	callID := "call-" + toolDef.GetName()
	callDef := configv1.HttpCallDefinition_builder{
		Id:           &callID,
		EndpointPath: &endpointPath,
		Method:       &method,
		Parameters:   params,
	}.Build()
	toolDef.SetCallId(callID)

	upstreamServiceConfigBuilder := configv1.UpstreamServiceConfig_builder{
		Name: &serviceID,
		HttpService: configv1.HttpUpstreamService_builder{
			Address: &baseURL,
			Tools:   []*configv1.ToolDefinition{toolDef},
			Calls:   map[string]*configv1.HttpCallDefinition{callID: callDef},
		}.Build(),
	}
	if authConfig != nil {
		upstreamServiceConfigBuilder.UpstreamAuthentication = authConfig
	}
	config := upstreamServiceConfigBuilder.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	RegisterServiceViaAPI(t, regClient, req)
	t.Logf("HTTP Service '%s' registration request sent via API: %s %s%s", serviceID, httpMethod, baseURL, endpointPath)
}

// RegisterWebsocketService registers a WebSocket service.
func RegisterWebsocketService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, baseURL, operationID string, authConfig *configv1.UpstreamAuthentication) {
	t.Helper()
	t.Logf("Registering Websocket service '%s' with endpoint: %s", serviceID, baseURL)

	callID := "call-" + operationID
	callDef := configv1.WebsocketCallDefinition_builder{
		Id: &callID,
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   &operationID,
		CallId: &callID,
	}.Build()

	upstreamServiceConfigBuilder := configv1.UpstreamServiceConfig_builder{
		Name: &serviceID,
		WebsocketService: configv1.WebsocketUpstreamService_builder{
			Address: &baseURL,
			Tools:   []*configv1.ToolDefinition{toolDef},
			Calls:   map[string]*configv1.WebsocketCallDefinition{callID: callDef},
		}.Build(),
	}
	if authConfig != nil {
		upstreamServiceConfigBuilder.UpstreamAuthentication = authConfig
	}
	config := upstreamServiceConfigBuilder.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	RegisterServiceViaAPI(t, regClient, req)
	t.Logf("Websocket Service '%s' registration request sent via API: %s", serviceID, baseURL)
}

// RegisterWebrtcService registers a WebRTC service.
func RegisterWebrtcService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, baseURL, operationID string, authConfig *configv1.UpstreamAuthentication) {
	t.Helper()
	t.Logf("Registering Webrtc service '%s' with endpoint: %s", serviceID, baseURL)

	callID := "call-" + operationID
	callDef := configv1.WebrtcCallDefinition_builder{
		Id: &callID,
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   &operationID,
		CallId: &callID,
	}.Build()

	upstreamServiceConfigBuilder := configv1.UpstreamServiceConfig_builder{
		Name: &serviceID,
		WebrtcService: configv1.WebrtcUpstreamService_builder{
			Address: &baseURL,
			Tools:   []*configv1.ToolDefinition{toolDef},
			Calls:   map[string]*configv1.WebrtcCallDefinition{callID: callDef},
		}.Build(),
	}
	if authConfig != nil {
		upstreamServiceConfigBuilder.UpstreamAuthentication = authConfig
	}
	config := upstreamServiceConfigBuilder.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	RegisterServiceViaAPI(t, regClient, req)
	t.Logf("Webrtc Service '%s' registration request sent via API: %s", serviceID, baseURL)
}

// RegisterStreamableMCPService registers a streamable MCP service (SSE).
func RegisterStreamableMCPService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, targetURL string, toolAutoDiscovery bool, authConfig *configv1.UpstreamAuthentication) {
	t.Helper()

	mcpStreamableHTTPConnection := configv1.McpStreamableHttpConnection_builder{
		HttpAddress: &targetURL,
	}.Build()

	callID := "call-hello"
	callDef := configv1.MCPCallDefinition_builder{
		Id: &callID,
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name: proto.String("hello"),
	}.Build()
	toolDef.SetCallId(callID)

	upstreamServiceConfigBuilder := configv1.UpstreamServiceConfig_builder{
		Name: &serviceID,
		McpService: configv1.McpUpstreamService_builder{
			ToolAutoDiscovery: &toolAutoDiscovery,
			HttpConnection:    mcpStreamableHTTPConnection,
			Tools:             []*configv1.ToolDefinition{toolDef},
			Calls:             map[string]*configv1.MCPCallDefinition{callID: callDef},
		}.Build(),
	}
	if authConfig != nil {
		upstreamServiceConfigBuilder.UpstreamAuthentication = authConfig
	}
	config := upstreamServiceConfigBuilder.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	RegisterServiceViaAPI(t, regClient, req)
	t.Logf("Streamable MCP HTTP Service '%s' registration request sent via API: URL %s", serviceID, targetURL)
}

// RegisterStdioMCPService registers an MCP service using stdio.
func RegisterStdioMCPService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, command string, toolAutoDiscovery bool) {
	t.Helper()
	parts := strings.Fields(command)
	require.True(t, len(parts) > 0, "Command for stdio service cannot be empty")
	commandName := parts[0]
	commandArgs := parts[1:]
	RegisterStdioService(t, regClient, serviceID, commandName, toolAutoDiscovery, commandArgs...)
}

// RegisterGRPCService registers a gRPC service.
func RegisterGRPCService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, grpcTargetAddress string, authConfig *configv1.UpstreamAuthentication) {
	t.Helper()

	upstreamServiceConfigBuilder := configv1.UpstreamServiceConfig_builder{
		Name: &serviceID,
		GrpcService: configv1.GrpcUpstreamService_builder{
			Address:       &grpcTargetAddress,
			UseReflection: proto.Bool(true),
		}.Build(),
	}
	if authConfig != nil {
		upstreamServiceConfigBuilder.UpstreamAuthentication = authConfig
	}
	config := upstreamServiceConfigBuilder.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	RegisterServiceViaAPI(t, regClient, req)
	t.Logf("gRPC Service '%s' registration request sent via API: target %s", serviceID, grpcTargetAddress)
}

// RegisterStdioService registers a raw stdio service.
func RegisterStdioService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, commandName string, toolAutoDiscovery bool, commandArgs ...string) {
	t.Helper()
	RegisterStdioServiceWithSetup(t, regClient, serviceID, commandName, toolAutoDiscovery, "", "", nil, nil, commandArgs...)
}

// RegisterStdioServiceWithSetup registers a stdio service with setup steps.
func RegisterStdioServiceWithSetup(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, commandName string, toolAutoDiscovery bool, workingDir, containerImage string, setupCommands []string, env map[string]string, commandArgs ...string) {
	t.Helper()

	var secretEnv map[string]*configv1.SecretValue
	if env != nil {
		secretEnv = make(map[string]*configv1.SecretValue)
		for k, v := range env {
			secretEnv[k] = configv1.SecretValue_builder{
				PlainText: &v,
			}.Build()
		}
	}

	stdioConnection := configv1.McpStdioConnection_builder{
		Command:          &commandName,
		Args:             commandArgs,
		WorkingDirectory: &workingDir,
		ContainerImage:   &containerImage,
		SetupCommands:    setupCommands,
		Env:              secretEnv,
	}.Build()

	mcpService := configv1.McpUpstreamService_builder{
		ToolAutoDiscovery: &toolAutoDiscovery,
		StdioConnection:   stdioConnection,
	}.Build()

	upstreamServiceConfigBuilder := configv1.UpstreamServiceConfig_builder{
		Name:       &serviceID,
		McpService: mcpService,
	}
	cfg := upstreamServiceConfigBuilder.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: cfg,
	}.Build()

	fullCommand := append([]string{commandName}, commandArgs...)
	t.Logf("Registering stdio service '%s' with command: %v", serviceID, fullCommand)
	time.Sleep(250 * time.Millisecond)
	RegisterServiceViaAPI(t, regClient, req)
}

// RegisterOpenAPIService registers an OpenAPI service.
func RegisterOpenAPIService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, openAPISpecPath, serverURLOverride string, authConfig *configv1.UpstreamAuthentication) {
	t.Helper()
	absSpecPath, err := filepath.Abs(openAPISpecPath)
	require.NoError(t, err)
	_, err = os.Stat(absSpecPath)
	require.NoError(t, err, "OpenAPI spec file not found: %s", absSpecPath)
	specContent, err := os.ReadFile(absSpecPath) //nolint:gosec // Test file
	require.NoError(t, err)
	spec := string(specContent)

	openapiServiceDef := configv1.OpenapiUpstreamService_builder{
		SpecContent: &spec,
		Address:     &serverURLOverride,
	}.Build()

	upstreamServiceConfigBuilder := configv1.UpstreamServiceConfig_builder{
		Name:           &serviceID,
		OpenapiService: openapiServiceDef,
	}
	if authConfig != nil {
		upstreamServiceConfigBuilder.UpstreamAuthentication = authConfig
	}
	config := upstreamServiceConfigBuilder.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	RegisterServiceViaAPI(t, regClient, req)
	t.Logf("OpenAPI Service '%s' registration request sent via API (spec: %s, intended server: %s)", serviceID, absSpecPath, serverURLOverride)
}

// RegisterHTTPServiceWithJSONRPC registers an HTTP service using the JSON-RPC endpoint.
func RegisterHTTPServiceWithJSONRPC(t *testing.T, mcpanyEndpoint, serviceID, baseURL, operationID, endpointPath, httpMethod string, authConfig *configv1.UpstreamAuthentication) {
	t.Helper()
	t.Logf("Registering HTTP service '%s' via JSON-RPC with endpoint path: %s", serviceID, endpointPath)

	httpMethodEnumName := "HTTP_METHOD_" + strings.ToUpper(httpMethod)
	if _, ok := configv1.HttpCallDefinition_HttpMethod_value[httpMethodEnumName]; !ok {
		t.Fatalf("Invalid HTTP method provided: %s", httpMethod)
	}
	method := configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value[httpMethodEnumName])

	toolDef := configv1.ToolDefinition_builder{Name: &operationID}.Build()
	callID := "call-" + toolDef.GetName()
	callDef := configv1.HttpCallDefinition_builder{
		Id:           &callID,
		EndpointPath: &endpointPath,
		Method:       &method,
	}.Build()
	toolDef.SetCallId(callID)

	upstreamServiceConfigBuilder := configv1.UpstreamServiceConfig_builder{
		Name: &serviceID,
		HttpService: configv1.HttpUpstreamService_builder{
			Address: &baseURL,
			Tools:   []*configv1.ToolDefinition{toolDef},
			Calls:   map[string]*configv1.HttpCallDefinition{callID: callDef},
		}.Build(),
	}
	if authConfig != nil {
		upstreamServiceConfigBuilder.UpstreamAuthentication = authConfig
	}
	config := upstreamServiceConfigBuilder.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	// Use protojson to marshal the request to JSON
	jsonBytes, err := protojson.Marshal(req)
	require.NoError(t, err)

	var params json.RawMessage = jsonBytes

	jsonRPCReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "registration/register",
		"params":  params,
		"id":      "1",
	}

	reqBody, err := json.Marshal(jsonRPCReq)
	require.NoError(t, err)

	resp, err := http.Post(mcpanyEndpoint, "application/json", bytes.NewBuffer(reqBody)) //nolint:gosec
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK")

	var rpcResp struct {
		Result json.RawMessage  `json:"result"`
		Error  *MCPJSONRPCError `json:"error"`
	}
	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	require.NoError(t, err, "Failed to decode JSON-RPC response")
	require.Nil(t, rpcResp.Error, "Received JSON-RPC error: %v", rpcResp.Error)

	t.Logf("HTTP Service '%s' registration request sent via JSON-RPC successfully.", serviceID)
}

// MCPJSONRPCError represents a JSON-RPC error.
type MCPJSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *MCPJSONRPCError) Error() string {
	return fmt.Sprintf("JSON-RPC Error: Code=%d, Message=%s, Data=%v", e.Code, e.Message, e.Data)
}

// ListTools calls tools/list via JSON-RPC.
func (s *MCPANYTestServerInfo) ListTools(ctx context.Context) (*mcp.ListToolsResult, error) {
	reqBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "tools/list",
		"params":  map[string]interface{}{},
		"id":      1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.HTTPEndpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var rpcResp struct {
		Result *mcp.ListToolsResult `json:"result"`
		Error  *MCPJSONRPCError     `json:"error"`
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(bodyBytes, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w. Body: %s", err, string(bodyBytes))
	}

	if rpcResp.Error != nil {
		return nil, rpcResp.Error
	}

	return rpcResp.Result, nil
}

// CallTool calls tools/call via JSON-RPC.
func (s *MCPANYTestServerInfo) CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	reqBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "tools/call",
		"params":  params,
		"id":      1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.HTTPEndpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var rpcResp struct {
		Result *mcp.CallToolResult `json:"result"`
		Error  *MCPJSONRPCError    `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, rpcResp.Error
	}

	return rpcResp.Result, nil
}

// WaitForPortFromLogs waits for a log line indicating the server is listening and extracts the address.
func WaitForPortFromLogs(t *testing.T, mp *ManagedProcess, serverName string) (string, error) {
	t.Helper()
	var port string
	// Regex to find port=ADDRESS. We expect it might be quoted.
	re := regexp.MustCompile(`port=([^ ]+)`)

	checkLog := func() bool {
		output := mp.StdoutString()
		lines := strings.Split(output, "\n")
		// Reverse iterate to find latest logs first?
		// No, standard iteration is fine, usually only one startup log.
		for _, line := range lines {
			// Look for server name. It is usually logged as server="Name" or server=Name
			if strings.Contains(line, fmt.Sprintf(`server="%s"`, serverName)) || strings.Contains(line, fmt.Sprintf(`server=%s`, serverName)) {
				if strings.Contains(line, "listening") {
					matches := re.FindStringSubmatch(line)
					if len(matches) > 1 {
						port = matches[1]
						port = strings.Trim(port, `"`)
						return true
					}
				}
			}
		}
		return false
	}

	require.Eventually(t, checkLog, McpAnyServerStartupTimeout, 100*time.Millisecond, "Failed to find listening port for %s in logs.\nStdout:\n%s", serverName, mp.StdoutString())
	return port, nil
}
