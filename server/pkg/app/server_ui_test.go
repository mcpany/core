// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestApp(t *testing.T) *Application {
	return NewApplication()
}

func TestRun_UI(t *testing.T) {
	// Create temporary ui directory in current package dir
	err := os.Mkdir("ui", 0755)
	if err != nil && !os.IsExist(err) {
		t.Fatalf("Failed to create ui dir: %v", err)
	}
	defer os.RemoveAll("ui")

	err = os.WriteFile("ui/index.html", []byte("<html><body>Hello</body></html>"), 0644)
	require.NoError(t, err)

	// Capture logs to find the dynamic port
	logging.ForTestsOnlyResetLogger()
	var buf ThreadSafeBuffer
	logging.Init(slog.LevelInfo, &buf)

	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := newTestApp(t)

	errChan := make(chan error, 1)

	go func() {
		// Run server with dynamic port
		errChan <- app.Run(ctx, fs, false, "localhost:0", "localhost:0", nil, "", 5*time.Second)
	}()

	// Wait for server to start and retrieve port from logs
	var baseURL string
	require.Eventually(t, func() bool {
		logs := buf.String()
		if strings.Contains(logs, "HTTP server listening") {
			// Extract port from logs: port=127.0.0.1:xxxxx or port="[::1]:xxxxx"
			// Regex to find port key-value pair
			re := regexp.MustCompile(`port=["']?([^"'\s]+)["']?`)
			lines := strings.Split(logs, "\n")
			for _, line := range lines {
				if strings.Contains(line, "HTTP server listening") && strings.Contains(line, "MCP Any HTTP") {
					portMatch := re.FindStringSubmatch(line)
					if len(portMatch) > 1 {
						baseURL = "http://" + portMatch[1]
						return true
					}
				}
			}
		}
		return false
	}, 10*time.Second, 100*time.Millisecond, "Server fail to start or log port")

	// Verify UI endpoint
	require.NotEmpty(t, baseURL)
	require.Eventually(t, func() bool {
		resp, err := http.Get(baseURL + "/ui/")
		if err != nil {
			return false
		}
		defer func() { _ = resp.Body.Close() }()
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond, "UI endpoint should be reachable")

	cancel()
	err = <-errChan
	assert.NoError(t, err)
}
