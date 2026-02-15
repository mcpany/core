package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestValidateConfigHandler_FilesystemLeak(t *testing.T) {
	// Create a dummy executable file to test with
	tmpFile, err := os.CreateTemp("", "dummy_exec")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// Make it executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		t.Fatal(err)
	}

	// We'll use this existing file as a target
	existingPath := tmpFile.Name()
	nonExistentPath := "/this/path/definitely/does/not/exist/at/all/12345"

	tests := []struct {
		name          string
		command       string
		expectedValid bool
		expectError   bool
	}{
		{
			name:          "Existing File",
			command:       existingPath,
			expectedValid: true, // Should pass validation
			expectError:   false,
		},
		{
			name:          "Non-Existent File",
			command:       nonExistentPath,
			expectedValid: true, // Should ALSO pass validation (check skipped to prevent leak)
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]string{
				"content": "upstream_services:\n  - name: test-service\n    command_line_service:\n      command: " + tt.command,
			}
			jsonBody, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/config/validate", bytes.NewBuffer(jsonBody))
			w := httptest.NewRecorder()

			ValidateConfigHandler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			var resp ValidateConfigResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Valid != tt.expectedValid {
				t.Errorf("expected valid %v, got %v. Errors: %v", tt.expectedValid, resp.Valid, resp.Errors)
			}
		})
	}
}
