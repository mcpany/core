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

// TestGzipBypassOptimization verifies that large payloads which trigger the buffer bypass
// still result in correctly gzipped output.
func TestGzipBypassOptimization(t *testing.T) {
	// Create a payload larger than minSize (1400) to trigger potential bypass
	largePayload := strings.Repeat("BypassBuffer ", 200) // ~2600 bytes

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		// Perform a single large write.
		// If the optimization is working, this should bypass the internal buffer
		// but still result in a valid gzip stream.
		n, err := w.Write([]byte(largePayload))
		if err != nil {
			t.Errorf("Write failed: %v", err)
		}
		if n != len(largePayload) {
			t.Errorf("Short write: got %d, expected %d", n, len(largePayload))
		}
	})

	gzipHandler := GzipCompressionMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	gzipHandler.ServeHTTP(rec, req)

	// Check headers
	if ce := rec.Header().Get("Content-Encoding"); ce != "gzip" {
		t.Errorf("Expected Content-Encoding: gzip, got %q", ce)
	}

	// Verify content
	reader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read gzip body: %v", err)
	}

	if string(body) != largePayload {
		t.Errorf("Body mismatch. Got length %d, expected %d", len(body), len(largePayload))
	}
}

// TestGzipBypassSniffing verifies that Content-Type sniffing works correctly
// when the buffer is bypassed.
func TestGzipBypassSniffing(t *testing.T) {
	// HTML content to trigger text/html detection
	htmlPayload := strings.Repeat("<html><body><h1>Hello World</h1></body></html>", 50) // ~2400 bytes

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do NOT set Content-Type header explicitly
		// w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(htmlPayload))
	})

	gzipHandler := GzipCompressionMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	gzipHandler.ServeHTTP(rec, req)

	// Check Content-Type was sniffed correctly
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Expected Content-Type to start with text/html, got %q", ct)
	}

	// Check Content-Encoding
	if ce := rec.Header().Get("Content-Encoding"); ce != "gzip" {
		t.Errorf("Expected Content-Encoding: gzip, got %q", ce)
	}

	// Verify content
	reader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read gzip body: %v", err)
	}

	if string(body) != htmlPayload {
		t.Errorf("Body mismatch. Got length %d, expected %d", len(body), len(htmlPayload))
	}
}
