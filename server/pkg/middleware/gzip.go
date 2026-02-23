// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// ⚡ Bolt: Minimum size threshold for gzip compression.
// Compressing very small payloads (< 1400 bytes, approx 1 MTU) often degrades performance
// due to CPU overhead and increased size from headers/framing.
const minSize = 1400

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

// ⚡ BOLT: Buffer pool to reduce allocations for small responses.
// Randomized Selection from Top 5 High-Impact Targets.
type pooledBuffer struct {
	data []byte
}

var byteBufferPool = sync.Pool{
	New: func() interface{} {
		return &pooledBuffer{data: make([]byte, 0, minSize)}
	},
}

// GzipCompressionMiddleware returns a middleware that compresses HTTP responses using Gzip.
//
// Summary: Middleware that compresses HTTP responses using Gzip if supported by the client.
//
// Parameters:
//   - next: http.Handler. The next handler in the chain.
//
// Returns:
//   - http.Handler: The wrapped handler that performs compression.
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

		pb := byteBufferPool.Get().(*pooledBuffer)
		pb.data = pb.data[:0]

		gzw := &gzipResponseWriter{
			ResponseWriter: w,
			pool:           &pool,
			buf:            pb,
			code:           http.StatusOK, // Default status code
		}
		defer gzw.Close()

		next.ServeHTTP(gzw, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
	pool   *sync.Pool

	headerWritten bool
	code          int
	buf           *pooledBuffer
}

// Write writes the data to the connection as part of an HTTP reply.
//
// Summary: Writes data to the response, buffering until compression threshold is met.
//
// Parameters:
//   - b: []byte. The data to write.
//
// Returns:
//   - int: The number of bytes written.
//   - error: An error if the write fails.
//
// Side Effects:
//   - May flush headers and start compression if the buffer threshold is exceeded.
func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	// If we are already compressing, write to gzip writer
	if w.writer != nil {
		return w.writer.Write(b)
	}

	// If we decided NOT to compress (headers written, no writer), pass through
	if w.headerWritten {
		return w.ResponseWriter.Write(b)
	}

	// ⚡ BOLT: Optimization: Bypass buffer for large writes to avoid unnecessary allocation and copy.
	// Randomized Selection from Top 5 High-Impact Targets
	if len(w.buf.data) == 0 && len(b) >= minSize {
		// Ensure Content-Type is set if missing (Sniffing)
		if w.Header().Get("Content-Type") == "" {
			detectBuf := b
			if len(detectBuf) > 512 {
				detectBuf = detectBuf[:512]
			}
			w.Header().Set("Content-Type", http.DetectContentType(detectBuf))
		}

		// Flush buffer (which is empty, but sets up headers and writer)
		if err := w.flushBuffer(true); err != nil {
			return 0, err
		}
		// Write directly to the gzip writer
		return w.writer.Write(b)
	}

	// Otherwise, buffer the data
	w.buf.data = append(w.buf.data, b...)

	// Check if we crossed the threshold
	if len(w.buf.data) >= minSize {
		// Flush buffer and start gzipping
		if err := w.flushBuffer(true); err != nil {
			return 0, err
		}
		// Since flushBuffer wrote the buffer to the gzip writer,
		// and we are now in gzip mode, future writes go to w.writer.
		// However, flushBuffer consumed w.buf (which contained b).
		// So we return len(b) as we successfully "wrote" it (into the buffer -> gzip).
		return len(b), nil
	}

	return len(b), nil
}

// WriteHeader captures the status code.
//
// Summary: Captures the status code for later writing.
//
// Parameters:
//   - code: int. The HTTP status code.
//
// Returns:
//   None.
//
// Side Effects:
//   - If content is not compressible, headers are flushed immediately.
func (w *gzipResponseWriter) WriteHeader(code int) {
	if w.headerWritten {
		return
	}
	w.code = code

	// Check immediately if content is not compressible
	contentType := w.Header().Get("Content-Type")
	if !isCompressible(contentType) {
		// If not compressible, we can flush/write headers immediately and stop buffering checks.
		// passing false to flushBuffer will write headers and any buffered data.
		_ = w.flushBuffer(false)
	}
}

