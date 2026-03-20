package app

import (
	"encoding/json"
	"net/http"

	"github.com/mcpany/core/server/pkg/middleware"
)

// handleDebugSeedToolUsage seeds tool usage stats directly into the Prometheus registry.
//
// Summary: Returns a handler that seeds tool usage data for testing.
//
// Returns:
//   - http.HandlerFunc: The HTTP handler.
//
// Side Effects:
//   - Modifies global prometheus metrics.
func (a *Application) handleDebugSeedToolUsage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var stats []ToolAnalytics
		if err := json.NewDecoder(r.Body).Decode(&stats); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		for _, stat := range stats {
			if stat.TotalCalls > 0 {
				successCount := int(float64(stat.TotalCalls) * (stat.SuccessRate / 100.0))
				errorCount := int(stat.TotalCalls) - successCount

				if successCount > 0 {
					middleware.InjectToolExecutionForTesting(stat.Name, stat.ServiceID, "success", "", successCount)
				}
				if errorCount > 0 {
					middleware.InjectToolExecutionForTesting(stat.Name, stat.ServiceID, "error", "mock_error", errorCount)
				}
			}
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}
}
