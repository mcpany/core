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

// responseWrapper wraps http.ResponseWriter to capture status and body.
type responseWrapper struct {
	http.ResponseWriter
	body        *bytes.Buffer
	status      int
	maxBodySize int64
	overflow    bool
}

func (w *responseWrapper) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWrapper) Write(b []byte) (int, error) {
	if !w.overflow {
		if int64(w.body.Len()+len(b)) > w.maxBodySize {
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

// readCloserWrapper wraps a Reader and a Closer.
type readCloserWrapper struct {
	io.Reader
	io.Closer
}

// Middleware returns the http middleware.
func (d *Debugger) Middleware(next http.Handler) http.Handler {
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

		// Skip body capture for WebSockets
		if r.Header.Get("Upgrade") == "websocket" {
			next.ServeHTTP(w, r)
			return
		}

		rw := &responseWrapper{
			ResponseWriter: w,
			body:           bytes.NewBufferString(""),
			status:         http.StatusOK, // Default
			maxBodySize:    d.maxBodySize,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		// Check content type
		respContentType := rw.Header().Get("Content-Type")
		var respBody string
		if isTextContent(respContentType) {
			respBody = rw.body.String()
		} else {
			respBody = "[Binary or Non-Text Data]"
		}

		entry := DebugEntry{
			ID:              reqID,
			Timestamp:       start,
			Method:          r.Method,
			Path:            r.URL.Path,
			Status:          rw.status,
			Duration:        duration,
			RequestHeaders:  r.Header,
			ResponseHeaders: rw.Header(),
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
	// The ring buffer returns items in rotation order, not chronological.
	// We might want to sort them? Or just return as is and let UI sort.
	// UI can sort by timestamp.
	return entries
}

// Handler returns a http.HandlerFunc to view entries.
func (d *Debugger) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(d.Entries())
	}
}
