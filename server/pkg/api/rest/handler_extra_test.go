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

func (f *failWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write failed")
}

func TestValidateConfigHandler_WriteFailure(t *testing.T) {
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

func TestValidateConfigHandler_RespondWithValidationErrors_WriteFailure(t *testing.T) {
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
