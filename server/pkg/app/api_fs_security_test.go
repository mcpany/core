// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestServiceValidateFilesystemLeak(t *testing.T) {
	app := NewApplication()
	handler := app.handleServiceValidate()

	tests := []struct {
		name        string
		path        string
		expectValid bool
		expectError string
        expectCode  int
	}{
		{
			name:        "Sensitive File Exists (/etc/hosts)",
			path:        "/etc/hosts",
			expectValid: false,
            expectError: "not allowed",
            expectCode:  http.StatusBadRequest,
		},
		{
			name:        "Non-existent File (Absolute)",
			path:        "/non_existent_file_12345",
			expectValid: false,
			expectError: "not allowed",
            expectCode:  http.StatusBadRequest,
		},
        {
            name: "Non-existent File (Relative - Allowed but Missing)",
            path: "missing_in_cwd.txt", // Relative to CWD, so Allowed, but missing
            expectValid: false,
            expectError: "does not exist",
            expectCode:  http.StatusBadRequest,
        },
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fsSvc := &configv1.FilesystemUpstreamService{}
			fsSvc.SetRootPaths(map[string]string{"/virtual": tc.path})

			svc := &configv1.UpstreamServiceConfig{}
			svc.SetName("test-fs-service")
			svc.SetFilesystemService(fsSvc)

			body, _ := protojson.Marshal(svc)
			req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

            if tc.expectCode != 0 {
			    assert.Equal(t, tc.expectCode, rec.Code)
            } else {
                assert.Equal(t, http.StatusOK, rec.Code)
            }

			var resp map[string]any
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tc.expectValid {
				assert.True(t, resp["valid"].(bool), "Expected valid=true")
			} else {
                if val, ok := resp["valid"]; ok {
				    assert.False(t, val.(bool), "Expected valid=false")
                }
				if tc.expectError != "" {
					assert.Contains(t, resp["error"], tc.expectError)
				}
			}
		})
	}
}
