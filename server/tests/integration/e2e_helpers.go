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
	"strconv"
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

	"github.com/mcpany/core/server/pkg/app"
	"github.com/spf13/afero"
)

// CreateTempConfigFile creates a temporary configuration file for the configured upstream service.
//
// t is the t.
// config holds the configuration settings.
//
// Returns the result.
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
//
// t is the t.
//
// Returns the result.
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
//
// p is the p.
//
// Returns the result.
// Returns an error if the operation fails.
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
//
// t is the t.
//
// Returns the result.
func ProjectRoot(t *testing.T) string {
	t.Helper()
	root, err := GetProjectRoot()
	require.NoError(t, err)
	return root
}

const (
	// McpAnyServerStartupTimeout is the timeout for the server to start.
	McpAnyServerStartupTimeout = 300 * time.Second
	// ServiceStartupTimeout is the timeout for services to start up.
	ServiceStartupTimeout = 120 * time.Second
	// TestWaitTimeShort is a short wait time for tests.
	TestWaitTimeShort = 120 * time.Second
	// TestWaitTimeMedium is the default timeout for medium duration tests.
	TestWaitTimeMedium = 480 * time.Second
	// TestWaitTimeLong is the default timeout for long duration tests.
	TestWaitTimeLong = 10 * time.Minute
	// RetryInterval is the interval between retries.
	RetryInterval           = 250 * time.Millisecond
	localHeaderMcpSessionID = "Mcp-Session-Id"
	dockerCmd               = "docker"
	sudoCmd                 = "sudo"
	// LoopbackIP is the default loopback IP for testing.
	LoopbackIP              = "127.0.0.1"
	loopbackIP              = LoopbackIP
	dynamicBindAddr         = loopbackIP + ":0"
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
			cmd := exec.CommandContext(context.Background(), dockerCmd, "info")
			if err := cmd.Run(); err == nil {
				dockerCommand = dockerCmd
				dockerArgs = []string{}
				return
			}
		}

		// If direct access fails, check for passwordless sudo.
		if _, err := exec.LookPath(sudoCmd); err == nil {
			cmd := exec.CommandContext(context.Background(), sudoCmd, "-n", dockerCmd, "info")
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
//
// Returns the result.
// Returns an error if the operation fails.
func GetProjectRoot() (string, error) {
	var err error
	findRootOnce.Do(func() {
		// Allow overriding via environment variable
		if envRoot := os.Getenv("MCPANY_PROJECT_ROOT"); envRoot != "" {
			// Validate that the directory exists and contains go.mod
			if _, err := os.Stat(filepath.Join(envRoot, "go.mod")); err == nil {
				projectRoot = envRoot
				return
			}
			// If invalid, ignore and fallback to auto-detection (useful if env var is set for container but running on host)
		}

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

// --- Helper: Find Free Port ---.
var portMutex sync.Mutex

// FindFreePort finds a free TCP port on localhost.
//
// t is the t.
//
// Returns the result.
func FindFreePort(t *testing.T) int {
	portMutex.Lock()
	defer portMutex.Unlock()
	t.Helper()
	addr, err := net.ResolveTCPAddr("tcp", dynamicBindAddr)
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
//
// t is the t.
// label is the label.
// command is the command.
// args is the args.
// env is the env.
//
// Returns the result.
func NewManagedProcess(t *testing.T, label, command string, args []string, env []string) *ManagedProcess {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), command, args...)
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

	// Ensure process is stopped when test ends to avoid race conditions on t.Logf
	t.Cleanup(mp.Stop)

	return mp
}

// Cmd returns the underlying exec.Cmd.
//
// Returns the result.
func (mp *ManagedProcess) Cmd() *exec.Cmd {
	return mp.cmd
}

// Start starts the process.
//
// Returns an error if the operation fails.
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
		// Ensure we close waitDone AFTER all logging is complete to prevent data races
		// where the test proceeds and finishes (invalidating 't') while we are still logging.
		defer close(mp.waitDone)

		err := mp.cmd.Wait()
		// Log output regardless of error, can be useful for debugging successful exits too
		// mp.t.Logf("[%s] Process %s finished. Stdout:\n%s\nStderr:\n%s", mp.label, mp.cmd.Path, mp.stdout.String(), mp.stderr.String())
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

// Allow patching for testing.
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
//
// Returns the result.
func (mp *ManagedProcess) StdoutString() string { return mp.stdout.String() }

// StderrString returns the captured stderr as a string.
//
// Returns the result.
func (mp *ManagedProcess) StderrString() string { return mp.stderr.String() }

// WaitForText waits for specific text to appear in the process's stdout.
//
// t is the t.
// text is the text.
// timeout is the timeout.
func (mp *ManagedProcess) WaitForText(t *testing.T, text string, timeout time.Duration) {
	t.Helper()
	require.Eventually(t, func() bool {
		return strings.Contains(mp.StdoutString(), text)
	}, timeout, RetryInterval, "Text '%s' not found in stdout for process '%s' in time.\nStdout: %s\nStderr: %s", text, mp.label, mp.StdoutString(), mp.StderrString())
}

// WaitForTCPPort waits for a TCP port to become open and accepting connections.
//
// t is the t.
// port is the port.
// timeout is the timeout.
func WaitForTCPPort(t *testing.T, port int, timeout time.Duration) {
	t.Helper()
	require.Eventually(t, func() bool {
		d := net.Dialer{Timeout: 100 * time.Millisecond}
		conn, err := d.DialContext(context.Background(), "tcp", net.JoinHostPort(loopbackIP, strconv.Itoa(port)))
		if err != nil {
			return false // Port is not open yet
		}
		_ = conn.Close()
		return true // Port is open
	}, timeout, 250*time.Millisecond, "Port %d did not become available in time", port)
}

// WaitForGRPCReady waits for a gRPC server to become ready by attempting to connect.
//
// t is the t.
// grpcAddress is the grpcAddress.
// timeout is the timeout.
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
//
// t is the t.
// url is the url.
// timeout is the timeout.
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
//
// t is the t.
// url is the url.
// timeout is the timeout.
func WaitForHTTPHealth(t *testing.T, url string, timeout time.Duration) {
	t.Helper()
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	require.Eventually(t, func() bool {
		req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
		if err != nil {
			return false
		}
		resp, err := client.Do(req)
		if err != nil {
			return false
		}
		defer func() { _ = resp.Body.Close() }()
		return resp.StatusCode == http.StatusOK
	}, timeout, 250*time.Millisecond, "URL %s did not become healthy in time", url)
}

// IsDockerSocketAccessible checks if the Docker daemon is accessible.
//
// Returns true if successful.
func IsDockerSocketAccessible() bool {
	dockerExe, dockerArgs := getDockerCommand()

	cmd := exec.CommandContext(context.Background(), dockerExe, append(dockerArgs, "info")...) //nolint:gosec // Test helper
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// --- Mock Service Start Helpers (External Processes) ---

// StartDockerContainer starts a docker container with the given image and args.
//
// t is the t.
// imageName is the imageName.
// containerName is the containerName.
// runArgs is the runArgs.
// command is the command.
//
// Returns the result.
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

	stopCmd := exec.CommandContext(context.Background(), dockerExe, buildArgs("stop", containerName)...) //nolint:gosec // Test helper
	_ = stopCmd.Run()                                                                                    // Ignore error, it might not be running

	rmCmd := exec.CommandContext(context.Background(), dockerExe, buildArgs("rm", containerName)...) //nolint:gosec // Test helper
	_ = rmCmd.Run()                                                                                  // Ignore error, it might not exist

	dockerRunArgs := []string{"run", "--name", containerName, "--rm"}
	dockerRunArgs = append(dockerRunArgs, runArgs...)
	dockerRunArgs = append(dockerRunArgs, imageName)
	dockerRunArgs = append(dockerRunArgs, command...)

	startCmd := exec.CommandContext(context.Background(), dockerExe, buildArgs(dockerRunArgs...)...) //nolint:gosec // Test helper
	// Capture stderr for better error reporting
	var stderr bytes.Buffer
	startCmd.Stderr = &stderr

	// Use Run instead of Start for 'docker run -d' to ensure the command completes
	// and the container is running before proceeding.
	err := startCmd.Run()
	require.NoError(t, err, "failed to start docker container %s. Stderr: %s", imageName, stderr.String())

	cleanupFunc = func() {
		t.Logf("Stopping and removing docker container: %s", containerName)

		stopCleanupCmd := exec.CommandContext(context.Background(), dockerExe, buildArgs("stop", containerName)...) //nolint:gosec // Test helper
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
	MetricsEndpoint          string
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
//
// t is the t.
//
// Returns the result.
func StartWebsocketEchoServer(t *testing.T) *WebsocketEchoServerInfo {
	t.Helper()

	port := FindFreePort(t)
	addr := net.JoinHostPort(loopbackIP, strconv.Itoa(port))

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
//
// t is the t.
// testName is the testName.
// configContent is the configContent.
//
// Returns the result.
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
//
// t is the t.
// testName is the testName.
// extraArgs is the extraArgs.
//
// Returns the result.
func StartMCPANYServer(t *testing.T, testName string, extraArgs ...string) *MCPANYTestServerInfo {
	return StartMCPANYServerWithClock(t, testName, true, extraArgs...)
}

// StartMCPANYServerWithNoHealthCheck starts the MCP Any server but skips the health check.
//
// t is the t.
// testName is the testName.
// extraArgs is the extraArgs.
//
// Returns the result.
func StartMCPANYServerWithNoHealthCheck(t *testing.T, testName string, extraArgs ...string) *MCPANYTestServerInfo {
	return StartMCPANYServerWithClock(t, testName, false, extraArgs...)
}

// StartInProcessMCPANYServer starts an in-process MCP Any server for testing.
//
// t is the t.
// _ is an unused parameter.
// apiKey is the apiKey.
//
// Returns the result.
func StartInProcessMCPANYServer(t *testing.T, _ string, apiKey ...string) *MCPANYTestServerInfo {
	t.Helper()

	var actualAPIKey string
	if len(apiKey) > 0 {
		actualAPIKey = apiKey[0]
	}

	_, err := GetProjectRoot()
	require.NoError(t, err, "Failed to get project root")

	// Use port 0 for dynamic allocation to avoid race conditions
	jsonrpcAddress := ":0"
	grpcRegAddress := ":0"

	ctx, cancel := context.WithCancel(context.Background())

	// Use unique DB path to avoid SQLITE_BUSY conflicts
	dbFile, err := os.CreateTemp(t.TempDir(), "mcpany-in-process-*.db")
	require.NoError(t, err)
	dbPath := dbFile.Name()
	require.NoError(t, dbFile.Close())
	t.Setenv("MCPANY_DB_PATH", dbPath)

	appRunner := app.NewApplication()
	runErrCh := make(chan error, 1)
	go func() {
		defer cancel() // Ensure WaitForStartup doesn't hang if Run returns
		opts := app.RunOptions{
			Ctx:             ctx,
			Fs:              afero.NewOsFs(),
			Stdio:           false,
			JSONRPCPort:     jsonrpcAddress,
			GRPCPort:        grpcRegAddress,
			ConfigPaths:     []string{},
			APIKey:          actualAPIKey,
			ShutdownTimeout: 5 * time.Second,
		}
		err := appRunner.Run(opts)
		if err != nil {
			if ctx.Err() == nil {
				t.Logf("Application run error: %v", err)
			}
			runErrCh <- err
		}
		close(runErrCh)
	}()

	// Wait for startup or failure
	startupErrCh := make(chan error, 1)
	go func() {
		startupErrCh <- appRunner.WaitForStartup(ctx)
	}()

	select {
	case err := <-startupErrCh:
		require.NoError(t, err, "Failed to wait for application startup")
	case err := <-runErrCh:
		require.NoError(t, err, "Application run failed prematurely")
	case <-time.After(McpAnyServerStartupTimeout):
		require.Fail(t, "Startup timed out")
	}

	// Retrieve dynamically allocated ports
	jsonrpcPort := int(appRunner.BoundHTTPPort.Load())
	grpcRegPort := int(appRunner.BoundGRPCPort.Load())

	// Fallback/Safety check
	if jsonrpcPort == 0 || grpcRegPort == 0 {
		// Retry logic or fail? WaitForStartup should guarantee connection, but let's check.
		// If WaitForStartup uses HTTP health check internally using the configured port...
		// Wait, WaitForStartup in app.go checks readiness.
		// But app.Run populates Bound*Port BEFORE calling readiness check loop?
		// Let's assume so.
		t.Logf("Warning: Bound ports are 0 after startup. HTTP: %d, gRPC: %d", jsonrpcPort, grpcRegPort)
	}

	jsonrpcEndpoint := fmt.Sprintf("http://%s:%d", loopbackIP, jsonrpcPort)
	grpcRegEndpoint := net.JoinHostPort(loopbackIP, strconv.Itoa(grpcRegPort))
	mcpRequestURL := jsonrpcEndpoint + "/mcp"
	if actualAPIKey != "" {
		mcpRequestURL += "?api_key=" + actualAPIKey
	}

	// Verify gRPC connection
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
//
// t is the t.
//
// Returns the result.
// Returns the result.
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

	// Use -p -1 to let NATS pick a random free port
	cmd := exec.CommandContext(context.Background(), natsServerBin, "-p", "-1", "-a", loopbackIP) //nolint:gosec // Test helper

	// Capture output to find the port
	// We need a thread-safe buffer because we might read while it writes (though wait loop handles this?)
	// Actually, we can just use a pipe reading approach or shared buffer
	// Use threadSafeBuffer to avoid race conditions when reading logs while process writes
	var stdout, stderr threadSafeBuffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Start()
	require.NoError(t, err, "Failed to start nats-server")

	// Wait for port log
	var natsPort int
	// NATS logs: "Listening for client connections on 127.0.0.1:4222"
	// Or sometimes on stderr depending on config/version? Usually stdout.
	regexPort := regexp.MustCompile(`Listening for client connections on [^:]+:(\d+)`)

	start := time.Now()
	found := false
	for time.Since(start) < 10*time.Second {
		output := stdout.String() + stderr.String()
		matches := regexPort.FindStringSubmatch(output)
		if len(matches) >= 2 {
			if _, err := fmt.Sscanf(matches[1], "%d", &natsPort); err != nil {
				t.Logf("failed to parse nats port: %v", err)
				continue
			}
			found = true
			break
		}
		// Also check if process exited
		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			require.Fail(t, "nats-server exited unexpectedly", "Output:\n%s", output)
		}
		time.Sleep(100 * time.Millisecond)
	}
	require.True(t, found, "Failed to find NATS port in logs within timeout. Output:\n%s\nStderr:\n%s", stdout.String(), stderr.String())

	natsURL := fmt.Sprintf("nats://%s:%d", loopbackIP, natsPort)
	WaitForTCPPort(t, natsPort, 5*time.Second) // Double check availability

	cleanup := func() {
		_ = cmd.Process.Kill()
	}
	return natsURL, cleanup
}

// StartRedisContainer starts a Redis container for testing.
// StartRedisContainer starts a Redis container for testing.
func StartRedisContainer(t *testing.T) (redisAddr string, cleanupFunc func()) {
	t.Helper()
	require.True(t, IsDockerSocketAccessible(), "Docker is not running or accessible. Please start Docker to run this test.")

	containerName := fmt.Sprintf("mcpany-redis-test-%d", time.Now().UnixNano())
	// Use port 0 for dynamic host port allocation
	runArgs := []string{
		"-d",
		"-p", net.JoinHostPort(loopbackIP, "0") + ":6379",
	}

	command := []string{
		"redis-server",
		"--bind", "0.0.0.0",
	}

	cleanup := StartDockerContainer(t, "mirror.gcr.io/library/redis:latest", containerName, runArgs, command...)

	// Inspect the container to get the assigned port
	dockerExe, dockerBaseArgs := getDockerCommand()
	var hostPort int

	// Wait for port to be assigned (it happens immediately on start, but good to retry on inspect failure)
	require.Eventually(t, func() bool {
		// Safely append to a new slice to avoid modifying backing array of dockerBaseArgs if it has capacity
		portArgs := make([]string, 0, len(dockerBaseArgs)+3)
		portArgs = append(portArgs, dockerBaseArgs...)
		portArgs = append(portArgs, "port", containerName, "6379/tcp")

		cmd := exec.CommandContext(context.Background(), dockerExe, portArgs...) //nolint:gosec // Test helper
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("docker port failed: %v, output: %s", err, string(output))
			return false
		}
		// Output format: 127.0.0.1:32768
		outputStr := strings.TrimSpace(string(output))
		// Handle potential multiple lines or mappings, just take the first one
		lines := strings.Split(outputStr, "\n")
		if len(lines) == 0 {
			return false
		}
		parts := strings.Split(lines[0], ":")
		if len(parts) < 2 {
			return false
		}
		portStr := parts[len(parts)-1]
		p, err := strconv.Atoi(portStr)
		if err != nil {
			return false
		}
		hostPort = p
		return true
	}, 10*time.Second, 500*time.Millisecond, "Failed to inspect container port")

	redisAddr = net.JoinHostPort(loopbackIP, strconv.Itoa(hostPort))
	t.Logf("Redis started at %s", redisAddr)

	// Wait for Redis to be ready
	require.Eventually(t, func() bool {
		// Use redis-cli to ping the server
		pingArgs := append(dockerBaseArgs, "exec", containerName, "redis-cli", "ping") //nolint:gocritic // Helper
		cmd := exec.CommandContext(context.Background(), dockerExe, pingArgs...)       //nolint:gosec // Test helper
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
//
// t is the t.
// testName is the testName.
// healthCheck is the healthCheck.
// extraArgs is the extraArgs.
//
// Returns the result.
func StartMCPANYServerWithClock(t *testing.T, testName string, healthCheck bool, extraArgs ...string) *MCPANYTestServerInfo {
	t.Helper()

	root, err := GetProjectRoot()
	require.NoError(t, err, "Failed to get project root")
	mcpanyBinary := filepath.Join(root, "../build/bin/server")

	fmt.Printf("DEBUG: Using MCPANY binary from: %s\n", mcpanyBinary)

	// Use port 0 to let the OS assign free ports
	jsonrpcPortArg := dynamicBindAddr
	grpcRegPortArg := dynamicBindAddr
	natsURL, natsCleanup := StartNatsServer(t)

	// Use unique DB path
	dbFile, err := os.CreateTemp(t.TempDir(), "mcpany-db-*.db")
	require.NoError(t, err)
	dbPath := dbFile.Name()
	_ = dbFile.Close()

	args := []string{
		"run",
		"--mcp-listen-address", jsonrpcPortArg,
		"--grpc-port", grpcRegPortArg,
		"--db-path", dbPath,
	}
	args = append(args, extraArgs...)
	env := []string{"MCPANY_LOG_LEVEL=debug", "NATS_URL=" + natsURL, "MCPANY_DANGEROUS_ALLOW_LOCAL_IPS=true", "MCPANY_ENABLE_FILE_CONFIG=true"}
	if sudo, ok := os.LookupEnv("USE_SUDO_FOR_DOCKER"); ok {
		env = append(env, "USE_SUDO_FOR_DOCKER="+sudo)
	}
	// Metrics port
	// Use port 0 for dynamic allocation
	metricsPortArg := dynamicBindAddr
	env = append(env, "MCPANY_METRICS_LISTEN_ADDRESS="+metricsPortArg)
	t.Logf("Using dynamic metrics address request: %s", metricsPortArg)

	absMcpAnyBinaryPath, err := filepath.Abs(mcpanyBinary)
	require.NoError(t, err, "Failed to get absolute path for MCPANY binary: %s", mcpanyBinary)
	_, err = os.Stat(absMcpAnyBinaryPath)
	require.NoError(t, err, "MCPANY binary not found at %s. Run 'make build'.", absMcpAnyBinaryPath)

	// Generate a random API key for this test
	apiKey := fmt.Sprintf("test-key-%d", time.Now().UnixNano())
	args = append(args, "--api-key", apiKey)

	mcpProcess := NewManagedProcess(t, "MCPANYServer-"+testName, absMcpAnyBinaryPath, args, env)
	mcpProcess.cmd.Dir = root
	err = mcpProcess.Start()
	require.NoError(t, err, "Failed to start MCPANY server. Stderr: %s", mcpProcess.StderrString())

	// Wait for ports to be assigned and logged
	var jsonrpcPort, grpcRegPort, metricsPort int

	// Regex patterns to extract ports from logs.
	// Matches: msg="HTTP server listening" ... port=127.0.0.1:12345
	httpPortRegex := regexp.MustCompile(`msg="HTTP server listening".*?port=[^:]+:(\d+)`)
	// Matches: msg="gRPC server listening" ... port=127.0.0.1:12345
	// OR: INFO grpc_weather_server: Listening on port port=43523
	grpcPortRegex := regexp.MustCompile(`(?:msg="gRPC server listening".*?port=[^:]+:(\d+))|(?:Listening on port port=(\d+))`)
	// Matches: Metrics server listening on port 12345
	metricsPortRegex := regexp.MustCompile(`Metrics server listening on port (\d+)`)

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
		if metricsPort == 0 {
			matches := metricsPortRegex.FindStringSubmatch(stdout)
			if len(matches) >= 2 {
				if _, err := fmt.Sscanf(matches[1], "%d", &metricsPort); err != nil {
					t.Logf("failed to parse metrics port: %v", err)
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
		jsonrpcEndpoint = fmt.Sprintf("http://%s:%d", loopbackIP, jsonrpcPort)
		// Include API Key in URL query param for easy auth
		mcpRequestURL = fmt.Sprintf("%s/mcp?api_key=%s", jsonrpcEndpoint, apiKey)
	}
	if grpcRegPort != 0 {
		grpcRegEndpoint = net.JoinHostPort(loopbackIP, strconv.Itoa(grpcRegPort))
	}

	httpClient := &http.Client{Timeout: 2 * time.Second}
	var grpcRegConn *grpc.ClientConn
	var registrationClient apiv1.RegistrationServiceClient

	// Create a random Session ID for this server instance (client side ID)
	sessionID := fmt.Sprintf("test-session-%d", time.Now().UnixNano())

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
			// Use the URL with API key for health check too?
			// But health check usually hits healthz? NO, this checks mcpRequestURL (SSE/POST).
			// mcpRequestURL now has ?api_key=...
			req, err := http.NewRequestWithContext(ctx, "GET", mcpRequestURL, nil)
			if err != nil {
				return false
			}
			resp, err := httpClient.Do(req)
			if err != nil {
				return false
			}
			defer func() { _ = resp.Body.Close() }()
			// GET /mcp might fail method not allowed if only POST is supported?
			// But we just check if it's reachable.
			// Actually, server/pkg/transport/http/sse.go supports GET for SSE.
			return true
		}, McpAnyServerStartupTimeout, RetryInterval, "MCPANY HTTP endpoint %s not healthy.", mcpRequestURL)
	} else if healthCheck {
		// stdio mode health check
		mcpProcess.WaitForText(t, "MCPANY server is ready", McpAnyServerStartupTimeout) // Assumption or skipped?
	}

	t.Logf("MCPANY Server process started. MCP Endpoint Base: %s, gRPC Reg: %s, SessionID: %s, APIKey: %s", jsonrpcEndpoint, grpcRegEndpoint, sessionID, apiKey)

	return &MCPANYTestServerInfo{
		Process:                  mcpProcess,
		JSONRPCEndpoint:          jsonrpcEndpoint,
		HTTPEndpoint:             mcpRequestURL,
		GrpcRegistrationEndpoint: grpcRegEndpoint,
		MetricsEndpoint:          net.JoinHostPort(loopbackIP, strconv.Itoa(metricsPort)),
		HTTPClient:               httpClient,
		GRPCRegConn:              grpcRegConn,
		RegistrationClient:       registrationClient,
		NatsURL:                  natsURL,
		SessionID:                sessionID,
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

// Initialize performs the MCP initialization handshake.
//
// ctx is the context for the request.
//
// Returns an error if the operation fails.
func (s *MCPANYTestServerInfo) Initialize(ctx context.Context) error {
	// 1. Send initialize request
	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
		"id": 1,
	}

	reqBody, _ := json.Marshal(initReq)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.HTTPEndpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	// Do NOT set Mcp-Session-Id for initialize, let server generate it if needed.
	// Or maybe we need to support both modes?
	// If we send it, server says 404 session not found. So we shouldn't send it for new session?

	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("initialize failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// Capture Session ID from response header
	if sid := resp.Header.Get("Mcp-Session-Id"); sid != "" {
		s.SessionID = sid
		// Update T.Log to show we got a session ID
		if s.T != nil {
			s.T.Logf("Obtained Session ID from server: %s", s.SessionID)
		}
	}

	// We don't strictly need to parse result for tests unless we use capabilities,
	// but we MUST send initialized notification.

	// 2. Send initialized notification
	notifyReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
		"params":  map[string]interface{}{},
	}
	reqBody, _ = json.Marshal(notifyReq)
	httpReq, err = http.NewRequestWithContext(ctx, "POST", s.HTTPEndpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	httpReq.Header.Set("Mcp-Session-Id", s.SessionID)

	respNotify, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() { _ = respNotify.Body.Close() }()

	if respNotify.StatusCode != http.StatusOK && respNotify.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(respNotify.Body)
		return fmt.Errorf("initialized notification failed: status=%d, body=%s", respNotify.StatusCode, string(body))
	}
	return nil
}

// parseMCPResponse parses the response body, handling both JSON and SSE formats.
func parseMCPResponse(_ *testing.T, resp *http.Response) ([]byte, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "text/event-stream" || bytes.HasPrefix(bodyBytes, []byte("event: ")) {
		// Simple SSE parser
		lines := bytes.Split(bodyBytes, []byte("\n"))
		for _, line := range lines {
			if bytes.HasPrefix(line, []byte("data: ")) {
				data := bytes.TrimPrefix(line, []byte("data: "))
				return data, nil
			}
		}
		return nil, fmt.Errorf("failed to find valid JSON in SSE response. Body: %s", string(bodyBytes))
	}
	return bodyBytes, nil
}

// ListTools calls tools/list via JSON-RPC.
//
// ctx is the context for the request.
//
// Returns the result.
// Returns an error if the operation fails.
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
	if s.SessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", s.SessionID)
	}

	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	data, err := parseMCPResponse(s.T, resp)
	if err != nil {
		return nil, err
	}

	var rpcResp struct {
		Result *mcp.ListToolsResult `json:"result"`
		Error  *MCPJSONRPCError     `json:"error"`
	}
	if err := json.Unmarshal(data, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w. Body: %s", err, string(data))
	}

	if rpcResp.Error != nil {
		return nil, rpcResp.Error
	}

	return rpcResp.Result, nil
}

// CallTool calls tools/call via JSON-RPC.
//
// ctx is the context for the request.
// params is the params.
//
// Returns the result.
// Returns an error if the operation fails.
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
	if s.SessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", s.SessionID)
	}

	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	data, err := parseMCPResponse(s.T, resp)
	if err != nil {
		return nil, err
	}

	var rpcResp struct {
		Result *mcp.CallToolResult `json:"result"`
		Error  *MCPJSONRPCError    `json:"error"`
	}
	if err := json.Unmarshal(data, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, rpcResp.Error
	}

	return rpcResp.Result, nil
}

// RegisterServiceViaAPI registers a service using the gRPC API.
//
// t is the t.
// regClient is the regClient.
// req is the request object.
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
//
// t is the t.
// regClient is the regClient.
// serviceID is the serviceID.
// baseURL is the baseURL.
// operationID is the operationID.
// endpointPath is the endpointPath.
// httpMethod is the httpMethod.
// authConfig is the authConfig.
func RegisterHTTPService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, baseURL, operationID, endpointPath, httpMethod string, authConfig *configv1.Authentication) {
	t.Helper()
	toolDef := configv1.ToolDefinition_builder{
		Name: &operationID,
	}.Build()
	RegisterHTTPServiceWithParams(t, regClient, serviceID, baseURL, toolDef, endpointPath, httpMethod, nil, authConfig)
}

// RegisterHTTPServiceWithParams registers an HTTP service with parameters.
//
// t is the t.
// regClient is the regClient.
// serviceID is the serviceID.
// baseURL is the baseURL.
// toolDef is the toolDef.
// endpointPath is the endpointPath.
// httpMethod is the httpMethod.
// params is the params.
// authConfig is the authConfig.
func RegisterHTTPServiceWithParams(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, baseURL string, toolDef *configv1.ToolDefinition, endpointPath, httpMethod string, params []*configv1.HttpParameterMapping, authConfig *configv1.Authentication) {
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
		upstreamServiceConfigBuilder.UpstreamAuth = authConfig
	}
	config := upstreamServiceConfigBuilder.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	RegisterServiceViaAPI(t, regClient, req)
	t.Logf("HTTP Service '%s' registration request sent via API: %s %s%s", serviceID, httpMethod, baseURL, endpointPath)
}

// RegisterWebsocketService registers a WebSocket service.
//
// t is the t.
// regClient is the regClient.
// serviceID is the serviceID.
// baseURL is the baseURL.
// operationID is the operationID.
// authConfig is the authConfig.
func RegisterWebsocketService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, baseURL, operationID string, authConfig *configv1.Authentication) {
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
		upstreamServiceConfigBuilder.UpstreamAuth = authConfig
	}
	config := upstreamServiceConfigBuilder.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	RegisterServiceViaAPI(t, regClient, req)
	t.Logf("Websocket Service '%s' registration request sent via API: %s", serviceID, baseURL)
}

