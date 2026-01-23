// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package e2e_sequential

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/app"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadBug(t *testing.T) {
	// Find a free port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	_ = l.Close()

	// Setup app
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/config.yaml", []byte("global_settings:\n  mcp_listen_address: \""+addr+"\"\nupstream_services: []"), 0644)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application := app.NewApplication()
	errChan := make(chan error, 1)
	go func() {
		errChan <- application.Run(app.RunOptions{
			Ctx:             ctx,
			Fs:              fs,
			Stdio:           false,
			JSONRPCPort:     addr,
			GRPCPort:        "",
			ConfigPaths:     []string{"/config.yaml"},
			APIKey:          "",
			ShutdownTimeout: 5 * time.Second,
		})
	}()

	// Wait for server to start
	require.Eventually(t, func() bool {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			_ = conn.Close()
			return true
		}
		return false
	}, 5*time.Second, 100*time.Millisecond)

	baseURL := fmt.Sprintf("http://%s", addr)

	// Create multipart request with long filename containing multibyte chars
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// "你" is 3 bytes. 100 * 3 = 300 bytes.
	longFilename := strings.Repeat("你", 100) + ".txt"
	part, err := writer.CreateFormFile("file", longFilename)
	require.NoError(t, err)
	_, err = part.Write([]byte("content"))
	require.NoError(t, err)
	writer.Close()

	req, err := http.NewRequest("POST", baseURL+"/upload", &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyStr := string(body)

	// Expected sanitized filename: 85 "你" (255 bytes).
	// .txt is truncated because it exceeds 255 bytes.
	expectedFilename := strings.Repeat("你", 85)

	assert.Contains(t, bodyStr, expectedFilename)
	// Verify it does NOT contain the 86th "你"
	assert.NotContains(t, bodyStr, strings.Repeat("你", 86))

	// Also ensure body string is valid UTF-8
	assert.True(t, strings.ToValidUTF8(bodyStr, "") == bodyStr, "Response body should be valid UTF-8")

	// Shutdown
	cancel()
	err = <-errChan
	assert.NoError(t, err)
}
