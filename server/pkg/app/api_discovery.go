// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/mcpany/core/server/pkg/logging"
)

// handleDiscoveryStatus returns the status of auto-discovery providers.
<<<<<<< HEAD
//
// Summary: Returns discovery status.
//
// Parameters:
//   - w (http.ResponseWriter): The response writer.
//   - r (*http.Request): The HTTP request.
//
// Returns:
//   - None.
//
// Errors:
//   - Writes HTTP errors on failure.
//
// Side Effects:
//   - None.
=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
func (a *Application) handleDiscoveryStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if a.DiscoveryManager == nil {
		http.Error(w, "discovery manager not initialized", http.StatusServiceUnavailable)
		return
	}

	statuses := a.DiscoveryManager.GetStatuses()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(statuses); err != nil {
		logging.GetLogger().Error("Failed to encode discovery status", "error", err)
	}
}

// handleDiscoveryTrigger triggers a new auto-discovery run.
<<<<<<< HEAD
//
// Summary: Triggers discovery.
//
// Parameters:
//   - w (http.ResponseWriter): The response writer.
//   - r (*http.Request): The HTTP request.
//
// Returns:
//   - None.
//
// Errors:
//   - Writes HTTP errors on failure.
//
// Side Effects:
//   - Starts async discovery process.
=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
func (a *Application) handleDiscoveryTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if a.DiscoveryManager == nil {
		http.Error(w, "discovery manager not initialized", http.StatusServiceUnavailable)
		return
	}

	// Trigger async
	go func() {
		log := logging.GetLogger()
		log.Info("Manual discovery triggered via API")
		discovered := a.DiscoveryManager.Run(context.Background())

		if a.ServiceRegistry != nil {
			for _, svc := range discovered {
				log.Info("Registering discovered service", "name", svc.GetName())
				if _, _, _, err := a.ServiceRegistry.RegisterService(context.Background(), svc); err != nil {
					log.Error("Failed to register discovered service", "name", svc.GetName(), "error", err)
				}
			}
		}
	}()

	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte("{}"))
}
