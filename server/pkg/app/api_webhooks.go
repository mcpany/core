// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"io"
	"net/http"

	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"google.golang.org/protobuf/encoding/protojson"
)

func (a *Application) listWebhooksHandler(w http.ResponseWriter, r *http.Request) {
	a.systemWebhooksMu.RLock()
	defer a.systemWebhooksMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	// protojson doesn't marshal slices. We construct the array manually or use a wrapper.
	// Manual construction is safest.
	w.Write([]byte("["))
	for i, hook := range a.systemWebhooks {
		if i > 0 {
			w.Write([]byte(","))
		}
		b, err := protojson.Marshal(hook)
		if err != nil {
			logging.GetLogger().Error("Failed to marshal webhook", "error", err)
			continue
		}
		w.Write(b)
	}
	w.Write([]byte("]"))
}

func (a *Application) createWebhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	logging.GetLogger().Info("Webhook body DEBUG MARKER", "body", string(body))

	var req config_v1.SystemWebhookConfig
	if err := protojson.Unmarshal(body, &req); err != nil {
		logging.GetLogger().Error("Failed to decode webhook request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	logging.GetLogger().Info("Received webhook request", "name", req.GetName())

	a.systemWebhooksMu.Lock()
	defer a.systemWebhooksMu.Unlock()

	// Simple validation
	if req.GetName() == "" || req.GetUrlPath() == "" {
		http.Error(w, "Name and URL Path are required", http.StatusBadRequest)
		return
	}

	a.systemWebhooks = append(a.systemWebhooks, &req)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp, _ := protojson.Marshal(&req)
	w.Write(resp)
}
