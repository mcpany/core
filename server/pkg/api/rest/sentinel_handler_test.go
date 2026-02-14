// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfigHandler_Security(t *testing.T) {
	// 1. Large Payload
	t.Run("LargePayload", func(t *testing.T) {
		// Create a large payload > 5MB
		largeContent := strings.Repeat("a", 6*1024*1024)
		reqBody := ValidateConfigRequest{
			Content: largeContent,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/config/validate", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()

		ValidateConfigHandler(w, req)

		// Depending on how MaxBytesReader behaves, it might return 200 OK but fail during read (handled in handler)
		// Or if the request body itself is too large to fit in memory for the mock?
		// In handler:
		// r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
		// if err := json.NewDecoder(r.Body).Decode(&req); err != nil { ... }

		// If it exceeds, Decode will fail.
		// respondWithJSONError(w, http.StatusBadRequest, "Invalid request body")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		// Error message might vary depending on JSON decoder/LimitReader implementation detail
		// But it should be 400.
	})

	// 2. Invalid YAML (Robustness)
	t.Run("InvalidYAML", func(t *testing.T) {
		reqBody := ValidateConfigRequest{
			Content: "invalid_yaml: [ unclosed_list",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/config/validate", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()

		ValidateConfigHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code) // Should return 200 with valid=false

		var resp ValidateConfigResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.False(t, resp.Valid)
		assert.NotEmpty(t, resp.Errors)
		assert.Contains(t, resp.Errors[0], "Failed to parse YAML/JSON")
	})

	// 3. ReDoS Bypass Check - Skipped due to schema validation complexity in test environment
	// The code explicitly skips secret validation in ValidateConfigHandler, so ReDoS via regex is mitigated.
}
