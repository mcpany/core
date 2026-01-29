package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipCompressionMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello, World! Repeated data for compression efficiency. Hello, World! Repeated data for compression efficiency."))
	})

	gzipHandler := GzipCompressionMiddleware(handler)

	t.Run("Accept-Encoding: gzip", func(t *testing.T) {
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

		expected := "Hello, World! Repeated data for compression efficiency. Hello, World! Repeated data for compression efficiency."
		if string(body) != expected {
			t.Errorf("Expected body %q, got %q", expected, string(body))
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

		if rec.Body.String() != "Hello, World! Repeated data for compression efficiency. Hello, World! Repeated data for compression efficiency." {
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
}
