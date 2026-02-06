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
type ProviderStatus struct {
	Name            string
	Status          string // "OK", "ERROR"
	LastError       string
	LastRunAt       time.Time
	DiscoveredCount int
}

// Manager manages auto-discovery providers.
type Manager struct {
	providers []Provider
	mu        sync.RWMutex
	statuses  map[string]*ProviderStatus
}

// NewManager creates a new discovery manager.
func NewManager() *Manager {
	return &Manager{
		statuses: make(map[string]*ProviderStatus),
	}
}

// RegisterProvider registers a new provider.
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
func (m *Manager) Run(ctx context.Context) []*configv1.UpstreamServiceConfig {
	var allServices []*configv1.UpstreamServiceConfig
	log := logging.GetLogger()

	m.mu.RLock()
	providers := make([]Provider, len(m.providers))
	copy(providers, m.providers)
	m.mu.RUnlock()

	// ⚡ BOLT: Parallelize discovery to avoid blocking on slow providers.
	// Randomized Selection from Top 5 High-Impact Targets
	var wg sync.WaitGroup
	results := make([][]*configv1.UpstreamServiceConfig, len(providers))

	for i, p := range providers {
		wg.Add(1)
		go func(idx int, prov Provider) {
			defer wg.Done()
			log.Info("Running auto-discovery", "provider", prov.Name())
			services, err := prov.Discover(ctx)

			// Update status safely
			m.mu.Lock()
			status := &ProviderStatus{
				Name:      prov.Name(),
				LastRunAt: time.Now(),
			}

			if err != nil {
				log.Warn("Auto-discovery failed", "provider", prov.Name(), "error", err)
				status.Status = "ERROR"
				status.LastError = err.Error()
			} else {
				log.Info("Auto-discovery success", "provider", prov.Name(), "count", len(services))
				status.Status = "OK"
				status.DiscoveredCount = len(services)
				results[idx] = services
			}
			m.statuses[prov.Name()] = status
			m.mu.Unlock()
		}(i, p)
	}

	wg.Wait()

	// Aggregate results in order
	for _, res := range results {
		if res != nil {
			allServices = append(allServices, res...)
		}
	}

	return allServices
}

// GetStatuses returns the current status of all providers.
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
