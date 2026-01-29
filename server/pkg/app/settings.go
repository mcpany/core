// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"sync"
	"sync/atomic"

	config_v1 "github.com/mcpany/core/proto/config/v1"
)

// GlobalSettingsManager manages the global settings of the application in a thread-safe manner.
// It allows for dynamic updates to configuration values that are used across the application.
type GlobalSettingsManager struct {
	mu            sync.RWMutex
	apiKey        atomic.Value // stores string
	allowedIPs    atomic.Value // stores []string
	allowedOrigins atomic.Value // stores []string
}

// NewGlobalSettingsManager creates a new GlobalSettingsManager with initial values.
//
// Parameters:
//   - apiKey: The initial API key.
//   - allowedIPs: A list of allowed IP addresses.
//   - allowedOrigins: A list of allowed CORS origins.
//
// Returns:
//   - (*GlobalSettingsManager): A new instance of GlobalSettingsManager.
func NewGlobalSettingsManager(apiKey string, allowedIPs []string, allowedOrigins []string) *GlobalSettingsManager {
	m := &GlobalSettingsManager{}
	m.apiKey.Store(apiKey)
	m.allowedIPs.Store(allowedIPs)
	// If allowedOrigins is nil/empty and not initialized, we might want defaults.
	// But caller handles defaults.
	if allowedOrigins == nil {
		allowedOrigins = []string{}
	}
	m.allowedOrigins.Store(allowedOrigins)
	return m
}

// Update updates the internal settings based on the provided configuration and explicit overrides.
// It prioritizes the explicit API key over the configuration if provided.
//
// Parameters:
//   - settings: The global settings configuration.
//   - explicitAPIKey: An optional explicit API key (e.g., from CLI flags).
func (m *GlobalSettingsManager) Update(settings *config_v1.GlobalSettings, explicitAPIKey string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// API Key priority: Explicit arg > Config
	key := explicitAPIKey
	if key == "" && settings != nil {
		key = settings.GetApiKey()
	}
	m.apiKey.Store(key)

	var ips []string
	if settings != nil {
		ips = settings.GetAllowedIps()
	}
	m.allowedIPs.Store(ips)

	// Origins logic from server.go
	var origins []string
	if settings != nil {
		origins = settings.GetAllowedOrigins()
		if len(origins) == 0 && settings.GetLogLevel() == config_v1.GlobalSettings_LOG_LEVEL_DEBUG {
			origins = []string{"*"}
		}
	}
	m.allowedOrigins.Store(origins)
}

// GetAPIKey retrieves the current API key.
//
// Returns:
//   - (string): The active API key.
func (m *GlobalSettingsManager) GetAPIKey() string {
	val := m.apiKey.Load()
	if val == nil {
		return ""
	}
	return val.(string)
}

// GetAllowedIPs retrieves the current list of allowed IP addresses.
//
// Returns:
//   - ([]string): A slice of allowed IP addresses.
func (m *GlobalSettingsManager) GetAllowedIPs() []string {
	val := m.allowedIPs.Load()
	if val == nil {
		return nil
	}
	return val.([]string)
}

// GetAllowedOrigins retrieves the current list of allowed CORS origins.
//
// Returns:
//   - ([]string): A slice of allowed origins.
func (m *GlobalSettingsManager) GetAllowedOrigins() []string {
	val := m.allowedOrigins.Load()
	if val == nil {
		return nil
	}
	return val.([]string)
}
