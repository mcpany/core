package middleware

import (
	"context"
	"net/http"

	"github.com/mcpany/core/server/pkg/logging"
)

// AsyncRLTelemetryOrchestrator collects reasoning traces asynchronously.
type AsyncRLTelemetryOrchestrator struct {
	traceChan chan string
}

// NewAsyncRLTelemetryOrchestrator creates a new AsyncRLTelemetryOrchestrator.
func NewAsyncRLTelemetryOrchestrator(bufferSize int) *AsyncRLTelemetryOrchestrator {
	return &AsyncRLTelemetryOrchestrator{
		traceChan: make(chan string, bufferSize),
	}
}

// Start Background Processor
func (m *AsyncRLTelemetryOrchestrator) Start(ctx context.Context) {
	go func() {
		logger := logging.GetLogger().With("component", "async_rl_telemetry")
		for {
			select {
			case trace := <-m.traceChan:
				logger.DebugContext(ctx, "Exporting trace to RL pipeline", "trace_preview", trace[:min(len(trace), 50)])
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Handle implements the HTTP middleware interface.
func (m *AsyncRLTelemetryOrchestrator) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock logic: Look for trace delta in headers
		traceDelta := r.Header.Get("X-RL-Trace-Delta")
		if traceDelta != "" {
			select {
			case m.traceChan <- traceDelta:
				// Successfully queued
			default:
				// Buffer full, drop telemetry (non-blocking)
				logging.GetLogger().WarnContext(r.Context(), "Async RL trace dropped, buffer full")
			}
		}

		next.ServeHTTP(w, r)
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
