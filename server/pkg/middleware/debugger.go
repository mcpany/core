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

// DebugEntry defines the core structure for debug entry within the system.
//
// Summary: DebugEntry defines the core structure for debug entry within the system.
//
// Fields:
//   - Contains the configuration and state properties required for DebugEntry functionality.
type DebugEntry struct {
	ID              string        `json:"id"`
	TraceID         string        `json:"trace_id"`
	SpanID          string        `json:"span_id"`
	ParentID        string        `json:"parent_id,omitempty"`
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

// Debugger monitors and records traffic for inspection. Summary: Middleware that captures recent HTTP traffic for debugging purposes.
//
// Summary: Debugger monitors and records traffic for inspection. Summary: Middleware that captures recent HTTP traffic for debugging purposes.
//
// Fields:
//   - Contains the configuration and state properties required for Debugger functionality.
type Debugger struct {
	ring        *ring.Ring
	mu          sync.RWMutex
	limit       int
	maxBodySize int64
	ingress     chan DebugEntry
	done        chan struct{}
}

// NewDebugger creates a new Debugger middleware.
//
// Summary: Initializes the debugger with a fixed-size ring buffer.
//
// Parameters:
//   - size: int. The number of recent requests to keep in memory.
//
// Returns:
//   - *Debugger: The initialized debugger.
//
// Side Effects:
//   - Starts a background goroutine to process debug entries.
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

// Close stops the background processor. Summary: Shuts down the debugger and releases resources. Side Effects: - Closes the ingress channel. - Waits for the background processor to finish.
//
// Parameters:
//   - None
//
// Returns:
//   - None
//
// Errors:
//   - None
//
// Side Effects:
//   - None
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

// Write writes the data to the connection and captures it for the log.
//
// Summary: Writes data to the response and captures a copy for the debug log.
//
// Parameters:
//   - b: []byte. The data to write.
//
// Returns:
//   - int: The number of bytes written.
//   - error: An error if the write fails.
//
// Side Effects:
//   - Writes to the underlying http.ResponseWriter.
//   - Writes to the internal buffer for logging, truncating if necessary.
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

// WriteHeader stores the header into the persistent storage.
//
// Summary: Stores the header into the persistent storage.
//
// Parameters:
//   - statusCode (int): The status code parameter used in the operation.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Modifies global state, writes to the database, or establishes network connections.
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

// Handler returns the http handler.
//
// Summary: Returns an HTTP handler that captures traffic.
//
// Parameters:
//   - next: http.Handler. The next handler in the chain.
//
// Returns:
//   - http.Handler: The wrapped handler.
//
// Side Effects:
//   - Intercepts HTTP requests and responses.
//   - Generates trace and span IDs if missing.
//   - Captures request and response bodies (truncated).
//   - Sends debug entries to the ingress channel.
func (d *Debugger) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := uuid.New().String()

		// Trace Context Extraction (W3C Trace Context)
		// Format: 00-traceid-parentid-flags
		var traceID, parentID, spanID string
		traceParent := r.Header.Get("traceparent")
		if traceParent != "" {
			parts := strings.Split(traceParent, "-")
			if len(parts) == 4 {
				traceID = parts[1]
				parentID = parts[2]
			}
		}

		// Fallback to X-Trace-ID if available
		if traceID == "" {
			traceID = r.Header.Get("X-Trace-ID")
		}

		// Generate if missing
		if traceID == "" {
			traceID = strings.ReplaceAll(uuid.New().String(), "-", "")
		}
		// Generate new SpanID for this request
		spanID = strings.ReplaceAll(uuid.New().String(), "-", "")[:16]

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

		// Inject Trace Context
		ctx := r.Context()
		ctx = WithTraceContext(ctx, traceID, spanID, parentID)
		r = r.WithContext(ctx)

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
			TraceID:         traceID,
			SpanID:          spanID,
			ParentID:        parentID,
			Timestamp:       start,
			Method:          r.Method,
			Path:            r.URL.Path,
			Status:          blw.status,
			Duration:        duration,
			RequestHeaders:  redactHeaders(r.Header),
			ResponseHeaders: redactHeaders(blw.Header()),
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

// Entries returns the last captured entries. Summary: Retrieves the list of captured debug entries from the ring buffer. Returns: - []DebugEntry: A slice of the most recent captured requests and responses. Side Effects: - Acquires a read lock on the ring buffer.
//
// Summary: Entries returns the last captured entries. Summary: Retrieves the list of captured debug entries from the ring buffer. Returns: - []DebugEntry: A slice of the most recent captured requests and responses. Side Effects: - Acquires a read lock on the ring buffer.
//
// Parameters:
//   - None.
//
// Returns:
//   - ([]DebugEntry): The resulting []DebugEntry object containing the requested data.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// APIHandler returns a http.HandlerFunc to view entries. Summary: Returns an HTTP handler that exposes the debug entries as JSON. Returns: - http.HandlerFunc: The API handler function. Side Effects: - Encodes the entries to JSON and writes to the response.
//
// Summary: APIHandler returns a http.HandlerFunc to view entries. Summary: Returns an HTTP handler that exposes the debug entries as JSON. Returns: - http.HandlerFunc: The API handler function. Side Effects: - Encodes the entries to JSON and writes to the response.
//
// Parameters:
//   - None.
//
// Returns:
//   - (http.HandlerFunc): The resulting http.HandlerFunc object containing the requested data.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (d *Debugger) APIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(d.Entries())
	}
}

var sensitiveHeaders = map[string]struct{}{
	"Authorization": {},
	"Cookie":        {},
	"Set-Cookie":    {},
	"X-Api-Key":     {},
}

func redactHeaders(headers http.Header) http.Header {
	newHeaders := make(http.Header)
	for k, v := range headers {
		if _, ok := sensitiveHeaders[http.CanonicalHeaderKey(k)]; ok {
			newHeaders[k] = []string{"[REDACTED]"}
		} else {
			newHeaders[k] = v
		}
	}
	return newHeaders
}
