package app

import (
	"encoding/json"
	"net/http"

	"github.com/mcpany/core/server/pkg/middleware"
)

// SeedToolUsageRequest represents the body of the seeding request
type SeedToolUsageRequest struct {
	Count       int     `json:"count"`
	SuccessRate float64 `json:"successRate"`
	ToolName    string  `json:"toolName,omitempty"`
}

// SeedToolUsageHandler creates a handler for the debug endpoint to seed metrics
func SeedToolUsageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SeedToolUsageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Calculate successes and failures based on the success rate
		successes := int(float64(req.Count) * (req.SuccessRate / 100.0))
		if req.SuccessRate <= 1.0 { // If it was passed as a fraction like 0.85 instead of 85.0
			successes = int(float64(req.Count) * req.SuccessRate)
		}
		failures := req.Count - successes

		toolName := req.ToolName
		if toolName == "" {
			toolName = "builtin.mcp:list_roots" // default test tool
		}

		// Inject metrics
		middleware.InjectToolExecutionForTesting(toolName, "test-service", successes, failures)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "success",
			"seeded":    req.Count,
			"successes": successes,
			"failures":  failures,
			"tool":      toolName,
		})
	}
}
