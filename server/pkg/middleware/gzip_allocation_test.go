// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzipAllocationBypass_ContentTypeDetection_HTML(t *testing.T) {
	// Large HTML payload > minSize (1400)
	largeHTML := `<!DOCTYPE html><html><body>` + strings.Repeat("A", 2000) + `</body></html>`

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do not set Content-Type explicitly
		w.Write([]byte(largeHTML))
	})

	gzipHandler := GzipCompressionMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	gzipHandler.ServeHTTP(rec, req)

	// Verify Content-Encoding is gzip
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding: gzip, got %s", rec.Header().Get("Content-Encoding"))
	}

	// Verify Content-Type was detected correctly
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Expected Content-Type to contain text/html, got %s", ct)
	}

	// Verify body correctness
	reader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read gzip body: %v", err)
	}

	if string(body) != largeHTML {
		t.Errorf("Body mismatch")
	}
}
