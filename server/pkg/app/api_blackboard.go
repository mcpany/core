package app

import (
	"encoding/json"
	"net/http"

	"github.com/mcpany/core/server/pkg/middleware"
)

// handleBlackboardKeys returns an HTTP handler for retrieving all blackboard keys.
func (a *Application) handleBlackboardKeys() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// For the UI demo we return mocked items since standard global store is initialized differently.
		entries := []middleware.BlackboardEntry{
			{AgentID: "agent-a", Key: "session_token", Value: "abc-123", Intent: "auth"},
			{AgentID: "agent-b", Key: "last_query", Value: "select * from users", Intent: "database_read"},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(entries); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
