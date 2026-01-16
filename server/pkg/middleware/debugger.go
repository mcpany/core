// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"container/ring"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DebugEntry represents a captured HTTP request/response.
type DebugEntry struct {
	ID              string        `json:"id"`
	Timestamp       time.Time     `json:"timestamp"`
	Method          string        `json:"method"`
	Path            string        `json:"path"`
	Status          int           `json:"status"`
	Duration        time.Duration `json:"duration"`
	RequestHeaders  http.Header   `json:"request_headers"`
	ResponseHeaders http.Header   `json:"response_headers"`
	RequestBody     string        `json:"request_body,omitempty"`
	ResponseBody    string        `json:"response_body,omitempty"`
}

// Debugger monitors and records traffic for inspection.
type Debugger struct {
	ring        *ring.Ring
	mu          sync.Mutex
	limit       int
	maxBodySize int64
}

// NewDebugger creates a new Debugger middleware.
func NewDebugger(size int) *Debugger {
	return &Debugger{
		ring:        ring.New(size),
		limit:       size,
		maxBodySize: 10 * 1024, // 10KB default limit for body capture
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body        *bytes.Buffer
	maxBodySize int64
	overflow    bool
}

// Write writes the data to the connection and captures it for the log.
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	if !w.overflow {
		if int64(w.body.Len()+len(b)) > w.maxBodySize {
			// Capture what fits, then mark overflow
			remaining := w.maxBodySize - int64(w.body.Len())
			if remaining > 0 {
				w.body.Write(b[:remaining])
			}
			w.overflow = true
			w.body.WriteString("... [truncated]")
		} else {
			w.body.Write(b)
		}
	}
	return w.ResponseWriter.Write(b)
}

// WriteString writes the string to the connection and captures it for the log.
func (w *bodyLogWriter) WriteString(s string) (int, error) {
	if !w.overflow {
		if int64(w.body.Len()+len(s)) > w.maxBodySize {
			remaining := w.maxBodySize - int64(w.body.Len())
			if remaining > 0 {
				w.body.WriteString(s[:remaining])
			}
			w.overflow = true
			w.body.WriteString("... [truncated]")
		} else {
			w.body.WriteString(s)
		}
	}
	return w.ResponseWriter.WriteString(s)
}

// responseWriterInterceptor wraps http.ResponseWriter to capture status and body.
// It implements http.Flusher to support streaming responses.
type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode  int
	body        *bytes.Buffer
	maxBodySize int64
	overflow    bool
}

func (w *responseWriterInterceptor) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterInterceptor) Write(b []byte) (int, error) {
	if !w.overflow {
		if int64(w.body.Len()+len(b)) > w.maxBodySize {
			// Capture what fits, then mark overflow
			remaining := w.maxBodySize - int64(w.body.Len())
			if remaining > 0 {
				w.body.Write(b[:remaining])
			}
			w.overflow = true
			w.body.WriteString("... [truncated]")
		} else {
			w.body.Write(b)
		}
	}
	return w.ResponseWriter.Write(b)
}

// Flush implements http.Flusher.
func (w *responseWriterInterceptor) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// readCloserWrapper wraps a Reader and a Closer.
type readCloserWrapper struct {
	io.Reader
	io.Closer
}

