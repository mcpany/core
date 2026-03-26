// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type failWriter struct {
	http.ResponseWriter
}

// Write ...
// Summary: Write
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return 0, errors.New("write failed")
}

// TestValidateConfigHandler_WriteFailure ...
// Summary: TestValidateConfigHandler_WriteFailure
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Create a valid request body
	body := `{"content": "global_settings:\n  mcp_listen_address: :8080"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/config/validate", strings.NewReader(body))

	w := httptest.NewRecorder()
	fw := &failWriter{ResponseWriter: w}

	ValidateConfigHandler(fw, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

// TestValidateConfigHandler_RespondWithValidationErrors_WriteFailure ...
// Summary: TestValidateConfigHandler_RespondWithValidationErrors_WriteFailure
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Trigger validation error by passing invalid YAML
	body := `{"content": ": invalid yaml"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/config/validate", strings.NewReader(body))

	w := httptest.NewRecorder()
	fw := &failWriter{ResponseWriter: w}

	ValidateConfigHandler(fw, req)

	// respondWithValidationErrors should be called, which calls Encode, which fails, which calls http.Error -> 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}
