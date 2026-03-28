package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/mcpany/core/server/pkg/logging"
)

// IPSCMiddleware implements the Intent-Preserving Self-Correction (IPSC) middleware.
// It tracks recursion loops (via headers) and stops infinite "Cognitive Lock" refinement loops.
type IPSCMiddleware struct {
	maxCycles int
}

// NewIPSCMiddleware creates a new IPSCMiddleware with a specified maximum number of refinement cycles.
func NewIPSCMiddleware(maxCycles int) *IPSCMiddleware {
	return &IPSCMiddleware{
		maxCycles: maxCycles,
	}
}

// Handle implements the HTTP middleware interface.
func (m *IPSCMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logging.GetLogger().With("component", "ipsc_middleware")

		// The header X-UACO-IPSC contains the current cycle count, e.g. "cycle=3"
		ipscHeader := r.Header.Get("X-UACO-IPSC")
		if ipscHeader != "" {
			parts := strings.Split(ipscHeader, "=")
			if len(parts) == 2 && strings.ToLower(parts[0]) == "cycle" {
				cycles, err := strconv.Atoi(parts[1])
				if err == nil {
					if cycles >= m.maxCycles {
						logger.WarnContext(r.Context(), "Cognitive Lock detected: Max IPSC cycles exceeded", "cycles", cycles, "max", m.maxCycles)
						http.Error(w, fmt.Sprintf("IPSC Middleware: Correction Budget Exceeded (Cycles: %d)", cycles), http.StatusTooManyRequests)
						return
					}
					// Increment cycle for upstream if we wanted to mutate, but normally
					// the agent self-reports its cycle count. We'll just pass it through here.
				} else {
					logger.WarnContext(r.Context(), "Invalid X-UACO-IPSC header format", "header", ipscHeader)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}
