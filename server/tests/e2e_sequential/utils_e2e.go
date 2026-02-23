//go:build e2e

package e2e_sequential

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func BuildServer(t *testing.T) string {
	t.Helper()
	rootDir := findRootDir(t)
	binDir := filepath.Join(rootDir, "build", "bin")
	binPath := filepath.Join(binDir, "server")

	// Ensure bin dir exists
	err := os.MkdirAll(binDir, 0755)
	require.NoError(t, err)

	// Build
	// We assume we are in the server module root (where go.mod is), so path is ./cmd/server
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/server")
	cmd.Dir = rootDir
	// Output build errors
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build server: %s", string(out))

	return binPath
}

func findRootDir(t *testing.T) string {
	dir, err := os.Getwd()
	require.NoError(t, err)
	// Traverse up to find go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Root dir not found")
		}
		dir = parent
	}
}

// ServerProcess represents a running server process
type ServerProcess struct {
	Cmd     *exec.Cmd
	Port    string
	BaseURL string
	APIKey  string
	WorkDir string
}

func (s *ServerProcess) Stop() {
	if s.Cmd != nil && s.Cmd.Process != nil {
		_ = s.Cmd.Process.Kill()
		_ = s.Cmd.Wait()
	}
	if s.WorkDir != "" {
		_ = os.RemoveAll(s.WorkDir)
	}
}

func StartServerProcess(t *testing.T, binPath string, args ...string) *ServerProcess {
	t.Helper()
	// Create temporary work dir
	workDir, err := os.MkdirTemp("", "mcpany-e2e-*")
	require.NoError(t, err)

	port := getFreePort(t)
	apiKey := "e2e-test-key"

	// Prepare config file if not provided
	// We append flags.
	address := fmt.Sprintf("127.0.0.1:%s", port)
	runArgs := []string{"run", "--mcp-listen-address", address, "--api-key", apiKey, "--db-path", filepath.Join(workDir, "test.db")}
	runArgs = append(runArgs, args...)

	cmd := exec.Command(binPath, runArgs...)
	cmd.Dir = workDir // Run in temp dir

	// Override MCP_LISTEN_ADDRESS to ensure config file doesn't override flag with default
	cmd.Env = append(os.Environ(), "MCP_LISTEN_ADDRESS="+address)

	// Capture output for debugging
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Start()
	require.NoError(t, err)

	sp := &ServerProcess{
		Cmd:     cmd,
		Port:    port,
		BaseURL: fmt.Sprintf("http://127.0.0.1:%s", port),
		APIKey:  apiKey,
		WorkDir: workDir,
	}

	// Wait for health
	require.Eventually(t, func() bool {
		resp, err := http.Get(sp.BaseURL + "/healthz")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == 200
	}, 10*time.Second, 100*time.Millisecond, "Server failed to start. Stdout: %s, Stderr: %s", stdout.String(), stderr.String())

	return sp
}

func getFreePort(t *testing.T) string {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	require.NoError(t, err)
	l, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err)
	defer l.Close()
	return fmt.Sprintf("%d", l.Addr().(*net.TCPAddr).Port)
}

func SeedData(t *testing.T, baseURL, apiKey string, seedContent string) {
	req, err := http.NewRequest("POST", baseURL+"/api/v1/debug/seed", bytes.NewBuffer([]byte(seedContent)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	require.Equal(t, 200, resp.StatusCode, "Seeding failed: %s", string(body))
}

func StartEchoServer(t *testing.T) string {
	t.Helper()
	port := getFreePort(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write(body) // Echo back
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	l, err := net.Listen("tcp", "127.0.0.1:"+port)
	require.NoError(t, err)

	server := &http.Server{Handler: mux}
	go server.Serve(l)

	t.Cleanup(func() {
		server.Close()
	})

	return "http://127.0.0.1:" + port
}
