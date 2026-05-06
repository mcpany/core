package app

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mcpany/core/server/pkg/logging"
)

// EntropyScoreResponse represents the response for an entropy score request.
type EntropyScoreResponse struct {
	SessionID     string  `json:"sessionId"`
	CoherenceScore float64 `json:"coherenceScore"`
	Status        string  `json:"status"`
}

// EntropyGateRequest represents the request payload to configure an entropy gate.
type EntropyGateRequest struct {
	Threshold float64 `json:"threshold"`
	Action    string  `json:"action"`
}

// handleGetEntropy handles the GET /api/v1/entropy/{session_id} endpoint.
func (a *Application) handleGetEntropy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sessionID := vars["session_id"]
		if sessionID == "" {
			http.Error(w, "session_id is required", http.StatusBadRequest)
			return
		}

		response := EntropyScoreResponse{
			SessionID:     sessionID,
			CoherenceScore: 0.85,
			Status:        "aligned",
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logging.GetLogger().Error("Failed to encode entropy response", "error", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

// handlePostEntropyGate handles the POST /api/v1/policy/entropy-gate endpoint.
func (a *Application) handlePostEntropyGate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req EntropyGateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if req.Threshold < 0 || req.Threshold > 1 {
			http.Error(w, "Threshold must be between 0.0 and 1.0", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"status": "accepted"}`))
	}
}
