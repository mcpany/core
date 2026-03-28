package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAsyncRLTelemetryOrchestrator(t *testing.T) {
	orchestrator := NewAsyncRLTelemetryOrchestrator(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	orchestrator.Start(ctx)

	// Valid trace header
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-RL-Trace-Delta", "reasoning_step_1")

	rr := httptest.NewRecorder()
	handler := orchestrator.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Give the goroutine time to process
	time.Sleep(10 * time.Millisecond)
}

func TestAsyncRLTelemetryOrchestrator_BufferFull(t *testing.T) {
	// Create with buffer size 0, no background processor started
	orchestrator := NewAsyncRLTelemetryOrchestrator(0)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-RL-Trace-Delta", "reasoning_step_1")

	rr := httptest.NewRecorder()
	handler := orchestrator.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Should drop the trace without blocking
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