// flushBuffer transitions from buffering to writing.
// startGzip: true to enable gzip, false to write raw.
func (w *gzipResponseWriter) flushBuffer(startGzip bool) error {
	if w.headerWritten && w.writer == nil && startGzip {
		// This shouldn't happen if logic is correct, but safety check.
		// If headers were already written as raw, we can't switch to gzip.
		_, err := w.ResponseWriter.Write(w.buf.data)
		return err
	}
	if w.headerWritten {
		return nil
	}

	w.headerWritten = true

	// Ensure Content-Type is set if missing
	if w.Header().Get("Content-Type") == "" {
		// Use the first 512 bytes of buffer for detection
		detectBuf := w.buf.data
		if len(detectBuf) > 512 {
			detectBuf = detectBuf[:512]
		}
		w.Header().Set("Content-Type", http.DetectContentType(detectBuf))
	}

	if startGzip {
		w.Header().Del("Content-Length")
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")
		w.ResponseWriter.WriteHeader(w.code)

		w.writer = w.pool.Get().(*gzip.Writer)
		w.writer.Reset(w.ResponseWriter)

		if len(w.buf.data) > 0 {
			_, err := w.writer.Write(w.buf.data)
			// ⚡ BOLT: Cap pool buffer size to prevent memory leaks.
			// Randomized Selection from Top 5 High-Impact Targets
			if cap(w.buf.data) <= 65536 {
				byteBufferPool.Put(w.buf)
			}
			w.buf = nil // Release reference
			return err
		}
		return nil
	}

	w.ResponseWriter.WriteHeader(w.code)
	if len(w.buf.data) > 0 {
		_, err := w.ResponseWriter.Write(w.buf.data)
		// ⚡ BOLT: Cap pool buffer size to prevent memory leaks.
		// Randomized Selection from Top 5 High-Impact Targets
		if cap(w.buf.data) <= 65536 {
			byteBufferPool.Put(w.buf)
		}
		w.buf = nil
		return err
	}
	return nil
}

// ⚡ BOLT: Implemented Flush to support streaming responses (e.g. SSE).
// Randomized Selection from Top 5 High-Impact Targets
// Flush implements the http.Flusher interface.
//
// Summary: Flushes the compressed stream to the client.
//
// Parameters:
//   None.
//
// Returns:
//   None.
//
// Side Effects:
//   - Flushes the underlying response writer.
func (w *gzipResponseWriter) Flush() {
	// If we haven't written headers yet, we are still buffering.
	// Force flush the buffer to start gzip stream (even if small).
	// This ensures that explicit Flush calls are respected.
	if !w.headerWritten {
		_ = w.flushBuffer(true)
	}

	// Now flush the gzip writer if active
	if w.writer != nil {
		_ = w.writer.Flush()
	}

	// Finally, flush the underlying ResponseWriter
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Close closes the gzip writer and returns it to the pool.
//
// Summary: Closes the Gzip writer and releases resources.
//
// Parameters:
//   None.
//
// Returns:
//   None.
//
// Side Effects:
//   - Writes any remaining buffered data.
//   - Returns the gzip writer to the pool.
func (w *gzipResponseWriter) Close() {
	if w.writer != nil {
		_ = w.writer.Close()
		w.pool.Put(w.writer)
		w.writer = nil
		return
	}

	// If headers haven't been written, it means we are still buffering (Small Response).
	if !w.headerWritten {
		// We are at the end, so we know the total size = len(w.buf).
		// Set Content-Length optimization.
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", http.DetectContentType(w.buf.data))
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(w.buf.data)))
	}

	_ = w.flushBuffer(false)

	// In case flushBuffer didn't run or didn't clear buf
	if w.buf != nil {
		// ⚡ BOLT: Cap pool buffer size to prevent memory leaks.
		// Randomized Selection from Top 5 High-Impact Targets
		if cap(w.buf.data) <= 65536 {
			byteBufferPool.Put(w.buf)
		}
		w.buf = nil
	}
}
