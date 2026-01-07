// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"container/ring"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// DebugEntry represents a captured HTTP request/response.
type DebugEntry struct {
	Timestamp       time.Time     `json:"timestamp"`
	Method          string        `json:"method"`
	Path            string        `json:"path"`
	Status          int           `json:"status"`
	Duration        time.Duration `json:"duration"`
	RequestHeaders  http.Header   `json:"request_headers"`
	ResponseHeaders http.Header   `json:"response_headers"`
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

// Middleware returns the gin handler.
func (d *Debugger) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		duration := time.Since(start)

		entry := DebugEntry{
			Timestamp:       start,
			Method:          c.Request.Method,
			Path:            c.Request.URL.Path,
			Status:          c.Writer.Status(),
			Duration:        duration,
			RequestHeaders:  c.Request.Header,
			ResponseHeaders: c.Writer.Header(),
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
