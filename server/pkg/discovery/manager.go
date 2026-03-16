// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"
	"sync"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
)

// ProviderStatus represents the status of a discovery provider.
//
// Summary: ProviderStatus represents the status of a discovery provider.
//
// Summary: ProviderStatus represents the status of a discovery provider.
type ProviderStatus struct {
	Name            string
	Status          string // "OK", "ERROR"
	LastError       string
	LastRunAt       time.Time
	DiscoveredCount int
// Manager manages auto-discovery providers.
//
// Summary: Manager manages auto-discovery providers.

// Manager manages auto-discovery providers.
//
// Summary: Manager manages auto-discovery providers.
type Manager struct {
	providers []Provider
// NewManager creates a new discovery manager.
//
// Summary: NewManager creates a new discovery manager.
//
// Parameters:
//   - None.
//
// Returns:
//   - *Manager: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// RegisterProvider registers a new provider.
//
// Summary: RegisterProvider registers a new provider.
//
// Parameters:
//   - p (Provider): The provided p data.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (m *Manager) RegisterProvider(p Provider) {
	m.mu.Lock()
// Run runs all registered providers and returns the aggregated discovered services. It also updates the internal status of each provider.
//
// Summary: Run runs all registered providers and returns the aggregated discovered services. It also updates the internal status of each provider.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - ctx (context.Context): The cancellation and deadline context.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (m *Manager) Run(ctx context.Context) []*configv1.UpstreamServiceConfig {
	var allServices []*configv1.UpstreamServiceConfig
	log := logging.GetLogger()

	m.mu.RLock()
	providers := make([]Provider, len(m.providers))
	copy(providers, m.providers)
	m.mu.RUnlock()

	// ⚡ BOLT: Parallelized discovery providers to avoid waterfall latency.
	// Randomized Selection from Top 5 High-Impact Targets
	var wg sync.WaitGroup
	wg.Add(len(providers))

	for _, p := range providers {
		go func(p Provider) {
			defer wg.Done()
			log.Info("Running auto-discovery", "provider", p.Name())
			services, err := p.Discover(ctx)

			m.mu.Lock()
			defer m.mu.Unlock()

			status := &ProviderStatus{
				Name:      p.Name(),
				LastRunAt: time.Now(),
			}

			if err != nil {
				log.Warn("Auto-discovery failed", "provider", p.Name(), "error", err)
				status.Status = "ERROR"
				status.LastError = err.Error()
			} else {
				log.Info("Auto-discovery success", "provider", p.Name(), "count", len(services))
				status.Status = "OK"
				status.DiscoveredCount = len(services)
				allServices = append(allServices, services...)
// GetStatuses returns the current status of all providers.
//
// Summary: GetStatuses returns the current status of all providers.
//
// Parameters:
//   - None.
//
// Returns:
//   - []*ProviderStatus: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Parameters:
//   - None.
//
// Returns:
//   - []*ProviderStatus: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (m *Manager) GetStatuses() []*ProviderStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
// GetProviderStatus returns the status of a specific provider.
//
// Summary: GetProviderStatus returns the status of a specific provider.
//
// Parameters:
//   - name (string): The human-readable or system name.
//
// Returns:
//   - *ProviderStatus: The resulting object or data structure.
//   - bool: True if successful or valid, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Summary: GetProviderStatus returns the status of a specific provider.
//
// Parameters:
//   - name (string): The human-readable or system name.
//
// Returns:
//   - *ProviderStatus: The resulting object or data structure.
//   - bool: True if successful or valid, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (m *Manager) GetProviderStatus(name string) (*ProviderStatus, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.statuses[name]
	if !ok {
		return nil, false
	}
	sCopy := *s
	return &sCopy, true
}
