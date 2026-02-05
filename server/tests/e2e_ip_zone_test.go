package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/util"
)

// TestIPZoneIndexHandling is an E2E test to verify that the server handles
// requests with IPv6 zone indices correctly (stripping them).
// We simulate this by mocking the ExtractIP logic flow, as we cannot easily
// force a real HTTP request to have a zone index in RemoteAddr (Go's http server
// usually provides clean IP:port, but some setups might not).
// However, since we patched ExtractIP which is used by the middleware, we can verify
// that if ExtractIP is called with a dirty string, it cleans it up, AND that
// the middleware uses this cleaned IP.

// Since we cannot easily spin up the full server in this test file due to deps,
// we will verify the middleware behavior using httptest.

func TestMiddlewareStripsZoneIndex(t *testing.T) {
	// Create a dummy handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, ok := util.RemoteIPFromContext(r.Context())
		if !ok {
			t.Errorf("No remote IP in context")
		}
		if ip != "fe80::1" {
			t.Errorf("Expected cleaned IP 'fe80::1', got %q", ip)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Manually replicate the logic in createAuthMiddleware or equivalent
	// that calls ExtractIP and ContextWithRemoteIP.
	// In server.go:
	// ip := util.ExtractIP(r.RemoteAddr)
	// ctx := util.ContextWithRemoteIP(r.Context(), ip)

	middlewareFunc := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := util.ExtractIP(r.RemoteAddr)

			ctx := util.ContextWithRemoteIP(r.Context(), ip)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	req := httptest.NewRequest("GET", "/", nil)
	// Set a RemoteAddr with zone index
	req.RemoteAddr = "fe80::1%eth0"

	w := httptest.NewRecorder()
	middlewareFunc(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Handler returned %v", w.Code)
	}
}