// RegisterWebrtcService registers a WebRTC service.
//
// t is the t.
// regClient is the regClient.
// serviceID is the serviceID.
// baseURL is the baseURL.
// operationID is the operationID.
// authConfig is the authConfig.
func RegisterWebrtcService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, baseURL, operationID string, authConfig *configv1.Authentication) {
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
		upstreamServiceConfigBuilder.UpstreamAuth = authConfig
	}
	config := upstreamServiceConfigBuilder.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	RegisterServiceViaAPI(t, regClient, req)
	t.Logf("Webrtc Service '%s' registration request sent via API: %s", serviceID, baseURL)
}

// RegisterStreamableMCPService registers a streamable MCP service (SSE).
//
// t is the t.
// regClient is the regClient.
// serviceID is the serviceID.
// targetURL is the targetURL.
// toolAutoDiscovery is the toolAutoDiscovery.
// authConfig is the authConfig.
func RegisterStreamableMCPService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, targetURL string, toolAutoDiscovery bool, authConfig *configv1.Authentication) {
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
		upstreamServiceConfigBuilder.UpstreamAuth = authConfig
	}
	config := upstreamServiceConfigBuilder.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	RegisterServiceViaAPI(t, regClient, req)
	t.Logf("Streamable MCP HTTP Service '%s' registration request sent via API: URL %s", serviceID, targetURL)
}

