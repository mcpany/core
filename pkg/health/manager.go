// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/alexliesenfeld/health"
	pb_admin "github.com/mcpany/core/proto/admin/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// ServiceHealthState holds the current health state of a service.
type ServiceHealthState struct {
	Status        pb_admin.ServiceStatus
	LastError     string
	LastCheckTime time.Time
}

// Manager handles health checking for services.
type Manager struct {
	mu       sync.RWMutex
	checkers map[string]health.Checker
	statuses map[string]*ServiceHealthState
	stopCh   chan struct{}
}

// NewManager creates a new health manager.
func NewManager() *Manager {
	return &Manager{
		checkers: make(map[string]health.Checker),
		statuses: make(map[string]*ServiceHealthState),
		stopCh:   make(chan struct{}),
	}
}

// Start starts the background health checking loop.
func (m *Manager) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-m.stopCh:
				return
			case <-ticker.C:
				m.checkAll(ctx)
			}
		}
	}()
}

// Stop stops the background health checking loop.
func (m *Manager) Stop() {
	close(m.stopCh)
}

// RegisterService adds a service to be health checked.
func (m *Manager) RegisterService(serviceID string, config *configv1.UpstreamServiceConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	checker := NewChecker(config)
	if checker == nil {
		return
	}
	m.checkers[serviceID] = checker
	m.statuses[serviceID] = &ServiceHealthState{
		Status: pb_admin.ServiceStatus_SERVICE_STATUS_UNKNOWN,
	}

	// Run an initial check immediately in background
	go func() {
		m.checkService(context.Background(), serviceID, checker)
	}()
}

// UnregisterService removes a service from health checking.
func (m *Manager) UnregisterService(serviceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.checkers, serviceID)
	delete(m.statuses, serviceID)
}

// GetStatus returns the current health state of a service.
func (m *Manager) GetStatus(serviceID string) *ServiceHealthState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if state, ok := m.statuses[serviceID]; ok {
		// Return a copy
		return &ServiceHealthState{
			Status:        state.Status,
			LastError:     state.LastError,
			LastCheckTime: state.LastCheckTime,
		}
	}
	return nil
}

func (m *Manager) checkAll(ctx context.Context) {
	m.mu.RLock()
	// Copy checkers to avoid holding lock during check
	checkers := make(map[string]health.Checker, len(m.checkers))
	for k, v := range m.checkers {
		checkers[k] = v
	}
	m.mu.RUnlock()

	var wg sync.WaitGroup
	for id, checker := range checkers {
		wg.Add(1)
		go func(id string, checker health.Checker) {
			defer wg.Done()
			m.checkService(ctx, id, checker)
		}(id, checker)
	}
	wg.Wait()
}

func (m *Manager) checkService(ctx context.Context, id string, checker health.Checker) {
	// 5 second timeout for health check
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := checker.Check(ctx)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Verify service is still registered
	state, ok := m.statuses[id]
	if !ok {
		return
	}

	state.LastCheckTime = time.Now()
	if result.Status == health.StatusUp {
		state.Status = pb_admin.ServiceStatus_SERVICE_STATUS_HEALTHY
		state.LastError = ""
	} else {
		state.Status = pb_admin.ServiceStatus_SERVICE_STATUS_UNHEALTHY
		// Collect errors
		var errs []string
		for _, v := range result.Details {
			if v.Error != nil {
				errs = append(errs, v.Error.Error())
			}
		}
		if len(errs) > 0 {
			state.LastError = strings.Join(errs, "; ")
		} else {
			state.LastError = "Health check failed"
		}
	}
}
