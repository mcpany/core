// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/llm"
	"github.com/mcpany/core/server/pkg/logging"
)

// AutoCraftRequest represents a request to craft an MCP server config.
type AutoCraftRequest struct {
	bus.BaseMessage
	ID           string           `json:"id"`
	CreatedAt    time.Time        `json:"createdAt"`
	UsersGoal    string           `json:"usersGoal"`
	ServiceName  string           `json:"serviceName"` // Optional, if known
	ProviderType llm.ProviderType `json:"providerType"`
}

// AutoCraftResult represents the result of an auto-craft job.
type AutoCraftResult struct {
	bus.BaseMessage
	Status       string   `json:"status"` // "success", "failed", "in_progress"
	ConfigJSON   string   `json:"configJson,omitempty"`
	ErrorMessage string   `json:"errorMessage,omitempty"`
	Logs         []string `json:"logs,omitempty"`
}

// StartAutoCraftWorker starts the worker for auto-crafting.
// In Phase 1, this is a mock implementation.
func (w *Worker) StartAutoCraftWorker(ctx context.Context, llmStore *llm.ProviderStore) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		log := logging.GetLogger()

		// Subscribe to AutoCraft requests
		// Note: We need to define this topic in pkg/bus or just use a string if allowed.
		// Assuming "auto_craft_request" for now.
		reqBus, err := bus.GetBus[*AutoCraftRequest](w.busProvider, "auto_craft_request")
		if err != nil {
			log.Error("Failed to get auto craft request bus", "error", err)
			return
		}

		resBus, err := bus.GetBus[*AutoCraftResult](w.busProvider, "auto_craft_result")
		if err != nil {
			log.Error("Failed to get auto craft result bus", "error", err)
			return
		}

		unsubscribe := reqBus.Subscribe(ctx, "auto_craft_request", func(req *AutoCraftRequest) {
			w.pond.Submit(func() {
				w.processAutoCraftRequest(ctx, req, llmStore, resBus)
			})
		})

		w.mu.Lock()
		w.stopFuncs = append(w.stopFuncs, unsubscribe)
		w.mu.Unlock()
	}()
}

func (w *Worker) processAutoCraftRequest(ctx context.Context, req *AutoCraftRequest, store *llm.ProviderStore, resBus bus.Bus[*AutoCraftResult]) {
	log := logging.GetLogger()
	log.Info("Processing Auto Craft Request", "goal", req.UsersGoal)

	// Phase 1: Verify we have a key
	config, ok := store.GetConfig(req.ProviderType)
	if !ok || config.APIKey == "" {
		w.sendResult(ctx, resBus, req.CorrelationID(), &AutoCraftResult{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("No API key found for provider %s. Please configure it in Settings.", req.ProviderType),
		})
		return
	}

	// Phase 1: Mock "Researching"
	// Send "In Progress" update (optional, but good for UI)
	w.sendResult(ctx, resBus, req.CorrelationID(), &AutoCraftResult{
		Status: "in_progress",
		Logs:   []string{"Agent started...", "Searching for API documentation...", "Analyzing authentication methods..."},
	})

	time.Sleep(2 * time.Second) // Simulate work

	// Phase 1: Return a static success
	mockConfig := fmt.Sprintf(`{
  "name": "%s-generated",
  "version": "1.0.0",
  "commandLineService": {
    "command": "npx",
    "args": ["-y", "@modelcontextprotocol/server-%s"],
    "env": {
      "API_KEY": "YOUR_KEY_HERE"
    }
  }
}`, req.ServiceName, req.ServiceName)

	if req.ServiceName == "" {
		mockConfig = `{
  "name": "auto-crafted-service",
  "version": "1.0.0",
  "description": "Generated from goal: ` + req.UsersGoal + `"
}`
	}

	w.sendResult(ctx, resBus, req.CorrelationID(), &AutoCraftResult{
		Status:     "success",
		ConfigJSON: mockConfig,
		Logs:       []string{"Found documentation.", " identified Bearer Token auth.", "Generated configuration."},
	})
}

func (w *Worker) sendResult(ctx context.Context, b bus.Bus[*AutoCraftResult], cid string, res *AutoCraftResult) {
	res.BaseMessage = bus.BaseMessage{CID: cid}
	if err := b.Publish(ctx, "auto_craft_result", res); err != nil {
		logging.GetLogger().Error("Failed to publish auto craft result", "error", err)
	}
}
