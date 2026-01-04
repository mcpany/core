// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/marketplace"
)

func (a *Application) handleMarketplaceSubscriptions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			subs := a.MarketplaceManager.ListSubscriptions()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(subs)
		case http.MethodPost:
			var sub marketplace.Subscription
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &sub); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			created, err := a.MarketplaceManager.AddSubscription(sub)
			if err != nil {
				logging.GetLogger().Error("failed to add subscription", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(created)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleMarketplaceSubscriptionDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/marketplace/subscriptions/")
		if id == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}

		if strings.HasSuffix(id, "/sync") {
			a.handleMarketplaceSubscriptionSync(w, r, strings.TrimSuffix(id, "/sync"))
			return
		}

		switch r.Method {
		case http.MethodGet:
			sub, ok := a.MarketplaceManager.GetSubscription(id)
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(sub)
		case http.MethodPut:
			var sub marketplace.Subscription
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &sub); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			updated, err := a.MarketplaceManager.UpdateSubscription(id, sub)
			if err != nil {
				logging.GetLogger().Error("failed to update subscription", "error", err)
				// differentiate between not found (error) vs internal error if needed, but for now 500 is safe or 404 if error says so
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(updated)
		case http.MethodDelete:
			if err := a.MarketplaceManager.DeleteSubscription(id); err != nil {
				logging.GetLogger().Error("failed to delete subscription", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleMarketplaceSubscriptionSync(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := a.MarketplaceManager.SyncSubscription(r.Context(), id); err != nil {
		logging.GetLogger().Error("failed to sync subscription", "id", id, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{}"))
}
