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
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestHandleServiceValidate_Filesystem(t *testing.T) {
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
			fsSvc := configv1.FilesystemUpstreamService_builder{
				RootPaths: map[string]string{"/": tt.path},
			}.Build()

			svc := configv1.UpstreamServiceConfig_builder{
				Name:              proto.String("fs-service"),
				FilesystemService: fsSvc,
			}.Build()
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
			cmdSvc := configv1.CommandLineUpstreamService_builder{
				Command:          proto.String(tt.command),
				WorkingDirectory: proto.String(tt.workDir),
			}.Build()

			svc := configv1.UpstreamServiceConfig_builder{
				Name:               proto.String("cmd-service"),
				CommandLineService: cmdSvc,
			}.Build()
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

func TestHandleServiceValidate_DeepCheck(t *testing.T) {
	// Initialize Application with Factory
	app := &Application{
		UpstreamFactory: factory.NewUpstreamServiceFactory(pool.NewManager(), nil),
	}

	// We use a command that exists (passes static check) but fails at runtime (fails deep check).
	cmd := "ls"
	args := []string{"/nonexistent_path_for_test"}
	if os.PathSeparator == '\\' {
		cmd = "cmd"
		args = []string{"/c", "dir", "/nonexistent_path_for_test"}
	}

	stdio := configv1.McpStdioConnection_builder{
		Command: proto.String(cmd),
		Args:    args,
	}.Build()

	mcpSvc := configv1.McpUpstreamService_builder{
		StdioConnection: stdio,
	}.Build()

	svc := configv1.UpstreamServiceConfig_builder{
		Name:       proto.String("deep-fail-service"),
		McpService: mcpSvc,
	}.Build()

	body, _ := protojson.Marshal(svc)
	// Add ?check_connection=true
	req := httptest.NewRequest(http.MethodPost, "/services/validate?check_connection=true", bytes.NewReader(body))
	w := httptest.NewRecorder()

	app.handleServiceValidate().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Assert failure (this is what we want)
	assert.Contains(t, w.Body.String(), `"valid":false`)
	assert.Contains(t, w.Body.String(), "process exited")
}
