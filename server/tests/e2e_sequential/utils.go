package e2e_sequential

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func verifyEndpoint(t *testing.T, url string, expectedStatus int, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == expectedStatus {
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("Failed to verify endpoint %s within %v", url, timeout)
}

func StartServer(t *testing.T, rootDir string, configPath string, cfgContent string) (string, *exec.Cmd) {
	// Use port 0 for dynamic allocation
	// Ensure config has 127.0.0.1:0
	if !strings.Contains(cfgContent, "127.0.0.1:0") {
		// Replace any existing port with 0 if needed, or assume caller handled it.
		// But here we enforce dynamic port for reliability.
		// The template config1 has 127.0.0.1:0
	}

	err := os.WriteFile(configPath, []byte(cfgContent), 0644)
	require.NoError(t, err)

	serverBin := filepath.Join(rootDir, "build/bin/server")
	c := exec.Command(serverBin, "run", "--config-path", configPath, "--debug", "--api-key", "test-key")
	c.Env = os.Environ()

	// Capture stdout to find port
	stdout, err := c.StdoutPipe()
	require.NoError(t, err)
	c.Stderr = os.Stderr // Keep stderr on console

	err = c.Start()
	require.NoError(t, err)

	// Scan for port
	portChan := make(chan string)
	go func() {
		defer close(portChan)
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				chunk := string(buf[:n])
				// Look for port=127.0.0.1:XXXX or port=:XXXX
				// Log format: msg="HTTP server listening" ... port=127.0.0.1:53211
				if idx := strings.Index(chunk, "port="); idx != -1 {
					rest := chunk[idx+5:]
					if end := strings.IndexAny(rest, " \n\t"); end != -1 {
						addr := rest[:end]
						if strings.Contains(addr, ":") {
							_, p, _ := net.SplitHostPort(addr)
							portChan <- p
							return
						}
					}
				}
				// Also print to stdout for debugging
				fmt.Print(chunk)
			}
			if err != nil {
				return
			}
		}
	}()

	select {
	case p := <-portChan:
		if p == "" {
			t.Fatal("Failed to parse port from server logs")
		}
		return fmt.Sprintf("http://127.0.0.1:%s", p), c
	case <-time.After(10 * time.Second):
		t.Fatal("Timed out waiting for server start")
		return "", nil
	}
}
