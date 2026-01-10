// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"container/ring"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxBodySize = 10 * 1024 // 10KB
const maxCaptureSize = 1024 * 1024 // 1MB Limit for capture attempts

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
	RequestBody     string        `json:"request_body"`
	ResponseBody    string        `json:"response_body"`
}

// ReplayRequest represents a request to replay a call.
type ReplayRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// ReplayResponse represents the result of a replay.
type ReplayResponse struct {
	Status     int               `json:"status"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	DurationMs int64             `json:"duration_ms"`
}

// Debugger monitors and records traffic for inspection.
type Debugger struct {
	ring  *ring.Ring
	mu    sync.Mutex
	limit int
}

// NewDebugger creates a new Debugger middleware.
func NewDebugger(size int) *Debugger {
	return &Debugger{
		ring:  ring.New(size),
		limit: size,
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	if w.body.Len() < maxBodySize {
		w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (w bodyLogWriter) WriteString(s string) (int, error) {
	if w.body.Len() < maxBodySize {
		w.body.WriteString(s)
	}
	return w.ResponseWriter.WriteString(s)
}

// Middleware returns the gin handler.
func (d *Debugger) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqID := uuid.New().String()

		// Capture Request Body safely
		var reqBodyBytes []byte
		if c.Request.Body != nil {
			// Check Content-Length to avoid reading huge bodies
			if c.Request.ContentLength > int64(maxCaptureSize) {
				reqBodyBytes = []byte(fmt.Sprintf("<<Request body too large to capture (Content-Length: %d)>>", c.Request.ContentLength))
			} else {
				// Use MaxBytesReader to enforce limit during read
				// This protects against chunked encoding attacks or lies about Content-Length
				maxReader := http.MaxBytesReader(c.Writer, c.Request.Body, int64(maxCaptureSize))
				bodyBytes, err := io.ReadAll(maxReader)

				if err != nil {
					// If we hit the limit, MaxBytesReader returns an error.
					// We capture what we got, but we likely broke the stream for the handler.
					if len(bodyBytes) > 0 {
						// Fix gocritic appendAssign
						reqBodyBytes = append(reqBodyBytes, bodyBytes...)
						reqBodyBytes = append(reqBodyBytes, []byte("... <<Error/Truncated>>")...)
					} else {
						reqBodyBytes = []byte("<<Error reading body: " + err.Error() + ">>")
					}
					// Restore partial body so handler sees something (likely causing a parsing error, which is correct behavior for a truncated request)
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				} else {
					// Success
					reqBodyBytes = bodyBytes
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}
		}

		// Capture Response Body
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process request
		c.Next()

		duration := time.Since(start)

		// Prepare bodies for storage (truncated for display)
		reqBodyStr := string(reqBodyBytes)
		if len(reqBodyStr) > maxBodySize {
			reqBodyStr = reqBodyStr[:maxBodySize] + "...(truncated)"
		}

		respBodyStr := blw.body.String()
		if len(respBodyStr) > maxBodySize {
			respBodyStr = respBodyStr[:maxBodySize] + "...(truncated)"
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
			RequestBody:     reqBodyStr,
			ResponseBody:    respBodyStr,
		}

		d.mu.Lock()
		d.ring.Value = entry
		d.ring = d.ring.Next()
		d.mu.Unlock()
	}
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

// Handler returns a http.HandlerFunc to view entries.
func (d *Debugger) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, d.Entries())
	}
}

// ReplayHandler returns a http.HandlerFunc to execute a replay request.
func (d *Debugger) ReplayHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ReplayRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
			return
		}

		// Create Request
		var bodyReader io.Reader
		if req.Body != "" {
			bodyReader = strings.NewReader(req.Body)
		}

		// Use NewRequestWithContext
		httpReq, err := http.NewRequestWithContext(c.Request.Context(), req.Method, req.URL, bodyReader)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create request: " + err.Error()})
			return
		}

		// Add Headers
		for k, v := range req.Headers {
			httpReq.Header.Set(k, v)
		}

		// Execute
		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		start := time.Now()
		resp, err := client.Do(httpReq)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to execute request: " + err.Error()})
			return
		}

		// Read Response
		// Use MaxBytesReader here too for safety?
		// Replay responses could be large, but we only return JSON.
		// Let's limit response read to prevent memory issues here too.
		const maxReplayResponseSize = 1024 * 1024 // 1MB

		respBody, err := io.ReadAll(io.LimitReader(resp.Body, int64(maxReplayResponseSize)))
		// We ignore error from LimitReader (it just stops), but check ReadAll error
		_ = resp.Body.Close() // Explicitly ignore error on Close as we're done

		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to read response: " + err.Error()})
			return
		}
		duration := time.Since(start)

		// Prepare Response Headers
		respHeaders := make(map[string]string)
		for k, v := range resp.Header {
			respHeaders[k] = strings.Join(v, ", ")
		}

		c.JSON(http.StatusOK, ReplayResponse{
			Status:     resp.StatusCode,
			Headers:    respHeaders,
			Body:       string(respBody),
			DurationMs: duration.Milliseconds(),
		})
	}
}
