package app

import (
	"encoding/json"
	"net/http"

	"github.com/mcpany/core/src/interop"
)

func (a *Application) handleInteropTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var task interop.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	result, err := a.AdapterHub.RouteTask(r.Context(), &task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (a *Application) handleInteropAdapters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Currently AdapterHub doesn't expose ListAdapters, so we'll just return the names
	// of the ones we registered
	adapters := []string{"OpenClaw", "CrewAI", "AutoGen"}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"adapters": adapters,
	})
}
