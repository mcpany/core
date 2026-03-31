package app

import (
	"encoding/json"
	"net/http"

	"github.com/mcpany/core/src/interop"
)

// handleInterop routes interop tasks and tests integration for src/interop.
func (a *Application) handleInterop() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var task interop.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		hub := interop.NewAdapterHub()
		hub.RegisterAdapter(interop.NewOpenClawAdapter())
		hub.RegisterAdapter(interop.NewCrewAIAdapter())
		hub.RegisterAdapter(interop.NewAutoGenAdapter())

		res, err := hub.RouteTask(r.Context(), &task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}