// RegisterStdioMCPService registers an MCP service using stdio.
//
// t is the t.
// regClient is the regClient.
// serviceID is the serviceID.
// command is the command.
// toolAutoDiscovery is the toolAutoDiscovery.
func RegisterStdioMCPService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, command string, toolAutoDiscovery bool) {
	t.Helper()
	parts := strings.Fields(command)
	require.True(t, len(parts) > 0, "Command for stdio service cannot be empty")
	commandName := parts[0]
	commandArgs := parts[1:]
	RegisterStdioService(t, regClient, serviceID, commandName, toolAutoDiscovery, commandArgs...)
}

// RegisterGRPCService registers a gRPC service.
//
// t is the t.
// regClient is the regClient.
// serviceID is the serviceID.
// grpcTargetAddress is the grpcTargetAddress.
// authConfig is the authConfig.
func RegisterGRPCService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, grpcTargetAddress string, authConfig *configv1.Authentication) {
	t.Helper()

	upstreamServiceConfigBuilder := configv1.UpstreamServiceConfig_builder{
		Name: &serviceID,
		GrpcService: configv1.GrpcUpstreamService_builder{
			Address:       &grpcTargetAddress,
			UseReflection: proto.Bool(true),
		}.Build(),
	}
	if authConfig != nil {
		upstreamServiceConfigBuilder.UpstreamAuth = authConfig
	}
	config := upstreamServiceConfigBuilder.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	RegisterServiceViaAPI(t, regClient, req)
	t.Logf("gRPC Service '%s' registration request sent via API: target %s", serviceID, grpcTargetAddress)
}

