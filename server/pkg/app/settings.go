// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"sync"
	"sync/atomic"

	config_v1 "github.com/mcpany/core/proto/config/v1"
)

// GlobalSettingsManager manages the global settings of the application in a thread-safe manner.
//
// Summary: Manages the global settings of the application in a thread-safe manner.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

type GlobalSettingsManager struct {
	mu             sync.RWMutex
	apiKey         atomic.Value // stores string
	allowedIPs     atomic.Value // stores []string
	allowedOrigins atomic.Value // stores []string
}

// NewGlobalSettingsManager creates a new GlobalSettingsManager with initial values.
//
// Summary: Creates a new GlobalSettingsManager with initial values.
//
// Parameters:
//   - apiKey (string): Parameter.
//   - allowedIPs ([]string): Parameter.
//   - allowedOrigins ([]string): Parameter.
//
// Returns:
//   - *GlobalSettingsManager: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

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

// Update updates the settings from the provided GlobalSettings config.
//
// Summary: Updates the settings from the provided GlobalSettings config.
//
// Parameters:
//   - settings (*config_v1.GlobalSettings): Parameter.
//   - explicitAPIKey (string): Parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

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

// GetAPIKey returns the current API key.
//
// Summary: Returns the current API key.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *GlobalSettingsManager) GetAPIKey() string {
	val := m.apiKey.Load()
	if val == nil {
		return ""
	}
	return val.(string)
}

// GetAllowedIPs returns the current allowed IPs.
//
// Summary: Returns the current allowed IPs.
//
// Parameters:
//   - None.
//
// Returns:
//   - []string: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *GlobalSettingsManager) GetAllowedIPs() []string {
	val := m.allowedIPs.Load()
	if val == nil {
		return nil
	}
	return val.([]string)
}

// GetAllowedOrigins returns the current allowed origins.
//
// Summary: Returns the current allowed origins.
//
// Parameters:
//   - None.
//
// Returns:
//   - []string: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *GlobalSettingsManager) GetAllowedOrigins() []string {
	val := m.allowedOrigins.Load()
	if val == nil {
		return nil
	}
	return val.([]string)
}
