// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
)

// WebhookConfig represents the JSON structure expected by the UI.
type WebhookConfig struct {
	ID            string   `json:"id"`
	URL           string   `json:"url"`
	Events        []string `json:"events"`
	Active        bool     `json:"active"`
	LastTriggered string   `json:"last_triggered,omitempty"`
	Status        string   `json:"status,omitempty"`
}

// WebhookHandler handles webhook API requests.
type WebhookHandler struct {
	store config.Store
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(store config.Store) *WebhookHandler {
	return &WebhookHandler{store: store}
}

// ListWebhooks handles GET /api/v1/webhooks and GET /api/settings/webhooks
func (wh *WebhookHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	cfg, err := wh.store.Load(r.Context())
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}

	webhooks := []WebhookConfig{}

	for _, svc := range cfg.GetUpstreamServices() {
		for _, hook := range svc.GetPreCallHooks() {
			if hook.GetWebhook() != nil {
				active := true // Assume active unless we have a field for it
				webhooks = append(webhooks, WebhookConfig{
					ID:     hook.GetName(),
					URL:    hook.GetWebhook().GetUrl(),
					Events: []string{"pre_call"},
					Active: active,
					Status: "active",
				})
			}
		}
		for _, hook := range svc.GetPostCallHooks() {
			if hook.GetWebhook() != nil {
				active := true
				webhooks = append(webhooks, WebhookConfig{
					ID:     hook.GetName(),
					URL:    hook.GetWebhook().GetUrl(),
					Events: []string{"post_call"},
					Active: active,
					Status: "active",
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhooks)
}

// AddWebhook handles POST /api/v1/webhooks
func (wh *WebhookHandler) AddWebhook(w http.ResponseWriter, r *http.Request) {
	var req WebhookConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	cfg, err := wh.store.Load(r.Context())
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}

	// We'll attach it to the first upstream service as a pre_call_hook for now, or create a default service
	services := cfg.GetUpstreamServices()
	var targetService *configv1.UpstreamServiceConfig

	if len(services) > 0 {
		targetService = services[0]
	} else {
		// Cannot add webhook if no services exist in the current configuration model,
		// but let's assume one exists or we return an error.
		http.Error(w, "No upstream services configured to attach webhook to", http.StatusBadRequest)
		return
	}

	hookName := "webhook-" + uuid.New().String()[:8]
	newHook := configv1.CallHook_builder{
		Name: &hookName,
	}.Build()
	newHook.SetWebhook(configv1.WebhookConfig_builder{
		Url: req.URL,
	}.Build())

	// Default to pre_call if events not specified or contains "all" or "pre_call"
	isPostCall := false
	for _, ev := range req.Events {
		if ev == "post_call" {
			isPostCall = true
		}
	}

	if isPostCall {
		targetService.SetPostCallHooks(append(targetService.GetPostCallHooks(), newHook))
	} else {
		targetService.SetPreCallHooks(append(targetService.GetPreCallHooks(), newHook))
	}

	// Use type assertion to check if store supports SaveService
	if saveStore, ok := wh.store.(interface {
		SaveService(context.Context, *configv1.UpstreamServiceConfig) error
	}); ok {
		if err := saveStore.SaveService(r.Context(), targetService); err != nil {
			http.Error(w, "Failed to save config", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Store does not support saving services", http.StatusNotImplemented)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "id": hookName})
}

// DeleteWebhook handles DELETE /api/v1/webhooks/:id
func (wh *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	id := parts[4]

	cfg, err := wh.store.Load(r.Context())
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}

	var targetService *configv1.UpstreamServiceConfig
	var found bool

	for _, svc := range cfg.GetUpstreamServices() {
		var newPre []*configv1.CallHook
		for _, hook := range svc.GetPreCallHooks() {
			if hook.GetName() == id {
				found = true
				targetService = svc
			} else {
				newPre = append(newPre, hook)
			}
		}

		var newPost []*configv1.CallHook
		for _, hook := range svc.GetPostCallHooks() {
			if hook.GetName() == id {
				found = true
				targetService = svc
			} else {
				newPost = append(newPost, hook)
			}
		}

		if found {
			targetService.SetPreCallHooks(newPre)
			targetService.SetPostCallHooks(newPost)
			break
		}
	}

	if !found {
		http.Error(w, "Webhook not found", http.StatusNotFound)
		return
	}

	if saveStore, ok := wh.store.(interface {
		SaveService(context.Context, *configv1.UpstreamServiceConfig) error
	}); ok {
		if err := saveStore.SaveService(r.Context(), targetService); err != nil {
			http.Error(w, "Failed to save config", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// TestWebhook handles POST /api/v1/webhooks/:id/test
func (wh *WebhookHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 || parts[5] != "test" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	id := parts[4]

	cfg, err := wh.store.Load(r.Context())
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}

	var url string
	for _, svc := range cfg.GetUpstreamServices() {
		for _, hook := range svc.GetPreCallHooks() {
			if hook.GetName() == id && hook.GetWebhook() != nil {
				url = hook.GetWebhook().GetUrl()
				break
			}
		}
		for _, hook := range svc.GetPostCallHooks() {
			if hook.GetName() == id && hook.GetWebhook() != nil {
				url = hook.GetWebhook().GetUrl()
				break
			}
		}
	}

	if url == "" {
		http.Error(w, "Webhook not found", http.StatusNotFound)
		return
	}

	// Create a dummy CloudEvent for testing
	payload := `{"id":"test-123","source":"mcpany-dashboard","type":"com.mcpany.webhook.test","data":{"message":"Ping from MCP Any Dashboard!"}}`

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/cloudevents+json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Test request failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		http.Error(w, "Webhook returned error status", http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
