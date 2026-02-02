package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebuggerTraceExtraction(t *testing.T) {
	debugger := NewDebugger(10)
	defer debugger.Close()

	handler := debugger.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Case 1: W3C Traceparent
	t.Run("W3C Traceparent", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/trace", nil)
		traceID := "4bf92f3577b34da6a3ce929d0e0e4736"
		parentID := "00f067aa0ba902b7"
		req.Header.Set("traceparent", "00-"+traceID+"-"+parentID+"-01")
		handler.ServeHTTP(w, req)

		entries := waitForEntries(t, debugger, 1)
		assert.Equal(t, traceID, entries[0].TraceID)
		assert.Equal(t, parentID, entries[0].ParentID)
		assert.NotEmpty(t, entries[0].SpanID)
	})

	// Reset debugger ring by creating new one or consuming all (easier to create new one if needed, but here we can just wait for next entry if we send another)
	// But waitForEntries waits for *count*. If we already have 1, we might get the same one if we don't clear or advance.
	// `waitForEntries` calls `d.Entries()` which iterates ring.
	// Since we are consuming, we need to distinguish new entries.
	// Let's just create new debugger for simplicity.
}

func TestDebuggerXTraceID(t *testing.T) {
	debugger := NewDebugger(10)
	defer debugger.Close()

	handler := debugger.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Case 2: X-Trace-ID
	t.Run("X-Trace-ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/trace", nil)
		traceID := "custom-trace-id"
		req.Header.Set("X-Trace-ID", traceID)
		handler.ServeHTTP(w, req)

		entries := waitForEntries(t, debugger, 1)
		assert.Equal(t, traceID, entries[0].TraceID)
		assert.NotEmpty(t, entries[0].SpanID)
		assert.Empty(t, entries[0].ParentID)
	})
}

func TestDebuggerNoTrace(t *testing.T) {
	debugger := NewDebugger(10)
	defer debugger.Close()

	handler := debugger.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Case 3: No Trace Header
	t.Run("No Trace", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/trace", nil)
		handler.ServeHTTP(w, req)

		entries := waitForEntries(t, debugger, 1)
		assert.NotEmpty(t, entries[0].TraceID)
		assert.NotEmpty(t, entries[0].SpanID)
		assert.Empty(t, entries[0].ParentID)
	})
}
