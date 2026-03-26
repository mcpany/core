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
			fsSvc := &configv1.FilesystemUpstreamService{}
			fsSvc.SetRootPaths(map[string]string{"/": tt.path})

			svc := &configv1.UpstreamServiceConfig{}
			svc.SetName("fs-service")
			svc.SetFilesystemService(fsSvc)
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
			cmdSvc := &configv1.CommandLineUpstreamService{}
			cmdSvc.SetCommand(tt.command)
			cmdSvc.SetWorkingDirectory(tt.workDir)

			svc := &configv1.UpstreamServiceConfig{}
			svc.SetName("cmd-service")
			svc.SetCommandLineService(cmdSvc)
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
