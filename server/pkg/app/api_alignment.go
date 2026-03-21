// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
)

// SubagentStatus defines the structure for AIA heartbeat responses.
type SubagentStatus struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Status       string  `json:"status"`
	EntropyScore float64 `json:"entropyScore"`
	LastHeartbeat int64  `json:"lastHeartbeat"`
}

var currentAIAStatus = []SubagentStatus{
	{ID: "sa-101", Name: "Data Extractor", Status: "aligned", EntropyScore: 12.0, LastHeartbeat: 0},
	{ID: "sa-204", Name: "Code Synthesizer", Status: "drifting", EntropyScore: 68.0, LastHeartbeat: 0},
	{ID: "sa-092", Name: "Schema Validator", Status: "aligned", EntropyScore: 5.0, LastHeartbeat: 0},
	{ID: "sa-881", Name: "Network Probe", Status: "hijacked", EntropyScore: 94.0, LastHeartbeat: 0},
}

// handleActiveIntentAlignment handles the /api/v1/alignment/status endpoint.
//
// Summary: Retrieves the current status of AIA heartbeats and semantic drift across subagents.
//
// Parameters:
//   - w (http.ResponseWriter): The response writer.
//   - r (*http.Request): The HTTP request.
//
// Returns:
//   - None
//
// Errors:
//   - Returns 500 if JSON encoding fails.
//
// Side Effects:
//   - None.
func (a *Application) handleActiveIntentAlignment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.GetLogger()

		// Simulate dynamic changes
		now := time.Now().UnixMilli()
		for i := range currentAIAStatus {
			jitter := (rand.Float64() * 10) - 5
			newEntropy := currentAIAStatus[i].EntropyScore + jitter
			if newEntropy < 0 {
				newEntropy = 0
			}
			if newEntropy > 100 {
				newEntropy = 100
			}

			currentAIAStatus[i].EntropyScore = newEntropy

			if newEntropy > 85 {
				currentAIAStatus[i].Status = "hijacked"
			} else if newEntropy > 50 {
				currentAIAStatus[i].Status = "drifting"
			} else {
				currentAIAStatus[i].Status = "aligned"
			}

			currentAIAStatus[i].LastHeartbeat = now - int64(rand.Intn(500))
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(currentAIAStatus); err != nil {
			log.Error("Failed to encode alignment status", "error", err)
			http.Error(w, "Failed to encode status", http.StatusInternalServerError)
		}
	}
}
