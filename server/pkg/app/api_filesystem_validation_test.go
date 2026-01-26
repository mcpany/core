// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHandleServiceValidate_FilesystemPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows")
	}

	app := &Application{}
	tmpDir := t.TempDir()

	// Create a directory with no permissions
	lockedDir := filepath.Join(tmpDir, "locked")
	err := os.Mkdir(lockedDir, 0000)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedValid  bool
	}{
		{
			name:           "Locked Directory",
			path:           lockedDir,
			expectedStatus: http.StatusOK,
			expectedValid:  false, // Should be false if we check permissions correctly
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
			}
		})
	}
}
