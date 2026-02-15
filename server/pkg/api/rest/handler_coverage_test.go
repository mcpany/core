package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateConfigHandler_Coverage(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedValid  bool
		expectedErrors []string // Strings that should appear in the errors list
		missingErrors  []string // Strings that should NOT appear in the errors list
	}{
		{
			name: "Duplicate Service Names",
			body: `{"content": "upstream_services:\n  - name: service-1\n    http_service:\n      address: http://example.com/1\n  - name: service-1\n    http_service:\n      address: http://example.com/2"}`,
			expectedStatus: http.StatusOK,
			expectedValid:  false,
			expectedErrors: []string{"duplicate service name found"},
		},
		{
			name: "Invalid HTTP Address Scheme",
			body: `{"content": "upstream_services:\n  - name: service-1\n    http_service:\n      address: ftp://example.com/1"}`,
			expectedStatus: http.StatusOK,
			expectedValid:  false,
			expectedErrors: []string{"invalid http address scheme: ftp"},
		},
		{
			name: "Invalid Global Settings (Bind Address)",
			body: `{"content": "global_settings:\n  mcp_listen_address: 999.999.999.999"}`,
			expectedStatus: http.StatusOK,
			expectedValid:  false,
			expectedErrors: []string{"invalid mcp_listen_address"},
		},
		{
			name: "Security Bypass - Missing File (Ignored)",
			body: `{"content": "upstream_services:\n  - name: cmd-service\n    command_line_service:\n      command: /this/file/does/not/exist/surely"}`,
			expectedStatus: http.StatusOK,
			expectedValid:  true,
			missingErrors:  []string{"executable not found", "file not found"},
		},
		{
			name: "Security Bypass - Missing Env Var (Ignored)",
			// Note: We use environmentVariable (camelCase) to pass JSON schema validation,
			// even though YAML unmarshal might handle snake_case.
			body: `{"content": "upstream_services:\n  - name: auth-service\n    http_service:\n      address: http://example.com\n    upstream_auth:\n      api_key:\n        param_name: X-Key\n        value:\n          environmentVariable: THIS_ENV_VAR_MISSING"}`,
			expectedStatus: http.StatusOK,
			expectedValid:  true,
			missingErrors:  []string{"environment variable \"THIS_ENV_VAR_MISSING\" is not set"},
		},
		{
			name: "JSON Schema Validation - Type Mismatch",
			// 'name' expects a string, we give it an integer (in YAML it might be parsed as int or string depending on quotes, but usually int if no quotes)
			// Actually, name: 123 is valid YAML, but schema might require string.
			// However, Go yaml parser might unmarshal it as int, and then JSON schema validation checks the map.
			body: `{"content": "upstream_services:\n  - name: 12345\n    http_service:\n      address: http://example.com"}`,
			expectedStatus: http.StatusOK,
			expectedValid:  false,
			// The error message from jsonschema library depends on implementation.
			// It often says "expected string, but got number" or similar.
			// But wait, the proto definition also says `string name = 1`.
			// `engine.Unmarshal` using protojson/yaml might fail first?
			// `ValidateConfigAgainstSchema` runs first.
			// Let's assume schema validation catches it.
			expectedErrors: []string{"/upstream_services/0/name"},
		},
		{
			name: "Valid Config",
			body: `{"content": "upstream_services:\n  - name: valid-service\n    http_service:\n      address: http://example.com/api"}`,
			expectedStatus: http.StatusOK,
			expectedValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/config/validate", bytes.NewBufferString(tt.body))
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
					t.Errorf("expected valid %v, got %v. Errors: %v", tt.expectedValid, resp.Valid, resp.Errors)
				}

				for _, expectedErr := range tt.expectedErrors {
					found := false
					for _, actualErr := range resp.Errors {
						if strings.Contains(actualErr, expectedErr) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing %q, but got: %v", expectedErr, resp.Errors)
					}
				}

				for _, missingErr := range tt.missingErrors {
					for _, actualErr := range resp.Errors {
						if strings.Contains(actualErr, missingErr) {
							t.Errorf("did NOT expect error containing %q, but got it: %s", missingErr, actualErr)
						}
					}
				}
			}
		})
	}
}
