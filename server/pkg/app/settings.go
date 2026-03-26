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
//
// Summary: Thread-safe manager for global application settings.
type GlobalSettingsManager struct {
	mu             sync.RWMutex
	apiKey         atomic.Value // stores string
	allowedIPs     atomic.Value // stores []string
	allowedOrigins atomic.Value // stores []string
}

// NewGlobalSettingsManager creates a new GlobalSettingsManager with initial values.
//
// Summary: Initializes a new settings manager with the provided API key, IP allowlist, and CORS origins.
//
// Parameters:
//   - apiKey: string. The initial API key.
//   - allowedIPs: []string. The initial list of allowed IP addresses.
//   - allowedOrigins: []string. The initial list of allowed CORS origins.
//
// Returns:
//   - *GlobalSettingsManager: The initialized manager.
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
// Summary: Refreshes global settings from the configuration object.
//
// Parameters:
//   - settings: *config_v1.GlobalSettings. The new global settings configuration.
//   - explicitAPIKey: string. An explicitly provided API key (e.g. from CLI flags) that overrides the config.
//
// Returns:
//
//	None.
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
// Summary: Returns the security key currently active for the application, typically used for authentication.
//
// Returns:
//   - string: The API key.
//
// Parameters:
//   - None.
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
// Summary: Returns the list of IP addresses or CIDR blocks permitted to access the application.
//
// Returns:
//   - []string: A list of allowed IP CIDRs or addresses.
//
// Parameters:
//   - None.
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
// Summary: Returns the list of web origins permitted to make cross-origin requests to the application.
//
// Returns:
//   - []string: A list of allowed origins.
//
// Parameters:
//   - None.
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
