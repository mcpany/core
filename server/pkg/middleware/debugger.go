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
	mu          sync.RWMutex
	limit       int
	maxBodySize int64
	ingress     chan DebugEntry
	done        chan struct{}
}

// NewDebugger initializes a new Debugger with a specified ring buffer size.
//
// Parameters:
//   - size: int. The maximum number of entries to keep in the ring buffer.
//
// Returns:
//   - *Debugger: The initialized Debugger.
func NewDebugger(size int) *Debugger {
	d := &Debugger{
		ring:        ring.New(size),
		limit:       size,
		maxBodySize: 10 * 1024, // 10KB default limit for body capture
		ingress:     make(chan DebugEntry, size*2),
		done:        make(chan struct{}),
	}
	go d.process()
	return d
}

// process runs in the background to handle log entries.
func (d *Debugger) process() {
	for entry := range d.ingress {
		d.mu.Lock()
		d.ring.Value = entry
		d.ring = d.ring.Next()
		d.mu.Unlock()
	}
	close(d.done)
}

// Close stops the background processor.
func (d *Debugger) Close() {
	close(d.ingress)
	<-d.done
}

type bodyLogWriter struct {
	http.ResponseWriter
	body        *bytes.Buffer
	maxBodySize int64
	overflow    bool
	status      int
	wroteHeader bool
}

// Write intercepts the response write to capture the body.
//
// Parameters:
//   - b: []byte. The data to write.
//
// Returns:
//   - int: The number of bytes written.
//   - error: An error if the write fails.
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
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

// WriteHeader sends an HTTP response header with the provided status code.
//
// statusCode is the HTTP status code to send.
func (w *bodyLogWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.status = statusCode
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

// readCloserWrapper wraps a Reader and a Closer.
type readCloserWrapper struct {
	io.Reader
	io.Closer
}

// Handler returns the http handler wrapping the next handler to capture request and response details.
//
// Parameters:
//   - next: http.Handler. The next handler in the chain.
//
// Returns:
//   - http.Handler: The wrapped handler.
func (d *Debugger) Handler(next http.Handler) http.Handler {
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
				// We wrap the original closer to ensure it gets closed properly.
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

		blw := &bodyLogWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: w,
			maxBodySize:    d.maxBodySize,
			status:         http.StatusOK, // Default
		}

		// Process request
		next.ServeHTTP(blw, r)

		duration := time.Since(start)

		// Check content type to avoid storing binary data as string
		respContentType := blw.Header().Get("Content-Type")
		var respBody string
		if isTextContent(respContentType) {
			respBody = blw.body.String()
		} else {
			respBody = "[Binary or Non-Text Data]"
		}

		entry := DebugEntry{
			ID:              reqID,
			Timestamp:       start,
			Method:          r.Method,
			Path:            r.URL.Path,
			Status:          blw.status,
			Duration:        duration,
			RequestHeaders:  r.Header,
			ResponseHeaders: blw.Header(),
			RequestBody:     reqBody,
			ResponseBody:    respBody,
		}

		// ⚡ BOLT: Move ring buffer updates to background worker to avoid blocking request
		// Randomized Selection from Top 5 High-Impact Targets
		select {
		case d.ingress <- entry:
		default:
			// Buffer full, drop entry to preserve system stability
		}
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
//
// Returns the result.
func (d *Debugger) Entries() []DebugEntry {
	d.mu.RLock()
	defer d.mu.RUnlock()

	entries := make([]DebugEntry, 0, d.limit)
	d.ring.Do(func(p interface{}) {
		if p != nil {
			entries = append(entries, p.(DebugEntry))
		}
	})
	return entries
}

// APIHandler returns a http.HandlerFunc to view entries.
//
// Returns the result.
func (d *Debugger) APIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(d.Entries())
	}
}
