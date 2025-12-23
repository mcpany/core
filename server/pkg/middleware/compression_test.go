// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("Hello, World!"))
	})

	gzipHandler := GzipMiddleware(handler)

	t.Run("Compresses response when Accept-Encoding contains gzip", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		gzipHandler.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") != "gzip" {
			t.Errorf("Expected Content-Encoding to be gzip, got %s", rec.Header().Get("Content-Encoding"))
		}

		if rec.Header().Get("Vary") != "Accept-Encoding" {
			t.Errorf("Expected Vary header to be Accept-Encoding, got %s", rec.Header().Get("Vary"))
		}

		// Verify content is gzipped
		reader, err := gzip.NewReader(rec.Body)
		if err != nil {
			t.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer reader.Close()

		body, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Failed to read gzipped body: %v", err)
		}

		if string(body) != "Hello, World!" {
			t.Errorf("Expected body 'Hello, World!', got '%s'", string(body))
		}
	})

	t.Run("Does not compress when Accept-Encoding does not contain gzip", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		gzipHandler.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Expected Content-Encoding not to be gzip")
		}

		if rec.Body.String() != "Hello, World!" {
			t.Errorf("Expected body 'Hello, World!', got '%s'", rec.Body.String())
		}
	})

	t.Run("Does not compress non-compressible content types", func(t *testing.T) {
		imageHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte("fake png data"))
		})
		gzipImageHandler := GzipMiddleware(imageHandler)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		gzipImageHandler.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Expected Content-Encoding not to be gzip for image/png")
		}

		if rec.Body.String() != "fake png data" {
			t.Errorf("Expected body 'fake png data', got '%s'", rec.Body.String())
		}
	})

	t.Run("Skips compression for Upgrade requests", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Upgrade", "websocket")
		rec := httptest.NewRecorder()

		gzipHandler.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Expected Content-Encoding not to be gzip for Upgrade request")
		}
	})

	t.Run("Does not compress SSE but flushes", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			w.(http.Flusher).Flush()
		})
		gzipHandler := GzipMiddleware(handler)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		gzipHandler.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Expected Content-Encoding NOT to be gzip for SSE")
		}
		if !rec.Flushed {
			t.Error("Expected response to be flushed")
		}
	})

	t.Run("Does not compress 204 No Content", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
		gzipHandler := GzipMiddleware(handler)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		gzipHandler.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Expected Content-Encoding not to be gzip for 204")
		}
		if rec.Body.Len() > 0 {
			t.Errorf("Expected 204 response to have empty body, got %d bytes", rec.Body.Len())
		}
	})

	t.Run("Does not compress 304 Not Modified", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotModified)
		})
		gzipHandler := GzipMiddleware(handler)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		gzipHandler.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Expected Content-Encoding not to be gzip for 304")
		}
		if rec.Body.Len() > 0 {
			t.Errorf("Expected 304 response to have empty body, got %d bytes", rec.Body.Len())
		}
	})

	t.Run("Does not compress 206 Partial Content", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write([]byte("Partial data"))
		})
		gzipHandler := GzipMiddleware(handler)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		gzipHandler.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Expected Content-Encoding not to be gzip for 206")
		}
		if rec.Body.String() != "Partial data" {
			t.Errorf("Expected body 'Partial data', got '%s'", rec.Body.String())
		}
	})
}
