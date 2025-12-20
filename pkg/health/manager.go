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
	"google.golang.org/protobuf/proto"
)

// Manager handles health checking for registered services.
type Manager struct {
	mu       sync.RWMutex
	checkers map[string]health.Checker
	statuses map[string]*pb_admin.ServiceState
	stopCh   chan struct{}
}

// NewManager creates a new health Manager.
func NewManager() *Manager {
	return &Manager{
		checkers: make(map[string]health.Checker),
		statuses: make(map[string]*pb_admin.ServiceState),
		stopCh:   make(chan struct{}),
	}
}

// RegisterService adds a service to the health manager.
func (m *Manager) RegisterService(id string, config *configv1.UpstreamServiceConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	statusUnknown := pb_admin.ServiceStatus_SERVICE_STATUS_UNKNOWN

	checker := NewChecker(config)
	if checker == nil {
		// If no checker can be created (e.g. unknown type), just store Unknown state
		m.statuses[id] = &pb_admin.ServiceState{
			Config: config,
			Status: &statusUnknown,
		}
		return
	}

	m.checkers[id] = checker
	// Initial state
	now := time.Now().Unix()
	m.statuses[id] = &pb_admin.ServiceState{
		Config:        config,
		Status:        &statusUnknown,
		LastCheckTime: proto.Int64(now),
	}

	// Trigger an initial check in background
	go m.checkService(id, checker)
}

// UnregisterService removes a service from the health manager.
func (m *Manager) UnregisterService(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.checkers, id)
	delete(m.statuses, id)
}

// GetState returns the current state of a service.
func (m *Manager) GetState(id string) (*pb_admin.ServiceState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	state, ok := m.statuses[id]
	return state, ok
}

// Start starts the periodic health checks.
func (m *Manager) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-m.stopCh:
				return
			case <-ticker.C:
				m.runChecks(ctx)
			}
		}
	}()
}

// Stop stops the background health checks.
func (m *Manager) Stop() {
	close(m.stopCh)
}

func (m *Manager) runChecks(ctx context.Context) {
	// Snapshot checkers to avoid holding lock during checks
	m.mu.RLock()
	checkers := make(map[string]health.Checker)
	for id, c := range m.checkers {
		checkers[id] = c
	}
	m.mu.RUnlock()

	for id, checker := range checkers {
		if ctx.Err() != nil {
			return
		}
		// Check concurrently? For now, sequential is safer for resources.
		m.checkService(id, checker)
	}
}

func (m *Manager) checkService(id string, checker health.Checker) {
	// Create a short timeout for the check
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result := checker.Check(ctx)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure service wasn't removed while we were checking
	state, ok := m.statuses[id]
	if !ok {
		return
	}

	state.LastCheckTime = proto.Int64(time.Now().Unix())
	if result.Status == health.StatusUp {
		status := pb_admin.ServiceStatus_SERVICE_STATUS_HEALTHY
		state.Status = &status
		state.LastError = nil
	} else {
		status := pb_admin.ServiceStatus_SERVICE_STATUS_UNHEALTHY
		state.Status = &status
		// Aggregate errors from Details
		var errs []string
		for _, detail := range result.Details {
			if detail.Error != nil {
				errs = append(errs, detail.Error.Error())
			}
		}
		if len(errs) > 0 {
			state.LastError = proto.String(strings.Join(errs, "; "))
		} else {
			state.LastError = proto.String("Unknown error")
		}
	}
}
