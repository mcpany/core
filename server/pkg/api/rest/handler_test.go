package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateConfigHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		expectedValid  bool
		expectError    bool // If we expect validation errors
	}{
		{
			name:           "Valid Method",
			method:         http.MethodPost,
			body:           `{"content": "global_settings:\n  mcp_listen_address: :8080"}`,
			expectedStatus: http.StatusOK,
			expectedValid:  true,
		},
		{
			name:           "Invalid Method",
			method:         http.MethodGet,
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Empty Content",
			method:         http.MethodPost,
			body:           `{"content": ""}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid YAML",
			method:         http.MethodPost,
			body:           `{"content": ": invalid yaml"}`,
			expectedStatus: http.StatusOK,
			expectedValid:  false,
			expectError:    true,
		},
		{
			name:           "Invalid JSON Body",
			method:         http.MethodPost,
			body:           `{"content": "foo",}`, // Invalid JSON
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Request Body Too Large",
			method:         http.MethodPost,
			body:           `{"content": "` + strings.Repeat("a", 6<<20) + `"}`, // 6MB > 5MB limit
			// http.MaxBytesReader causes Read to return error, JSON decoder fails, handler returns 400.
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Semantic Check (Missing Command - Ignored due to Security)",
			method:         http.MethodPost,
			body:           `{"content": "upstream_services:\n  - name: broken-service\n    command_line_service:\n      command: /this/path/definitely/does/not/exist/12345"}`,
			expectedStatus: http.StatusOK,
			expectedValid:  true, // Security: We skip filesystem checks to prevent enumeration
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/config/validate", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()

			ValidateConfigHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp ValidateConfigResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if resp.Valid != tt.expectedValid {
					t.Errorf("expected valid %v, got %v", tt.expectedValid, resp.Valid)
				}

				if tt.expectError && len(resp.Errors) == 0 {
					t.Error("expected errors but got none")
				}
			}
		})
	}
}
