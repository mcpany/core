/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package integration

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	apiv1 "github.com/mcpxy/core/proto/api/v1"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

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

func ProjectRoot(t *testing.T) string {
	t.Helper()
	root, err := GetProjectRoot()
	require.NoError(t, err)
	return root
}

const (
	McpxyServerStartupTimeout = 30 * time.Second
	ServiceStartupTimeout     = 15 * time.Second
	TestWaitTimeShort         = 60 * time.Second
	TestWaitTimeMedium        = 60 * time.Second
	TestWaitTimeLong          = 5 * time.Minute
	RetryInterval             = 250 * time.Millisecond
	localHeaderMcpSessionID   = "Mcp-Session-Id"
)

// getDockerCommand returns the command and base arguments for running Docker,
// respecting the USE_SUDO_FOR_DOCKER environment variable.
func getDockerCommand() (string, []string) {
	if os.Getenv("USE_SUDO_FOR_DOCKER") == "true" {
		return "sudo", []string{"docker"}
	}
	return "docker", []string{}
}

// --- Binary Paths ---

var (
	projectRoot  string
	findRootOnce sync.Once
)

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
	return projectRoot, err
}

// --- Helper: Find Free Port ---
var portMutex sync.Mutex

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

// --- Process Management for External Services ---
type ManagedProcess struct {
	cmd                 *exec.Cmd
	t                   *testing.T
	wg                  sync.WaitGroup
	stdout              threadSafeBuffer
	stderr              threadSafeBuffer
	waitDone            chan struct{}
	label               string
	IgnoreExitStatusOne bool
}

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

func (mp *ManagedProcess) Cmd() *exec.Cmd {
	return mp.cmd
}

func (mp *ManagedProcess) Start() error {
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

func (mp *ManagedProcess) StdoutString() string { return mp.stdout.String() }
func (mp *ManagedProcess) StderrString() string { return mp.stderr.String() }

// WaitForTCPPort waits for a TCP port to become open and accepting connections.
func WaitForTCPPort(t *testing.T, port int, timeout time.Duration) {
	t.Helper()
	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 100*time.Millisecond)
		if err != nil {
			return false // Port is not open yet
		}
		conn.Close()
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
		defer conn.Close()

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
		defer resp.Body.Close()
		conn.Close()
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
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, timeout, 250*time.Millisecond, "URL %s did not become healthy in time", url)
}

