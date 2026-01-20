// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleValidate(t *testing.T) {
	app := &Application{}

	tests := []struct {
		name           string
		req            ValidateRequest
		expectedStatus int
		expectedValid  bool
		expectedMsg    string
	}{
		{
			name: "Valid JSON",
			req: ValidateRequest{
				Content: `{"upstream_services": [{"name": "test", "http_service": {"address": "http://localhost:8080"}}]}`,
				Format:  "json",
			},
			expectedStatus: http.StatusOK,
			expectedValid:  true,
		},
		{
			name: "Valid YAML",
			req: ValidateRequest{
				Content: "upstream_services:\n  - name: test\n    http_service:\n      address: http://localhost:8080",
				Format:  "yaml",
			},
			expectedStatus: http.StatusOK,
			expectedValid:  true,
		},
		{
			name: "Invalid JSON Syntax",
			req: ValidateRequest{
				Content: `{"upstream_services": [{"name": "test"`,
				Format:  "json",
			},
			expectedStatus: http.StatusBadRequest,
			expectedValid:  false,
			expectedMsg:    "Failed to parse json",
		},
		{
			name: "Invalid Schema (Missing required field)",
			req: ValidateRequest{
				Content: `{"upstream_services": [{"http_service": {"address": "http://localhost:8080"}}]}`,
				Format:  "json",
			},
			expectedStatus: http.StatusBadRequest,
			expectedValid:  false,
			expectedMsg:    `Deep validation failed: service "": service name is empty`,
		},
		{
			name: "Auto Detect JSON",
			req: ValidateRequest{
				Content: `{"upstream_services": [{"name": "test", "http_service": {"address": "http://localhost:8080"}}]}`,
			},
			expectedStatus: http.StatusOK,
			expectedValid:  true,
		},
		{
			name: "Empty Content",
			req: ValidateRequest{
				Content: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.req)
			req, err := http.NewRequest("POST", "/api/v1/validate", bytes.NewBuffer(body))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler := app.handleValidate()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus != http.StatusBadRequest || tt.req.Content != "" {
				var resp ValidateResponse
				err = json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedValid, resp.Valid)
				if tt.expectedMsg != "" {
					assert.Contains(t, resp.Message, tt.expectedMsg)
				}
			}
		})
	}
}
