// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/llm"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/worker"
)

// handleLLMProviders handles saving and retrieving LLM provider configurations.
func (a *Application) handleLLMProviders(store *llm.ProviderStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			configs := store.ListConfigs()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(configs)

		case http.MethodPost:
			var config llm.ProviderConfig
			// Limit body size to 1MB
			body, err := readBodyWithLimit(w, r, 1<<20)
			if err != nil {
				return
			}

			if err := json.Unmarshal(body, &config); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if err := store.SaveConfig(config); err != nil {
				logging.GetLogger().Error("failed to save LLM config", "error", err)
				http.Error(w, "Failed to save configuration", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// AutoCraftJobRequest represents the payload to start an auto-craft job.
type AutoCraftJobRequest struct {
	ServiceName  string           `json:"serviceName"`
	ProviderType llm.ProviderType `json:"providerType"`
	Goal         string           `json:"goal"`
}

// AutoCraftJobResponse represents the response after submitting a job.
type AutoCraftJobResponse struct {
	JobID string `json:"jobId"`
}

// handleAutoCraftJobs handles submission of auto-craft jobs.
func (a *Application) handleAutoCraftJobs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req AutoCraftJobRequest
		body, err := readBodyWithLimit(w, r, 1<<20)
		if err != nil {
			return
		}

		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.ServiceName == "" || req.Goal == "" {
			http.Error(w, "serviceName and goal are required", http.StatusBadRequest)
			return
		}

		// Create a unique ID for the job
		jobID := uuid.New().String()
		acReq := &worker.AutoCraftRequest{
			ID:           jobID,
			ServiceName:  req.ServiceName,  // Corrected from jobReq.ServiceName
			ProviderType: req.ProviderType, // Corrected from jobReq.ProviderType
			UsersGoal:    req.Goal,         // Corrected from jobReq.Goal
			CreatedAt:    time.Now(),
			BaseMessage:  bus.BaseMessage{CID: jobID},
		}

		// Publish to Bus
		reqBus, err := bus.GetBus[*worker.AutoCraftRequest](a.busProvider, "auto_craft_request")
		if err != nil {
			logging.GetLogger().Error("failed to get auto_craft_request bus", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Publish the request to the bus
		if err := reqBus.Publish(r.Context(), "auto_craft_request", acReq); err != nil {
			logging.GetLogger().Error("failed to publish auto craft request", "error", err)
			http.Error(w, "Failed to submit job", http.StatusInternalServerError)
			return
		}

		// Respond with the job ID
		resp := AutoCraftJobResponse{JobID: jobID}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func (a *Application) handleGetAutoCraftJob() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse Job ID from URL: /autocraft/jobs/{id}
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) == 0 {
			http.Error(w, "Invalid job ID", http.StatusBadRequest)
			return
		}
		jobID := parts[len(parts)-1]

		val, ok := a.AutoCraftJobs.Load(jobID)
		if !ok {
			// Phase 1: If not in memory, it might not exist or server restarted.
			// Return 404.
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(val)
	}
}
