// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Manager manages incoming system webhooks.
type Manager struct {
	toolManager tool.ManagerInterface
	webhooks    map[string]*configv1.SystemWebhookConfig
	mu          sync.RWMutex
}

// NewManager creates a new Webhook Manager.
func NewManager(toolManager tool.ManagerInterface) *Manager {
	return &Manager{
		toolManager: toolManager,
		webhooks:    make(map[string]*configv1.SystemWebhookConfig),
	}
}

// UpdateConfig updates the webhook configurations.
func (m *Manager) UpdateConfig(configs []*configv1.SystemWebhookConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.webhooks = make(map[string]*configv1.SystemWebhookConfig)
	for _, cfg := range configs {
		if cfg.GetUrlPath() != "" && !cfg.GetDisabled() {
			m.webhooks[cfg.GetUrlPath()] = cfg
		}
	}
}

// Handler returns an HTTP handler that dispatches webhooks.
func (m *Manager) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// The Mux might pass prefix "/webhooks/" so we might need to handle that.
		// For now assume path matches UrlPath exactly or we trim prefix in router.
		// Or we can iterate map (slow).
		// Best practice: The main router handles routing.
		// If we mount this at /webhooks/, then paths inside config should be relative or we match suffix.
		// Let's assume exact match on what's passed to us.

		m.mu.RLock()
		cfg, ok := m.webhooks[path]
		m.mu.RUnlock()

		if !ok {
			// Try matching with /webhooks prefix if not present in map
			// Or check if path passed by mux is stripped?
			// Let's just log and return 404
			logging.GetLogger().Debug("Webhook not found", "path", path)
			http.NotFound(w, r)
			return
		}

		// Validation
		if secret := cfg.GetSecret(); secret != "" {
			// Check query param or Header
			// Common header: X-Webhook-Secret
			headerSecret := r.Header.Get("X-Webhook-Secret")
			querySecret := r.URL.Query().Get("secret")

			if headerSecret != secret && querySecret != secret {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		// Read Body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		defer func() { _ = r.Body.Close() }()

		// Action
		if toolName := cfg.GetTriggerTool(); toolName != "" {
			// Execute Tool
			// We convert the webhook payload to tool input.
			// Assumption: The tool accepts a JSON object.
			// If payload is JSON, pass it as is?
			// If payload is form, convert to JSON?
			// Let's assume JSON for now.

			var toolInputs json.RawMessage
			if len(body) > 0 {
				toolInputs = json.RawMessage(body)
			} else {
				toolInputs = json.RawMessage("{}")
			}

			req := &tool.ExecutionRequest{
				ToolName:   toolName,
				ToolInputs: toolInputs,
			}

			// We need a context. Use request context.
			// But we might want to run this asynchronously?
			// Webhooks usually expect fast response.
			// If tool is slow, webhook might timeout.
			// For now, synchronous.

			result, err := m.toolManager.ExecuteTool(r.Context(), req)
			if err != nil {
				logging.GetLogger().Error("Webhook tool execution failed", "tool", toolName, "error", err)
				http.Error(w, fmt.Sprintf("Tool execution failed: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(result); err != nil {
				logging.GetLogger().Error("Failed to encode webhook response", "error", err)
			}
			return
		}

		// No action configured
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Webhook received (no action)"))
	})
}
