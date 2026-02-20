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

// TestGzipLargeWrite verifies that large writes bypass the buffer correctly.
// Although we cannot easily assert on internal buffer usage, we can verify that the output is correct.
func TestGzipLargeWrite(t *testing.T) {
	// A payload larger than minSize (1400) to trigger potential bypass logic.
	largePayload := strings.Repeat("A", 2000)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		// Write the large payload in one go.
		// If optimization works, this should write directly to gzip writer.
		n, err := w.Write([]byte(largePayload))
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		if n != len(largePayload) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(largePayload), n)
		}
	})

	gzipHandler := GzipCompressionMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	gzipHandler.ServeHTTP(rec, req)

	// Verify headers
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding: gzip, got %s", rec.Header().Get("Content-Encoding"))
	}

	// Verify body content
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
		t.Errorf("Body content mismatch. Got length %d, expected %d", len(body), len(largePayload))
	}
}

// TestGzipMultipleWrites verifies that mixed writes (small then large) work correctly.
func TestGzipMixedWrites(t *testing.T) {
	part1 := "Small" // < 1400
	part2 := strings.Repeat("B", 2000) // > 1400

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(part1))
		// At this point, part1 is buffered.
		// Writing part2 should flush buffer + write part2.
		w.Write([]byte(part2))
	})

	gzipHandler := GzipCompressionMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	gzipHandler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding: gzip, got %s", rec.Header().Get("Content-Encoding"))
	}

	reader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read gzip body: %v", err)
	}

	expected := part1 + part2
	if string(body) != expected {
		t.Errorf("Body content mismatch. Got length %d, expected %d", len(body), len(expected))
	}
}

// TestGzipLargeWriteSniff verifies that Content-Type is sniffed correctly when bypassing buffer.
func TestGzipLargeWriteSniff(t *testing.T) {
	// HTML content to trigger text/html detection
	largeHtml := "<!DOCTYPE html><html><body>" + strings.Repeat("A", 2000) + "</body></html>"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Deliberately NOT setting Content-Type
		w.Write([]byte(largeHtml))
	})

	gzipHandler := GzipCompressionMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	gzipHandler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding: gzip")
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("Expected Content-Type to start with text/html, got %s", contentType)
	}
}
