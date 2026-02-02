// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/topology"
	"google.golang.org/protobuf/encoding/protojson"
)

// SeedStateRequest represents the payload for the debug state seeding endpoint.
type SeedStateRequest struct {
	Services    []json.RawMessage       `json:"services"`
	Secrets     []json.RawMessage       `json:"secrets"`
	Traffic     []topology.TrafficPoint `json:"traffic"`
	Metrics     map[string]any          `json:"metrics"`
	Collections []json.RawMessage       `json:"collections"`
	Settings    json.RawMessage         `json:"settings"`
}

func (a *Application) handleDebugSeedState() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SeedStateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		logger := logging.GetLogger()

		// 1. Seed Global Settings
		if req.Settings != nil {
			var settings configv1.GlobalSettings
			if err := protojson.Unmarshal(req.Settings, &settings); err != nil {
				logger.Error("Failed to unmarshal settings", "error", err)
				http.Error(w, "Invalid settings definition", http.StatusBadRequest)
				return
			}
			if err := a.Storage.SaveGlobalSettings(ctx, &settings); err != nil {
				logger.Error("Failed to save global settings", "error", err)
				http.Error(w, "Failed to save global settings", http.StatusInternalServerError)
				return
			}
		}

		// 2. Seed Services
		if len(req.Services) > 0 {
			for _, raw := range req.Services {
				var svc configv1.UpstreamServiceConfig
				if err := protojson.Unmarshal(raw, &svc); err != nil {
					logger.Error("Failed to unmarshal service", "error", err)
					http.Error(w, "Invalid service definition", http.StatusBadRequest)
					return
				}
				if err := a.Storage.SaveService(ctx, &svc); err != nil {
					logger.Error("Failed to save service", "name", svc.GetName(), "error", err)
					http.Error(w, "Failed to save service", http.StatusInternalServerError)
					return
				}
			}
		}

		// 3. Seed Collections
		if len(req.Collections) > 0 {
			for _, raw := range req.Collections {
				var col configv1.Collection
				if err := protojson.Unmarshal(raw, &col); err != nil {
					logger.Error("Failed to unmarshal collection", "error", err)
					http.Error(w, "Invalid collection definition", http.StatusBadRequest)
					return
				}
				if err := a.Storage.SaveServiceCollection(ctx, &col); err != nil {
					logger.Error("Failed to save collection", "name", col.GetName(), "error", err)
					http.Error(w, "Failed to save collection", http.StatusInternalServerError)
					return
				}
			}
		}

		// 4. Seed Secrets
		if len(req.Secrets) > 0 {
			for _, raw := range req.Secrets {
				var secret configv1.Secret
				if err := protojson.Unmarshal(raw, &secret); err != nil {
					logger.Error("Failed to unmarshal secret", "error", err)
					http.Error(w, "Invalid secret definition", http.StatusBadRequest)
					return
				}
				if err := a.Storage.SaveSecret(ctx, &secret); err != nil {
					logger.Error("Failed to save secret", "id", secret.GetId(), "error", err)
					http.Error(w, "Failed to save secret", http.StatusInternalServerError)
					return
				}
			}
		}

		// 5. Seed Traffic
		if len(req.Traffic) > 0 && a.TopologyManager != nil {
			a.TopologyManager.SeedTrafficHistory(req.Traffic)
		}

		// 6. Seed Metrics Cache
		if len(req.Metrics) > 0 {
			for k, v := range req.Metrics {
				a.setStatsCache(k, v)
			}
		}

		// 7. Reload Config
		if err := a.ReloadConfig(ctx, a.fs, a.configPaths); err != nil {
			logger.Error("Failed to reload config after seed", "error", err)
			http.Error(w, "Failed to reload config", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}
}
