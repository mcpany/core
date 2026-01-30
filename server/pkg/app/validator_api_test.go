// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
		expectedError  string
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
				Content: `
upstream_services:
  - name: test
    http_service:
      address: http://localhost:8080
`,
				Format: "yaml",
			},
			expectedStatus: http.StatusOK,
			expectedValid:  true,
		},
		{
			name: "Valid YAML Auto-detect",
			req: ValidateRequest{
				Content: `
upstream_services:
  - name: test
    http_service:
      address: http://localhost:8080
`,
			},
			expectedStatus: http.StatusOK,
			expectedValid:  true,
		},
		{
			name: "Malformed JSON",
			req: ValidateRequest{
				Content: `{"upstream_services": [}`,
				Format:  "json",
			},
			expectedStatus: http.StatusBadRequest,
			expectedValid:  false,
			expectedError:  "Failed to parse json",
		},
		{
			name: "Malformed YAML",
			req: ValidateRequest{
				Content: `
upstream_services:
  - name: test
    http_service:
      address: [unclosed
`,
				Format: "yaml",
			},
			expectedStatus: http.StatusBadRequest,
			expectedValid:  false,
			expectedError:  "Failed to parse YAML",
		},
		{
			name: "Schema Violation (Type Mismatch)",
			req: ValidateRequest{
				Content: `{"upstream_services": [{"name": 123}]}`, // name should be string
			},
			expectedStatus: http.StatusBadRequest,
			expectedValid:  false,
			expectedError:  "schema validation failed",
		},
		{
			name: "Deep Validation Error (Invalid Scheme)",
			req: ValidateRequest{
				Content: `{"upstream_services": [{"name": "test", "http_service": {"address": "ftp://localhost"}}]}`,
			},
			expectedStatus: http.StatusBadRequest,
			expectedValid:  false,
			expectedError:  "invalid http address scheme",
		},
		{
			name: "Deep Validation Error (Duplicate Service Name)",
			req: ValidateRequest{
				Content: `{"upstream_services": [{"name": "test", "http_service": {"address": "http://localhost:8080"}}, {"name": "test", "http_service": {"address": "http://localhost:8081"}}]}`,
			},
			expectedStatus: http.StatusBadRequest,
			expectedValid:  false,
			expectedError:  "duplicate service name",
		},
		{
			name: "Empty Content",
			req: ValidateRequest{
				Content: "",
			},
			expectedStatus: http.StatusBadRequest, // Handler returns 400 for empty content
			expectedValid:  false,
			expectedError:  "content is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.req)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
			w := httptest.NewRecorder()

			app.handleValidate().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			// If status is OK, verify the response body JSON
			if w.Code == http.StatusOK {
				var resp ValidateResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedValid, resp.Valid)

				if !tt.expectedValid && tt.expectedError != "" {
					assert.Contains(t, resp.Error, tt.expectedError)
				}
			} else if tt.expectedError != "" {
				// For non-200 responses, error might be plain text
				assert.Contains(t, strings.TrimSpace(w.Body.String()), tt.expectedError)
			}
		})
	}
}

func TestHandleValidate_LargePayload(t *testing.T) {
	app := &Application{}

	// Generate a payload slightly larger than 1MB
	largeContent := strings.Repeat("a", 1048576+100)
	reqData := ValidateRequest{
		Content: largeContent,
		Format:  "json",
	}
	body, err := json.Marshal(reqData)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
	w := httptest.NewRecorder()

	app.handleValidate().ServeHTTP(w, req)

	// Expect 413 Payload Too Large or 400 Bad Request.
	// readBodyWithLimit returns 413 if MaxBytesError is encountered, otherwise 400.
	if w.Code != http.StatusRequestEntityTooLarge && w.Code != http.StatusBadRequest {
		t.Errorf("Expected 413 or 400, got %d", w.Code)
	}
}
