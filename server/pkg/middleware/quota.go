package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/mcpany/core/server/pkg/logging"
)

// QuotaMonitorMiddleware implements the Dynamic Usage Quota Monitor.
// It intercepts tool calls and updates usage, blocking requests if the quota is exceeded.
type QuotaMonitorMiddleware struct {
	maxTokens int
}

// NewQuotaMonitorMiddleware creates a new QuotaMonitorMiddleware.
func NewQuotaMonitorMiddleware(maxTokens int) *QuotaMonitorMiddleware {
	return &QuotaMonitorMiddleware{
		maxTokens: maxTokens,
	}
}

// Handle implements the HTTP middleware interface.
func (m *QuotaMonitorMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logging.GetLogger().With("component", "quota_monitor")

		// Mock logic: we check a header for current usage
		usageHeader := r.Header.Get("X-Usage-Tokens")
		if usageHeader != "" {
			tokens, err := strconv.Atoi(usageHeader)
			if err == nil {
				if tokens >= m.maxTokens {
					logger.WarnContext(r.Context(), "Quota exceeded", "tokens", tokens, "max", m.maxTokens)
					http.Error(w, fmt.Sprintf("Quota Monitor: Mission Budget Exceeded (%d tokens)", tokens), http.StatusPaymentRequired)
					return
				}
			} else {
				logger.WarnContext(r.Context(), "Invalid X-Usage-Tokens header format", "header", usageHeader)
			}
		}

		next.ServeHTTP(w, r)
	})
}