// Middleware returns the gin handler.
func (d *Debugger) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqID := uuid.New().String()

		// Capture Request Body
		var reqBody string
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, d.maxBodySize))
			if err == nil {
				reqBody = string(bodyBytes)
				if int64(len(bodyBytes)) == d.maxBodySize {
					reqBody += "... [truncated]"
				}

				// Reset request body so downstream can read it all.
				// We wrap the original closer to ensure it gets closed properly.
				multiReader := io.MultiReader(bytes.NewReader(bodyBytes), c.Request.Body)
				c.Request.Body = &readCloserWrapper{
					Reader: multiReader,
					Closer: c.Request.Body,
				}
			}
		}

		// Prepare Response Body Capture
		// Skip for WebSockets
		if c.Request.Header.Get("Upgrade") == "websocket" {
			c.Next()
			return
		}

		blw := &bodyLogWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
			maxBodySize:    d.maxBodySize,
		}
		c.Writer = blw

		// Process request
		c.Next()

		duration := time.Since(start)

		// Check content type to avoid storing binary data as string
		respContentType := c.Writer.Header().Get("Content-Type")
		var respBody string
		if isTextContent(respContentType) {
			respBody = blw.body.String()
		} else {
			respBody = "[Binary or Non-Text Data]"
		}

		entry := DebugEntry{
			ID:              reqID,
			Timestamp:       start,
			Method:          c.Request.Method,
			Path:            c.Request.URL.Path,
			Status:          c.Writer.Status(),
			Duration:        duration,
			RequestHeaders:  c.Request.Header,
			ResponseHeaders: c.Writer.Header(),
			RequestBody:     reqBody,
			ResponseBody:    respBody,
		}

		d.mu.Lock()
		d.ring.Value = entry
		d.ring = d.ring.Next()
		d.mu.Unlock()
	}
}

// HTTPMiddleware returns a standard net/http middleware.
func (d *Debugger) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := uuid.New().String()

		// Capture Request Body
		var reqBody string
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, d.maxBodySize))
			if err == nil {
				reqBody = string(bodyBytes)
				if int64(len(bodyBytes)) == d.maxBodySize {
					reqBody += "... [truncated]"
				}

				// Reset request body so downstream can read it all.
				multiReader := io.MultiReader(bytes.NewReader(bodyBytes), r.Body)
				r.Body = &readCloserWrapper{
					Reader: multiReader,
					Closer: r.Body,
				}
			}
		}

		// Prepare Response Body Capture
		// Skip for WebSockets
		if r.Header.Get("Upgrade") == "websocket" {
			next.ServeHTTP(w, r)
			return
		}

		wi := &responseWriterInterceptor{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default
			body:           bytes.NewBuffer(nil),
			maxBodySize:    d.maxBodySize,
		}

		// Process request
		next.ServeHTTP(wi, r)

		duration := time.Since(start)

		// Check content type to avoid storing binary data as string
		respContentType := wi.Header().Get("Content-Type")
		var respBody string
		if isTextContent(respContentType) {
			respBody = wi.body.String()
		} else {
			respBody = "[Binary or Non-Text Data]"
		}

		entry := DebugEntry{
			ID:              reqID,
			Timestamp:       start,
			Method:          r.Method,
			Path:            r.URL.Path,
			Status:          wi.statusCode,
			Duration:        duration,
			RequestHeaders:  r.Header,
			ResponseHeaders: wi.Header(),
			RequestBody:     reqBody,
			ResponseBody:    respBody,
		}

		d.mu.Lock()
		d.ring.Value = entry
		d.ring = d.ring.Next()
		d.mu.Unlock()
	})
}

func isTextContent(contentType string) bool {
	if contentType == "" {
		return true // Assume text if unknown
	}
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "json") ||
		strings.Contains(contentType, "text") ||
		strings.Contains(contentType, "xml") ||
		strings.Contains(contentType, "form-urlencoded")
}

// Entries returns the last captured entries.
func (d *Debugger) Entries() []DebugEntry {
	d.mu.Lock()
	defer d.mu.Unlock()

	entries := make([]DebugEntry, 0, d.limit)
	d.ring.Do(func(p interface{}) {
		if p != nil {
			entries = append(entries, p.(DebugEntry))
		}
	})
	return entries
}

// Handler returns a gin handler to view entries.
func (d *Debugger) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, d.Entries())
	}
}

// HTTPHandler returns a standard http.HandlerFunc to view entries.
func (d *Debugger) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(d.Entries()); err != nil {
			http.Error(w, "Failed to encode entries", http.StatusInternalServerError)
		}
	}
}
