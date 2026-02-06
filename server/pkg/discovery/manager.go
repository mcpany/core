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
// Summary: Tracks the health and execution status of a discovery provider.
type ProviderStatus struct {
	// Name is the unique name of the provider.
	Name            string
	// Status is the current status string (e.g., "OK", "ERROR", "PENDING").
	Status          string
	// LastError holds the error message if the last run failed.
	LastError       string
	// LastRunAt is the timestamp of the last discovery execution.
	LastRunAt       time.Time
	// DiscoveredCount is the number of services discovered in the last run.
	DiscoveredCount int
}

// Manager manages auto-discovery providers.
//
// Summary: Orchestrates the execution and status tracking of multiple discovery providers.
type Manager struct {
	providers []Provider
	mu        sync.RWMutex
	statuses  map[string]*ProviderStatus
}

// NewManager creates a new discovery manager.
//
// Summary: Initializes a new discovery Manager.
//
// Returns:
//   - *Manager: A new, empty Manager instance.
func NewManager() *Manager {
	return &Manager{
		statuses: make(map[string]*ProviderStatus),
	}
}

// RegisterProvider registers a new provider.
//
// Summary: Adds a discovery provider to the manager.
//
// Parameters:
//   - p: Provider. The provider to register.
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
// Summary: Executes all discovery providers and aggregates results.
//
// Parameters:
//   - ctx: context.Context. The context for the discovery operations.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: A list of discovered service configurations.
//
// Side Effects:
//   - Updates the internal status of each provider.
//   - Logs discovery results and errors.
func (m *Manager) Run(ctx context.Context) []*configv1.UpstreamServiceConfig {
	var allServices []*configv1.UpstreamServiceConfig
	log := logging.GetLogger()

	m.mu.RLock()
	providers := make([]Provider, len(m.providers))
	copy(providers, m.providers)
	m.mu.RUnlock()

	for _, p := range providers {
		log.Info("Running auto-discovery", "provider", p.Name())
		services, err := p.Discover(ctx)

		m.mu.Lock()
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
		m.mu.Unlock()
	}

	return allServices
}

// GetStatuses returns the current status of all providers.
//
// Summary: Retrieves the status of all registered providers.
//
// Returns:
//   - []*ProviderStatus: A slice of provider statuses.
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
// Summary: Retrieves the status of a specific provider by name.
//
// Parameters:
//   - name: string. The name of the provider.
//
// Returns:
//   - *ProviderStatus: The status of the provider.
//   - bool: True if the provider was found, false otherwise.
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