// IsDockerSocketAccessible checks if the Docker daemon is accessible.
func IsDockerSocketAccessible() bool {
	dockerExe, dockerArgs := getDockerCommand()
	cmd := exec.Command(dockerExe, append(dockerArgs, "info")...)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// --- Mock Service Start Helpers (External Processes) ---

func StartDockerContainer(t *testing.T, imageName, containerName string, args ...string) (cleanupFunc func()) {
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
	stopCmd := exec.Command(dockerExe, buildArgs("stop", containerName)...)
	_ = stopCmd.Run() // Ignore error, it might not be running
	rmCmd := exec.Command(dockerExe, buildArgs("rm", containerName)...)
	_ = rmCmd.Run() // Ignore error, it might not exist

	runArgs := []string{"run", "--name", containerName, "--rm"}
	runArgs = append(runArgs, args...)
	runArgs = append(runArgs, imageName)

	startCmd := exec.Command(dockerExe, buildArgs(runArgs...)...)
	startCmd.Stdout = os.Stdout
	startCmd.Stderr = os.Stderr

	err := startCmd.Start()
	require.NoError(t, err, "failed to start docker container %s", imageName)

	cleanupFunc = func() {
		t.Logf("Stopping and removing docker container: %s", containerName)
		stopCleanupCmd := exec.Command(dockerExe, buildArgs("stop", containerName)...)
		err := stopCleanupCmd.Run()
		if err != nil {
			t.Logf("Failed to stop docker container %s: %v", containerName, err)
		}
	}

	// Give the container a moment to initialize
	time.Sleep(3 * time.Second)

	return cleanupFunc
}

// --- MCPXY Server Helper (External Process) ---
type MCPXYTestServerInfo struct {
	Process                  *ManagedProcess
	JSONRPCEndpoint          string
	HTTPEndpoint             string
	GrpcRegistrationEndpoint string
	SessionID                string
	HTTPClient               *http.Client
	GRPCRegConn              *grpc.ClientConn
	RegistrationClient       apiv1.RegistrationServiceClient
	CleanupFunc              func()
	T                        *testing.T
}

func StartMCPXYServer(t *testing.T, testName string, extraArgs ...string) *MCPXYTestServerInfo {
	t.Helper()

	root, err := GetProjectRoot()
	require.NoError(t, err, "Failed to get project root")
	mcpxyBinary := filepath.Join(root, "build/bin/server")

	t.Logf("Using MCPXY binary from: %s", mcpxyBinary)

	jsonrpcPort := FindFreePort(t)
	grpcRegPort := FindFreePort(t)
	for grpcRegPort == jsonrpcPort {
		grpcRegPort = FindFreePort(t)
	}

	args := []string{
		"--jsonrpc-port", fmt.Sprintf("%d", jsonrpcPort),
		"--grpc-port", fmt.Sprintf("%d", grpcRegPort),
	}
	args = append(args, extraArgs...)
	env := []string{"MCPXY_LOG_LEVEL=debug"}
	if sudo, ok := os.LookupEnv("USE_SUDO_FOR_DOCKER"); ok {
		env = append(env, "USE_SUDO_FOR_DOCKER="+sudo)
	}

	absMcpxyBinaryPath, err := filepath.Abs(mcpxyBinary)
	require.NoError(t, err, "Failed to get absolute path for MCPXY binary: %s", mcpxyBinary)
	_, err = os.Stat(absMcpxyBinaryPath)
	require.NoError(t, err, "MCPXY binary not found at %s. Run 'make build'.", absMcpxyBinaryPath)

	mcpProcess := NewManagedProcess(t, "MCPXYServer-"+testName, absMcpxyBinaryPath, args, env)
	mcpProcess.cmd.Dir = root
	err = mcpProcess.Start()
	require.NoError(t, err, "Failed to start MCPXY server. Stderr: %s", mcpProcess.StderrString())

	jsonrpcEndpoint := fmt.Sprintf("http://127.0.0.1:%d", jsonrpcPort)
	grpcRegEndpoint := fmt.Sprintf("127.0.0.1:%d", grpcRegPort)

	mcpRequestURL := jsonrpcEndpoint + "/mcp"
	httpClient := &http.Client{Timeout: 2 * time.Second}

	t.Logf("MCPXY server health check target URL: %s", mcpRequestURL)

	var grpcRegConn *grpc.ClientConn
	require.Eventually(t, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var errDial error
		grpcRegConn, errDial = grpc.NewClient(grpcRegEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if errDial != nil {
			t.Logf("MCPXY gRPC registration endpoint at %s not ready: %v", grpcRegEndpoint, errDial)
			return false
		}
		state := grpcRegConn.GetState()
		if state == connectivity.Ready || state == connectivity.Idle {
			t.Logf("Successfully connected to MCPXY gRPC registration endpoint at %s with state %s", grpcRegEndpoint, state)
			return true
		}
		if !grpcRegConn.WaitForStateChange(ctx, state) {
			t.Logf("MCPXY gRPC registration endpoint at %s did not transition from %s", grpcRegEndpoint, state)
			grpcRegConn.Close()
			return false
		}
		t.Logf("Successfully connected to MCPXY gRPC registration endpoint at %s", grpcRegEndpoint)
		return true
	}, McpxyServerStartupTimeout, RetryInterval, "MCPXY gRPC registration endpoint at %s did not become healthy in time.\nFinal Stdout: %s\nFinal Stderr: %s", grpcRegEndpoint, mcpProcess.StdoutString(), mcpProcess.StderrString())

	registrationClient := apiv1.NewRegistrationServiceClient(grpcRegConn)

	// Wait for the HTTP/JSON-RPC endpoint to be ready
	require.Eventually(t, func() bool {
		// Use a short timeout for the health check itself
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, "GET", mcpRequestURL, nil)
		if err != nil {
			t.Logf("Failed to create request for health check: %v", err)
			return false
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			t.Logf("MCPXY HTTP endpoint at %s not ready: %v", mcpRequestURL, err)
			return false
		}
		defer resp.Body.Close()
		// Any response (even an error like 405 Method Not Allowed) indicates the server is up and listening.
		t.Logf("MCPXY HTTP endpoint at %s is ready (status: %s)", mcpRequestURL, resp.Status)
		return true
	}, McpxyServerStartupTimeout, RetryInterval, "MCPXY HTTP endpoint at %s did not become healthy in time.\nFinal Stdout: %s\nFinal Stderr: %s", mcpRequestURL, mcpProcess.StdoutString(), mcpProcess.StderrString())

	t.Logf("MCPXY Server process started. MCP Endpoint Base: %s, gRPC Reg: %s", jsonrpcEndpoint, grpcRegEndpoint)

	return &MCPXYTestServerInfo{
		Process:                  mcpProcess,
		JSONRPCEndpoint:          jsonrpcEndpoint,
		HTTPEndpoint:             mcpRequestURL,
		GrpcRegistrationEndpoint: grpcRegEndpoint,
		HTTPClient:               httpClient,
		GRPCRegConn:              grpcRegConn,
		RegistrationClient:       registrationClient,
		CleanupFunc: func() {
			t.Logf("Cleaning up MCPXYTestServerInfo for %s...", testName)
			if grpcRegConn != nil {
				grpcRegConn.Close()
			}
			mcpProcess.Stop()
		},
		T: t,
	}
}

