package rest

import (
	"encoding/json"
	"net/http"

	"github.com/mcpany/core/src/interop"
)

// InteropTaskRequest defines the incoming request for task routing
type InteropTaskRequest struct {
	Framework string            `json:"framework"`
	Intent    string            `json:"intent"`
	Payload   map[string]string `json:"payload"`
}

// Global Hub instance for testing purposes
var defaultHub *interop.AdapterHub

func init() {
	defaultHub = interop.NewAdapterHub()
	defaultHub.RegisterAdapter(interop.NewOpenClawAdapter())
	defaultHub.RegisterAdapter(interop.NewCrewAIAdapter())
	defaultHub.RegisterAdapter(interop.NewAutoGenAdapter())
}

func InteropTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req InteropTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	task := &interop.Task{
		ID:        "web-task",
		Framework: req.Framework,
		Intent:    req.Intent,
		Payload:   req.Payload,
	}

	res, err := defaultHub.RouteTask(r.Context(), task)
	if err != nil {
		respondWithJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
