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

// ProviderStatus represents the public ProviderStatus entity.
//
// Summary: Defines the structured data model representing a status.
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
type ProviderStatus struct {
	Name            string
	Status          string // "OK", "ERROR"
	LastError       string
	LastRunAt       time.Time
	DiscoveredCount int
}

// Manager represents the public Manager entity.
//
// Summary: Coordinates operations and orchestrates lifecycle events for the  components.
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
type Manager struct {
	providers []Provider
	mu        sync.RWMutex
	statuses  map[string]*ProviderStatus
}

// NewManager serves as a public interface for interacting with NewManager.
//
// Summary: Constructs and returns an initialized manager ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewManager() *Manager {
	return &Manager{
		statuses: make(map[string]*ProviderStatus),
	}
}

// RegisterProvider serves as a public interface for interacting with RegisterProvider.
//
// Summary: Register the provider appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *Manager) RegisterProvider(p Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers = append(m.providers, p)
	m.statuses[p.Name()] = &ProviderStatus{
		Name:   p.Name(),
		Status: "PENDING",
	}
}

// Run serves as a public interface for interacting with Run.
//
// Summary: Run the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// GetStatuses serves as a public interface for interacting with GetStatuses.
//
// Summary: Fetches and returns the underlying statuses from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// GetProviderStatus serves as a public interface for interacting with GetProviderStatus.
//
// Summary: Fetches and returns the underlying provider status from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
