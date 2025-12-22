// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

var gzipPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(io.Discard)
	},
}

// GzipMiddleware is a middleware that compresses HTTP responses using Gzip.
// It skips compression for WebSocket upgrades, non-compressible content types,
// and specific status codes (204, 206, 304).
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Skip if connection is an upgrade (e.g. WebSocket)
		if strings.EqualFold(r.Header.Get("Connection"), "Upgrade") || r.Header.Get("Upgrade") != "" {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Add("Vary", "Accept-Encoding")

		smartWrapper := &smartGzipResponseWriter{
			ResponseWriter: w,
		}
		defer smartWrapper.Close()

		next.ServeHTTP(smartWrapper, r)
	})
}

// smartGzipResponseWriter decides whether to use gzip based on Content-Type and Status Code.
type smartGzipResponseWriter struct {
	http.ResponseWriter
	gz          *gzip.Writer
	wroteHeader bool
	useGzip     bool
}

func (w *smartGzipResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	// Check status codes that should not be compressed or have no body
	if code == http.StatusNoContent ||
		code == http.StatusNotModified ||
		code == http.StatusPartialContent ||
		(code >= 100 && code < 200) {
		w.useGzip = false
		w.ResponseWriter.WriteHeader(code)
		return
	}

	ct := w.Header().Get("Content-Type")
	if ct == "" {
		// Default to true if unknown, assuming text-based usually.
		w.useGzip = true
	} else {
		w.useGzip = isCompressible(ct)
	}

	if w.useGzip {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")

		// Acquire writer from pool
		gz := gzipPool.Get().(*gzip.Writer)
		gz.Reset(w.ResponseWriter)
		w.gz = gz
	}

	w.ResponseWriter.WriteHeader(code)
}

func (w *smartGzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		// Detect content type if missing
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", http.DetectContentType(b))
		}
		w.WriteHeader(http.StatusOK)
	}

	if w.useGzip {
		return w.gz.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (w *smartGzipResponseWriter) Flush() {
	if w.useGzip && w.gz != nil {
		w.gz.Flush()
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *smartGzipResponseWriter) Close() error {
	if w.useGzip && w.gz != nil {
		err := w.gz.Close()
		gzipPool.Put(w.gz)
		w.gz = nil
		return err
	}
	return nil
}

func isCompressible(contentType string) bool {
	ct := strings.ToLower(contentType)
	if strings.Contains(ct, "text/") ||
		strings.Contains(ct, "application/json") ||
		strings.Contains(ct, "application/javascript") ||
		strings.Contains(ct, "application/xml") ||
		strings.Contains(ct, "image/svg+xml") ||
		strings.Contains(ct, "application/wasm") ||
		strings.Contains(ct, "application/x-yaml") {
		return true
	}
	return false
}
