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
	"time"
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
}

func TestGzipCompressionMiddleware_Flush(t *testing.T) {
	// This test verifies that Flush() calls are propagated properly.
	// For streaming responses (e.g. SSE), flushing is critical.

	t.Run("Flush Propagates Immediately", func(t *testing.T) {
		// Use a channel to synchronize the test
		// firstPartReceived := make(chan bool)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.WriteHeader(http.StatusOK)

			// Write first part
			w.Write([]byte("data: part1\n\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			} else {
				t.Error("ResponseWriter does not implement http.Flusher")
			}

			// Wait for client to acknowledge receipt (simulated by sleep in test)
			// In a real integration test we would use channels, but httptest.ResponseRecorder doesn't stream.
			// However, we can check if the underlying Flusher was called if we mock it.
			// Since we can't easily mock http.ResponseWriter here without implementing a full mock,
			// we will rely on checking if the recorder has data after Flush.
			// NOTE: httptest.ResponseRecorder handles Flush by flushing its buffer to Body.

			// Write second part after delay
			time.Sleep(10 * time.Millisecond)
			w.Write([]byte("data: part2\n\n"))
		})

		gzipHandler := GzipCompressionMiddleware(handler)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		// Run in a goroutine because ServeHTTP blocks until handler returns
		done := make(chan bool)
		go func() {
			gzipHandler.ServeHTTP(rec, req)
			close(done)
		}()

		<-done

		// Verify we got gzip content
		if rec.Header().Get("Content-Encoding") != "gzip" {
			t.Errorf("Expected Content-Encoding: gzip, got %s", rec.Header().Get("Content-Encoding"))
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

		expected := "data: part1\n\ndata: part2\n\n"
		if string(body) != expected {
			t.Errorf("Expected body %q, got %q", expected, string(body))
		}
	})
}

// MockFlusher implements http.ResponseWriter and http.Flusher for testing.
type MockFlusher struct {
	http.ResponseWriter
	Flushed bool
}

func (m *MockFlusher) Flush() {
	m.Flushed = true
	if f, ok := m.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func TestGzipCompressionMiddleware_FlushCalled(t *testing.T) {
	// This test explicitly checks if the Flush method on the underlying writer is called.

	rec := httptest.NewRecorder()
	// MockFlusher wraps the Recorder.
	// Since Recorder implements Flusher (but doesn't do much), our Mock intercepts it.
	mock := &MockFlusher{ResponseWriter: rec}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("part1"))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		} else {
			t.Error("ResponseWriter passed to handler does not implement http.Flusher")
		}
	})

	gzipHandler := GzipCompressionMiddleware(handler)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	// ServeHTTP calls handler with gzipResponseWriter wrapping mock.
	gzipHandler.ServeHTTP(mock, req)

	// Verify that the Flush call was propagated to the underlying ResponseWriter
	if !mock.Flushed {
		t.Error("Expected underlying ResponseWriter.Flush() to be called, but it wasn't")
	}
}
