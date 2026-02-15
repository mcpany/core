package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecoveryMiddleware(t *testing.T) {
	// 1. Create a handler that panics
	panickingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	// 2. Wrap it with RecoveryMiddleware
	handler := RecoveryMiddleware(panickingHandler)

	// 3. Make a request
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// 4. Serve
	// This should not crash
	handler.ServeHTTP(w, req)

	// 5. Check response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Internal Server Error\n", w.Body.String())
}

func TestRecoveryWithCompliance(t *testing.T) {
	// 1. Create a handler that panics
	panickingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	// 2. Wrap it with RecoveryMiddleware AND JSONRPCComplianceMiddleware
	// Order: Compliance(Recovery(Handler))
	handler := JSONRPCComplianceMiddleware(RecoveryMiddleware(panickingHandler))

	// 3. Make a request (POST for JSON-RPC compliance trigger)
	req := httptest.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()

	// 4. Serve
	handler.ServeHTTP(w, req)

	// 5. Check response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	// Body should be JSON-RPC error
	// {"jsonrpc":"2.0","id":null,"error":{"code":-32603,"message":"Internal error"}} (Data omitted due to omitempty)
	body := w.Body.String()
	assert.Contains(t, body, `"code":-32603`)
	assert.Contains(t, body, `"message":"Internal error"`)
	// Data should be omitted or null, effectively safe.
	assert.NotContains(t, body, "something went wrong") // Stack trace/panic msg should be hidden
}
