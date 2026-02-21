// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mcpany/core/server/pkg/webhooks"
)

func (a *Application) handleWebhooks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			list := a.WebhooksManager.ListWebhooks()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(list)
		case http.MethodPost:
			var cfg webhooks.WebhookConfig
			if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if cfg.URL == "" {
				http.Error(w, "url is required", http.StatusBadRequest)
				return
			}
			a.WebhooksManager.AddWebhook(&cfg)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(cfg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleWebhookDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/webhooks/")
		parts := strings.Split(path, "/")
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}
		id := parts[0]

		if len(parts) > 1 && parts[1] == "test" {
			a.handleWebhookTest(w, r, id)
			return
		}

		switch r.Method {
		case http.MethodGet:
			wh, ok := a.WebhooksManager.GetWebhook(id)
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(wh)
		case http.MethodDelete:
			a.WebhooksManager.DeleteWebhook(id)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleWebhookTest(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := a.WebhooksManager.TestWebhook(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"success"}`))
}