func RegisterServiceViaAPI(t *testing.T, regClient apiv1.RegistrationServiceClient, req *apiv1.RegisterServiceRequest) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), TestWaitTimeShort)
	defer cancel()
	resp, err := regClient.RegisterService(ctx, req)
	require.NoError(t, err, "Failed to register service %s via API.", req.GetConfig().GetName())
	require.NotNil(t, resp, "Nil response from RegisterService API for %s", req.GetConfig().GetName())
	t.Logf("Service %s registered via API successfully. Message: %s, Discovered tools:\n%v", req.GetConfig().GetName(), resp.GetMessage(), resp.GetDiscoveredTools())
}

func RegisterHTTPService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, baseURL, operationID, endpointPath, httpMethod string, authConfig *configv1.UpstreamAuthentication) {
	t.Helper()
	RegisterHTTPServiceWithParams(t, regClient, serviceID, baseURL, operationID, endpointPath, httpMethod, nil, authConfig)
}

func RegisterHTTPServiceWithParams(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, baseURL, operationID, endpointPath, httpMethod string, params []*configv1.HttpParameterMapping, authConfig *configv1.UpstreamAuthentication) {
	t.Helper()
	t.Logf("Registering HTTP service '%s' with endpoint path: %s", serviceID, endpointPath)

	httpMethodEnumName := "HTTP_METHOD_" + strings.ToUpper(httpMethod)
	if _, ok := configv1.HttpCallDefinition_HttpMethod_value[httpMethodEnumName]; !ok {
		t.Fatalf("Invalid HTTP method provided: %s", httpMethod)
	}
	method := configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value[httpMethodEnumName])

	callDef := configv1.HttpCallDefinition_builder{
		EndpointPath:      proto.String(endpointPath),
		OperationId:       proto.String(operationID),
		Method:            &method,
		ParameterMappings: params,
	}.Build()

	configBuilder := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(serviceID),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(baseURL),
			Calls:   []*configv1.HttpCallDefinition{callDef},
		}.Build(),
	}

	if authConfig != nil {
		configBuilder.UpstreamAuthentication = authConfig
	}

	req := apiv1.RegisterServiceRequest_builder{
		Config: configBuilder.Build(),
	}.Build()

	RegisterServiceViaAPI(t, regClient, req)
	t.Logf("HTTP Service '%s' registration request sent via API: %s %s%s", serviceID, httpMethod, baseURL, endpointPath)
}

func RegisterWebsocketService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, baseURL, operationID string, authConfig *configv1.UpstreamAuthentication) {
	t.Helper()
	t.Logf("Registering Websocket service '%s' with endpoint: %s", serviceID, baseURL)

	configBuilder := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(serviceID),
		WebsocketService: configv1.WebsocketUpstreamService_builder{
			Address: proto.String(baseURL),
			Calls: []*configv1.WebsocketCallDefinition{
				configv1.WebsocketCallDefinition_builder{
					OperationId: proto.String(operationID),
				}.Build(),
			},
		}.Build(),
	}

	if authConfig != nil {
		configBuilder.UpstreamAuthentication = authConfig
	}

	req := apiv1.RegisterServiceRequest_builder{
		Config: configBuilder.Build(),
	}.Build()

	RegisterServiceViaAPI(t, regClient, req)
	t.Logf("Websocket Service '%s' registration request sent via API: %s", serviceID, baseURL)
}

func RegisterWebrtcService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, baseURL, operationID string, authConfig *configv1.UpstreamAuthentication) {
	t.Helper()
	t.Logf("Registering Webrtc service '%s' with endpoint: %s", serviceID, baseURL)

	configBuilder := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(serviceID),
		WebrtcService: configv1.WebrtcUpstreamService_builder{
			Address: proto.String(baseURL),
			Calls: []*configv1.WebrtcCallDefinition{
				configv1.WebrtcCallDefinition_builder{
					OperationId: proto.String(operationID),
				}.Build(),
			},
		}.Build(),
	}

	if authConfig != nil {
		configBuilder.UpstreamAuthentication = authConfig
	}

	req := apiv1.RegisterServiceRequest_builder{
		Config: configBuilder.Build(),
	}.Build()

	RegisterServiceViaAPI(t, regClient, req)
	t.Logf("Webrtc Service '%s' registration request sent via API: %s", serviceID, baseURL)
}

