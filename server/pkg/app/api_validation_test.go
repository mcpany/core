// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestHandleServiceValidate_Filesystem(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_UNSAFE_CONFIG", "true")
	app := &Application{}
	tmpDir := t.TempDir()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedValid  bool
	}{
		{
			name:           "Valid Path",
			path:           tmpDir,
			expectedStatus: http.StatusOK,
			expectedValid:  true,
		},
		{
			name:           "Invalid Path",
			path:           filepath.Join(tmpDir, "nonexistent"),
			expectedStatus: http.StatusOK, // The API returns 200 OK even for validation failure
			expectedValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &configv1.UpstreamServiceConfig{
				Name: proto.String("fs-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
					FilesystemService: &configv1.FilesystemUpstreamService{
						RootPaths: map[string]string{"/": tt.path},
					},
				},
			}
			body, _ := protojson.Marshal(svc)
			req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader(body))
			w := httptest.NewRecorder()
			app.handleServiceValidate().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedValid {
				assert.Contains(t, w.Body.String(), `"valid":true`)
			} else {
				assert.Contains(t, w.Body.String(), `"valid":false`)
				assert.Contains(t, w.Body.String(), "Filesystem path check failed")
			}
		})
	}
}

func TestHandleServiceValidate_CommandLine(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_UNSAFE_CONFIG", "true")
	app := &Application{}

	// Assuming "ls" (or "dir" on windows) exists.
	cmd := "ls"
	if os.PathSeparator == '\\' {
		cmd = "cmd" // approximate for windows
	}

	tests := []struct {
		name           string
		command        string
		workDir        string
		expectedStatus int
		expectedValid  bool
	}{
		{
			name:           "Valid Command",
			command:        cmd,
			expectedStatus: http.StatusOK,
			expectedValid:  true,
		},
		{
			name:           "Invalid Command",
			command:        "nonexistent_command_xyz",
			expectedStatus: http.StatusBadRequest, // Fails static validation in config package
			expectedValid:  false,
		},
		{
			name:           "Invalid Working Directory",
			command:        cmd,
			workDir:        "/nonexistent/dir",
			expectedStatus: http.StatusOK,
			expectedValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &configv1.UpstreamServiceConfig{
				Name: proto.String("cmd-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
					CommandLineService: &configv1.CommandLineUpstreamService{
						Command:          proto.String(tt.command),
						WorkingDirectory: proto.String(tt.workDir),
					},
				},
			}
			body, _ := protojson.Marshal(svc)
			req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader(body))
			w := httptest.NewRecorder()
			app.handleServiceValidate().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedValid {
				assert.Contains(t, w.Body.String(), `"valid":true`)
			} else {
				assert.Contains(t, w.Body.String(), `"valid":false`)
			}
		})
	}
}

func TestHandleServiceValidate_McpStdio(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_UNSAFE_CONFIG", "true")
	app := &Application{}

	// Test case: Invalid command (should fail handshake because it doesn't speak MCP)
	// We use "echo hello" which exits immediately or prints "hello" and not JSON-RPC.
	echoCmd := "echo"
	if os.PathSeparator == '\\' {
		echoCmd = "cmd" // cmd /c echo hello ? Too complex for cross platform, just rely on echo usually available or fails command check
	}

	// Case 1: Bad Protocol (Handshake fail)
	t.Run("Handshake Fail", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			Name: proto.String("mcp-service-invalid"),
			ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
				McpService: &configv1.McpUpstreamService{
					ConnectionType: &configv1.McpUpstreamService_StdioConnection{
						StdioConnection: &configv1.McpStdioConnection{
							Command: proto.String(echoCmd),
							Args:    []string{"hello"},
						},
					},
				},
			},
		}
		body, _ := protojson.Marshal(svc)
		req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader(body))
		w := httptest.NewRecorder()
		app.handleServiceValidate().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"valid":false`)
		// It will fail because "echo hello" isn't a valid MCP server.
		// Error might be "process exited" or "decode error".
		// We just check it's invalid and has some error detail.
		assert.Contains(t, w.Body.String(), "error")
	})

	// Case 2: Non-existent Command
	t.Run("Non-existent Command", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			Name: proto.String("mcp-service-missing"),
			ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
				McpService: &configv1.McpUpstreamService{
					ConnectionType: &configv1.McpUpstreamService_StdioConnection{
						StdioConnection: &configv1.McpStdioConnection{
							Command: proto.String("nonexistent_command_xyz_123"),
						},
					},
				},
			},
		}
		body, _ := protojson.Marshal(svc)
		req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader(body))
		w := httptest.NewRecorder()
		app.handleServiceValidate().ServeHTTP(w, req)

		// Static validation might catch this and return 400
		if w.Code == http.StatusBadRequest {
			assert.Contains(t, w.Body.String(), `"valid":false`)
			// Error message from validator might vary but usually indicates failure
		} else {
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), `"valid":false`)
			assert.Contains(t, w.Body.String(), "not found")
		}
	})
}
