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
// Summary: Represents a ProviderStatus.
type ProviderStatus struct {
	Name            string
	Status          string // "OK", "ERROR"
	LastError       string
	LastRunAt       time.Time
	DiscoveredCount int
}

// Manager manages auto-discovery providers.
//
// Summary: Represents a Manager.
type Manager struct {
	providers []Provider
	mu        sync.RWMutex
	statuses  map[string]*ProviderStatus
}

// NewManager creates a new manager.
//
// Summary: Creates a new manager.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - *Manager: The result.
func NewManager() *Manager {
	return &Manager{
		statuses: make(map[string]*ProviderStatus),
	}
}

// RegisterProvider registerProvider register provider.
//
// Summary: RegisterProvider register provider.
//
// Parameters: - None.
//   - p (Provider): The p.
//
// Returns: - None.
//   - None.
func (m *Manager) RegisterProvider(p Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers = append(m.providers, p)
	m.statuses[p.Name()] = &ProviderStatus{
		Name:   p.Name(),
		Status: "PENDING",
	}
}

// Run run run.
//
// Summary: Run run.
//
// Parameters: - None.
//   - ctx (context.Context): The context for the request.
//
// Returns: - None.
//   - []*configv1.UpstreamServiceConfig: The result.
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

// GetStatuses retrieves the statuses.
//
// Summary: Retrieves the statuses.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - []*ProviderStatus: The result.
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

// GetProviderStatus retrieves the provider status.
//
// Summary: Retrieves the provider status.
//
// Parameters: - None.
//   - name (string): The name.
//
// Returns: - None.
//   - *ProviderStatus: The result.
//   - bool: The result.
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