// RegisterStdioService registers a raw stdio service.
//
// t is the t.
// regClient is the regClient.
// serviceID is the serviceID.
// commandName is the commandName.
// toolAutoDiscovery is the toolAutoDiscovery.
// commandArgs is the commandArgs.
func RegisterStdioService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, commandName string, toolAutoDiscovery bool, commandArgs ...string) {
	t.Helper()
	RegisterStdioServiceWithSetup(t, regClient, serviceID, commandName, toolAutoDiscovery, "", "", nil, nil, commandArgs...)
}

// RegisterStdioServiceWithSetup registers a stdio service with setup steps.
//
// t is the t.
// regClient is the regClient.
// serviceID is the serviceID.
// commandName is the commandName.
// toolAutoDiscovery is the toolAutoDiscovery.
// workingDir is the workingDir.
// containerImage is the containerImage.
// setupCommands is the setupCommands.
// env is the env.
// commandArgs is the commandArgs.
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
//
// t is the t.
// regClient is the regClient.
// serviceID is the serviceID.
// openAPISpecPath is the openAPISpecPath.
// serverURLOverride is the serverURLOverride.
// authConfig is the authConfig.
func RegisterOpenAPIService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, openAPISpecPath, serverURLOverride string, authConfig *configv1.Authentication) {
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
		upstreamServiceConfigBuilder.UpstreamAuth = authConfig
	}
	config := upstreamServiceConfigBuilder.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	RegisterServiceViaAPI(t, regClient, req)
	t.Logf("OpenAPI Service '%s' registration request sent via API (spec: %s, intended server: %s)", serviceID, absSpecPath, serverURLOverride)
}

