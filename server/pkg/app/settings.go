// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
// NewGlobalSettingsManager creates a new GlobalSettingsManager with initial values.
//
// Summary: Initializes the global settings manager.
//
// Parameters:
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//   - apiKey: string. The initial API key.
//   - allowedIPs: []string. The initial list of allowed IP addresses.
//   - allowedOrigins: []string. The initial list of allowed CORS origins.
//
// Returns:
//   - *GlobalSettingsManager: The initialized manager.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func NewGlobalSettingsManager(apiKey string, allowedIPs []string, allowedOrigins []string) *GlobalSettingsManager {
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
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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
// GetAPIKey returns the current API key.
//
// Summary: Retrieves the active API key.
//
// Returns:
//   - string: The API key.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (m *GlobalSettingsManager) GetAPIKey() string {
// GetAllowedIPs returns the current allowed IPs.
//
// Summary: Retrieves the list of allowed IP addresses.
//
// Returns:
//   - []string: A list of allowed IP CIDRs or addresses.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (m *GlobalSettingsManager) GetAllowedIPs() []string {
// GetAllowedOrigins returns the current allowed origins.
//
// Summary: Retrieves the list of allowed CORS origins.
//
// Returns:
//   - []string: A list of allowed origins.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (m *GlobalSettingsManager) GetAllowedOrigins() []string {
	val := m.allowedOrigins.Load()
	if val == nil {
		return nil
	}
	return val.([]string)
}
