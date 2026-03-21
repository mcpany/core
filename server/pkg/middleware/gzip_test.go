// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzipCompressionMiddleware(t *testing.T) {
	// Helper to generate large string (> 1400 bytes)
	largePayload := strings.Repeat("Hello, World! ", 150) // ~2100 bytes

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(largePayload))
	})

	gzipHandler := GzipCompressionMiddleware(handler)

	t.Run("Accept-Encoding: gzip (Large Payload)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		gzipHandler.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") != "gzip" {
			t.Errorf("Expected Content-Encoding: gzip, got %s", rec.Header().Get("Content-Encoding"))
		}

		// Verify content is gzipped
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
			t.Errorf("Expected body length %d, got %d", len(largePayload), len(body))
		}
	})

	t.Run("Accept-Encoding: gzip (Small Payload)", func(t *testing.T) {
		smallPayload := "Small payload < 1400 bytes"
		smallHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(smallPayload))
		})
		gzipSmallHandler := GzipCompressionMiddleware(smallHandler)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		gzipSmallHandler.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Expected no Content-Encoding: gzip for small payload")
		}

		if rec.Body.String() != smallPayload {
			t.Errorf("Expected body %q, got %q", smallPayload, rec.Body.String())
		}

		// Check Content-Length is set automatically by our buffering logic
		if rec.Header().Get("Content-Length") == "" {
			t.Error("Expected Content-Length to be set for small buffered response")
		}
	})

	t.Run("Accept-Encoding: identity", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		// No Accept-Encoding or identity
		rec := httptest.NewRecorder()

		gzipHandler.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Expected no Content-Encoding: gzip")
		}

		if rec.Body.String() != largePayload {
			t.Errorf("Unexpected body content")
		}
	})

	t.Run("Non-compressible Content-Type", func(t *testing.T) {
		imageHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			w.Write([]byte("fake image data"))
		})
		gzipImageHandler := GzipCompressionMiddleware(imageHandler)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		gzipImageHandler.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Expected no Content-Encoding: gzip for image/png")
		}
	})

	t.Run("Empty Response", func(t *testing.T) {
		emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
		gzipEmptyHandler := GzipCompressionMiddleware(emptyHandler)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		gzipEmptyHandler.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Expected no Content-Encoding for empty response")
		}
	})

	t.Run("Flush Support", func(t *testing.T) {
		// Custom mock to capture flush state
		mock := &mockFlusher{
			header: make(http.Header),
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Write([]byte("data: part1\n\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			} else {
				t.Error("ResponseWriter does not implement http.Flusher")
			}
			w.Write([]byte("data: part2\n\n"))
		})

		gzipHandler := GzipCompressionMiddleware(handler)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		gzipHandler.ServeHTTP(mock, req)

		// Check if flushed content was compressed
		if len(mock.flushedSnapshots) == 0 {
			t.Error("Expected Flush to be called")
		} else {
			// First snapshot should have compressed data
			// Note: We can't easily decompress partial gzip stream without framing hack,
			// but we can check if it's not empty and has gzip header
			firstFlush := mock.flushedSnapshots[0]
			if len(firstFlush) == 0 {
				t.Error("Expected data to be written before flush")
			}
			// Check for Gzip magic bytes
			if len(firstFlush) > 2 && firstFlush[0] == 0x1f && firstFlush[1] == 0x8b {
				// Good
			} else {
				t.Errorf("Expected gzip header in flushed data, got: %x", firstFlush)
			}
		}
	})
}

type mockFlusher struct {
	header           http.Header
	body             bytes.Buffer
	flushedSnapshots [][]byte
	code             int
}

func (m *mockFlusher) Header() http.Header {
	return m.header
}

func (m *mockFlusher) Write(b []byte) (int, error) {
	return m.body.Write(b)
}

func (m *mockFlusher) WriteHeader(statusCode int) {
	m.code = statusCode
}

func (m *mockFlusher) Flush() {
	// Snapshot current body
	snapshot := make([]byte, m.body.Len())
	copy(snapshot, m.body.Bytes())
	m.flushedSnapshots = append(m.flushedSnapshots, snapshot)
}
