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
// Summary: Status of a discovery provider.
type ProviderStatus struct {
	// Name is the name of the provider.
	Name string
	// Status indicates the current state (e.g., "OK", "ERROR", "PENDING").
	Status string
	// LastError contains the error message if the last run failed.
	LastError string
	// LastRunAt is the timestamp of the last discovery attempt.
	LastRunAt time.Time
	// DiscoveredCount is the number of services discovered in the last run.
	DiscoveredCount int
}

// Manager manages auto-discovery providers.
//
// It handles registration of providers and orchestrates their execution to find
// local or remote services.
type Manager struct {
	providers []Provider
	mu        sync.RWMutex
	statuses  map[string]*ProviderStatus
}

// NewManager creates a new discovery manager.
//
// Summary: Initializes a new discovery manager.
//
// Returns:
//   - *Manager: The initialized manager.
func NewManager() *Manager {
	return &Manager{
		statuses: make(map[string]*ProviderStatus),
	}
}

// RegisterProvider registers a new provider.
//
// Summary: Adds a discovery provider.
//
// Parameters:
//   - p: Provider. The provider instance to register.
func (m *Manager) RegisterProvider(p Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers = append(m.providers, p)
	m.statuses[p.Name()] = &ProviderStatus{
		Name:   p.Name(),
		Status: "PENDING",
	}
}

// Run runs all registered providers and returns the aggregated discovered services.
// It also updates the internal status of each provider.
//
// Summary: Executes all discovery providers.
//
// Parameters:
//   - ctx: context.Context. The context for the discovery operations.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: A list of discovered service configurations.
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
			}
			m.statuses[p.Name()] = status
		}(p)
	}

	wg.Wait()

	return allServices
}

// GetStatuses returns the current status of all providers.
//
// Summary: Retrieves provider statuses.
//
// Returns:
//   - []*ProviderStatus: A list of status objects for all registered providers.
func (m *Manager) GetStatuses() []*ProviderStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make([]*ProviderStatus, 0, len(m.statuses))
	for _, s := range m.statuses {
		// Copy to avoid race conditions if caller modifies it (though we return pointers to structs created in Run)
		// But map iteration order is random, maybe sort?
		// For now just return list.
		sCopy := *s
		statuses = append(statuses, &sCopy)
	}
	return statuses
}

// GetProviderStatus returns the status of a specific provider.
//
// Summary: Retrieves the status of a specific provider.
//
// Parameters:
//   - name: string. The name of the provider.
//
// Returns:
//   - *ProviderStatus: The status of the provider.
//   - bool: True if the provider exists, false otherwise.
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