// RegisterHTTPServiceWithJSONRPC registers an HTTP service using the JSON-RPC endpoint.
//
// t is the t.
// mcpanyEndpoint is the mcpanyEndpoint.
// serviceID is the serviceID.
// baseURL is the baseURL.
// operationID is the operationID.
// endpointPath is the endpointPath.
// httpMethod is the httpMethod.
// authConfig is the authConfig.
func RegisterHTTPServiceWithJSONRPC(t *testing.T, mcpanyEndpoint, serviceID, baseURL, operationID, endpointPath, httpMethod string, authConfig *configv1.Authentication) {
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
		upstreamServiceConfigBuilder.UpstreamAuth = authConfig
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

	httpReq, err := http.NewRequestWithContext(context.Background(), "POST", mcpanyEndpoint, bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(httpReq)
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

// WaitForPortFromLogs waits for a log line indicating the server is listening and extracts the address.
//
// t is the t.
// mp is the mp.
// serverName is the serverName.
//
// Returns the result.
// Returns an error if the operation fails.
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

// MCPJSONRPCError represents a JSON-RPC error.
type MCPJSONRPCError struct {
	// Code is the error code.
	Code int `json:"code"`
	// Message is the error message.
	Message string `json:"message"`
	// Data is optional additional error data.
	Data interface{} `json:"data,omitempty"`
}

// Error implements the error interface.
//
// Returns the result.
func (e *MCPJSONRPCError) Error() string {
	return fmt.Sprintf("JSON-RPC Error: Code=%d, Message=%s, Data=%v", e.Code, e.Message, e.Data)
}
