// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mcpany/core/server/pkg/alerts"
)

func (a *Application) handleAlerts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			list := a.AlertsManager.ListAlerts()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(list)
		case http.MethodPost:
			var alert alerts.Alert
			if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			created := a.AlertsManager.CreateAlert(&alert)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(created)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleAlertWebhook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			url := a.AlertsManager.GetWebhookURL()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"url": url})
		case http.MethodPost:
			var body struct {
				URL string `json:"url"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			a.AlertsManager.SetWebhookURL(body.URL)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"url": body.URL})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleAlertDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/alerts/")
		if id == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			alert := a.AlertsManager.GetAlert(id)
			if alert == nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(alert)
		case http.MethodPatch:
			var updates alerts.Alert
			if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			updated := a.AlertsManager.UpdateAlert(id, &updates)
			if updated == nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(updated)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleAlertRules() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			list := a.AlertsManager.ListRules()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(list)
		case http.MethodPost:
			var rule alerts.AlertRule
			if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			created := a.AlertsManager.CreateRule(&rule)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(created)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleAlertRuleDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/alerts/rules/")
		if id == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			rule := a.AlertsManager.GetRule(id)
			if rule == nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(rule)
		case http.MethodPut:
			var updates alerts.AlertRule
			if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			updated := a.AlertsManager.UpdateRule(id, &updates)
			if updated == nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(updated)
		case http.MethodDelete:
			if err := a.AlertsManager.DeleteRule(id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
