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

var compressibleContentTypes = []string{
	"text/html",
	"text/css",
	"text/plain",
	"text/javascript",
	"application/javascript",
	"application/x-javascript",
	"application/json",
	"application/xml",
	"text/xml",
	"image/svg+xml",
}

func isCompressible(contentType string) bool {
	if contentType == "" {
		return true // Assume text/plain or similar if unknown? Or sniff? net/http defaults to sniffing.
	}
	// Handle charset parameters (e.g. "text/html; charset=utf-8")
	ct := strings.Split(contentType, ";")[0]
	ct = strings.TrimSpace(ct)

	for _, t := range compressibleContentTypes {
		if t == ct {
			return true
		}
	}
	return false
}

// GzipCompressionMiddleware returns a middleware that compresses HTTP responses using Gzip.
// It checks the Accept-Encoding header and only compresses if the client supports gzip.
// It also checks the Content-Type to ensure we only compress compressible types.
func GzipCompressionMiddleware(next http.Handler) http.Handler {
	pool := sync.Pool{
		New: func() interface{} {
			w, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
			return w
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// specific check for websocket upgrade requests
		if r.Header.Get("Upgrade") != "" {
			next.ServeHTTP(w, r)
			return
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gzw := &gzipResponseWriter{
			ResponseWriter: w,
			pool:           &pool,
		}
		defer gzw.Close()

		next.ServeHTTP(gzw, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer      *gzip.Writer
	pool        *sync.Pool
	wroteHeader bool
	shouldGzip  bool
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if w.writer != nil {
		return w.writer.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	contentType := w.Header().Get("Content-Type")

	if isCompressible(contentType) {
		w.shouldGzip = true
		w.ResponseWriter.Header().Del("Content-Length")
		w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		w.ResponseWriter.Header().Add("Vary", "Accept-Encoding")

		// Initialize gzip writer
		w.writer = w.pool.Get().(*gzip.Writer)
		w.writer.Reset(w.ResponseWriter)
	}

	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Close() {
	if w.writer != nil {
		_ = w.writer.Close()
		w.pool.Put(w.writer)
		w.writer = nil
	}
}
