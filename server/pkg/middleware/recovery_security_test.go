// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/stretchr/testify/assert"
)

func TestRecoveryMiddleware_Leakage(t *testing.T) {
	// Reset logger to capture output
	logging.ForTestsOnlyResetLogger()

	var buf bytes.Buffer
	// Initialize logger to write to buffer
	logging.Init(slog.LevelInfo, &buf)

	// Create a handler that panics
	panickingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	// Wrap it with RecoveryMiddleware
	handler := RecoveryMiddleware(panickingHandler)

	// Make a request with sensitive query param
	secretValue := "supersecret123"
	req := httptest.NewRequest("GET", "/api/v1/resource?api_key="+secretValue+"&public=value", nil)
	w := httptest.NewRecorder()

	// Serve
	handler.ServeHTTP(w, req)

	// Check output
	logs := buf.String()

	// Expect the secret to be ABSENT
	if strings.Contains(logs, secretValue) {
		t.Fatalf("Expected logs NOT to contain secret value %q, but got: %s", secretValue, logs)
	}

	// Verify the redacted URL is present
	// URL encoding might affect this, but RedactURL returns encoded string if modified.
	// In the URL, `?api_key=...` might be encoded. But `[REDACTED]` is what we put in.
	if !strings.Contains(logs, "api_key=%5BREDACTED%5D") && !strings.Contains(logs, "api_key=[REDACTED]") {
		// RedactURL sets the value to [REDACTED], and then q.Encode() URL-encodes it to %5BREDACTED%5D.
		// Wait, `RedactDSN` explicitly decoded it back for readability. `RedactURL` just calls `q.Encode()`.
		// So it will likely be `%5BREDACTED%5D`.
		// Let's check both just in case.
		t.Fatalf("Expected logs to contain redacted URL 'api_key=%%5BREDACTED%%5D' or 'api_key=[REDACTED]', but got: %s", logs)
	}

    // Also assert we got 500
    assert.Equal(t, http.StatusInternalServerError, w.Code)
}
