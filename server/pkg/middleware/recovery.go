package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/mcpany/core/server/pkg/logging"
)

// RecoveryMiddleware recovers from panics in the handler chain, logs the panic,
// and returns a generic 500 Internal Server Error response.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				log := logging.GetLogger()
				stack := string(debug.Stack())
				log.Error("Panic recovered", "error", err, "stack", stack, "url", r.URL.String(), "method", r.Method)

				// Return generic 500 error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