func RegisterStreamableMCPService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, targetURL string, toolAutoDiscovery bool, authConfig *configv1.UpstreamAuthentication) {
	t.Helper()
	configBuilder := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(serviceID),
		McpService: configv1.McpUpstreamService_builder{
			ToolAutoDiscovery: proto.Bool(toolAutoDiscovery),
			HttpConnection: (&configv1.McpStreamableHttpConnection_builder{
				HttpAddress: proto.String(targetURL),
			}).Build(),
		}.Build(),
	}

	if authConfig != nil {
		configBuilder.UpstreamAuthentication = authConfig
	}

	req := apiv1.RegisterServiceRequest_builder{
		Config: configBuilder.Build(),
	}.Build()
	RegisterServiceViaAPI(t, regClient, req)
	t.Logf("Streamable MCP HTTP Service '%s' registration request sent via API: URL %s", serviceID, targetURL)
}

func RegisterStdioMCPService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, command string, toolAutoDiscovery bool) {
	t.Helper()
	parts := strings.Fields(command)
	require.True(t, len(parts) > 0, "Command for stdio service cannot be empty")
	commandName := parts[0]
	commandArgs := parts[1:]
	RegisterStdioService(t, regClient, serviceID, commandName, toolAutoDiscovery, commandArgs...)
}

func RegisterGRPCService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, grpcTargetAddress string, authConfig *configv1.UpstreamAuthentication) {
	t.Helper()

	configBuilder := configv1.UpstreamServiceConfig_builder{
		Name: &serviceID,
		GrpcService: configv1.GrpcUpstreamService_builder{
			Address:       proto.String(grpcTargetAddress),
			UseReflection: proto.Bool(true),
		}.Build(),
	}

	if authConfig != nil {
		configBuilder.UpstreamAuthentication = authConfig
	}

	req := apiv1.RegisterServiceRequest_builder{
		Config: configBuilder.Build(),
	}.Build()
	RegisterServiceViaAPI(t, regClient, req)
	t.Logf("gRPC Service '%s' registration request sent via API: target %s", serviceID, grpcTargetAddress)
}

func RegisterStdioService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, commandName string, toolAutoDiscovery bool, commandArgs ...string) {
	t.Helper()
	RegisterStdioServiceWithSetup(t, regClient, serviceID, commandName, toolAutoDiscovery, "", "", nil, commandArgs...)
}

func RegisterStdioServiceWithSetup(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, commandName string, toolAutoDiscovery bool, workingDir, containerImage string, setupCommands []string, commandArgs ...string) {
	t.Helper()
	stdioConnection := configv1.McpStdioConnection_builder{
		Command:          &commandName,
		Args:             commandArgs,
		WorkingDirectory: &workingDir,
		ContainerImage:   &containerImage,
		SetupCommands:    setupCommands,
	}.Build()
	mcpService := configv1.McpUpstreamService_builder{
		ToolAutoDiscovery: &toolAutoDiscovery,
		StdioConnection:   stdioConnection,
	}.Build()
	cfg := configv1.UpstreamServiceConfig_builder{
		Name:       &serviceID,
		McpService: mcpService,
	}.Build()
	req := apiv1.RegisterServiceRequest_builder{
		Config: cfg,
	}.Build()
	fullCommand := append([]string{commandName}, commandArgs...)
	t.Logf("Registering stdio service '%s' with command: %v", serviceID, fullCommand)
	time.Sleep(250 * time.Millisecond)
	RegisterServiceViaAPI(t, regClient, req)
}

func RegisterOpenAPIService(t *testing.T, regClient apiv1.RegistrationServiceClient, serviceID, openAPISpecPath, serverURLOverride string, authConfig *configv1.UpstreamAuthentication) {
	t.Helper()
	absSpecPath, err := filepath.Abs(openAPISpecPath)
	require.NoError(t, err)
	_, err = os.Stat(absSpecPath)
	require.NoError(t, err, "OpenAPI spec file not found: %s", absSpecPath)
	specContent, err := os.ReadFile(absSpecPath)
	require.NoError(t, err)
	spec := string(specContent)
	openapiServiceDef := configv1.OpenapiUpstreamService_builder{
		OpenapiSpec: &spec,
		Address:     &serverURLOverride,
	}.Build()
	configBuilder := configv1.UpstreamServiceConfig_builder{
		Name:           &serviceID,
		OpenapiService: openapiServiceDef,
	}

	if authConfig != nil {
		configBuilder.UpstreamAuthentication = authConfig
	}

	req := apiv1.RegisterServiceRequest_builder{
		Config: configBuilder.Build(),
	}.Build()
	RegisterServiceViaAPI(t, regClient, req)
	t.Logf("OpenAPI Service '%s' registration request sent via API (spec: %s, intended server: %s)", serviceID, absSpecPath, serverURLOverride)
}

type MCPJSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *MCPJSONRPCError) Error() string {
	return fmt.Sprintf("JSON-RPC Error: Code=%d, Message=%s, Data=%v", e.Code, e.Message, e.Data)
}
