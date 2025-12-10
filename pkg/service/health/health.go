// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Status represents the health status of a service.
type Status int

const (
	// StatusUnknown means the health status has not been determined yet.
	StatusUnknown Status = iota
	// StatusHealthy means the service is healthy and reachable.
	StatusHealthy
	// StatusUnhealthy means the service is unhealthy and should not be used.
	StatusUnhealthy
)

// String returns the string representation of the Status.
func (s Status) String() string {
	switch s {
	case StatusHealthy:
		return "Healthy"
	case StatusUnhealthy:
		return "Unhealthy"
	default:
		return "Unknown"
	}
}

// Checkable is an interface for services that can be health-checked.
type Checkable interface {
	// ID returns the unique identifier of the service.
	ID() string
	// HealthCheck performs the health check and returns an error if the service is unhealthy.
	HealthCheck(ctx context.Context) error
	// Interval returns the duration between health checks.
	Interval() time.Duration
}

// Checker is the interface for the health checking manager.
type Checker interface {
	// Start begins the health checking process for the given service.
	Start(service Checkable)
	// Stop ceases health checks for the given service ID.
	Stop(serviceID string)
	// Status returns the current health status of the service.
	Status(serviceID string) Status
	// Shutdown stops all health checks.
	Shutdown()
}

// Manager is the concrete implementation of the Checker.
type Manager struct {
	statuses  sync.Map // map[string]Status
	stopChans sync.Map // map[string]chan struct{}
}

// NewManager creates a new health check Manager.
func NewManager() *Manager {
	return &Manager{}
}

// Start begins the health checking process for a given service.
func (m *Manager) Start(service Checkable) {
	if _, ok := m.stopChans.Load(service.ID()); ok {
		log.Debug().Str("service_id", service.ID()).Msg("Health check already started for service")
		return
	}

	stopChan := make(chan struct{})
	m.stopChans.Store(service.ID(), stopChan)
	m.statuses.Store(service.ID(), StatusUnknown)

	go m.runCheck(service, stopChan)
}

// Stop ceases health checks for a given service ID.
func (m *Manager) Stop(serviceID string) {
	if ch, ok := m.stopChans.Load(serviceID); ok {
		close(ch.(chan struct{}))
		m.stopChans.Delete(serviceID)
		m.statuses.Delete(serviceID)
		log.Info().Str("service_id", serviceID).Msg("Stopped health check for service")
	}
}

// Status returns the current health status of a service.
func (m *Manager) Status(serviceID string) Status {
	if status, ok := m.statuses.Load(serviceID); ok {
		return status.(Status)
	}
	return StatusUnknown
}

// Shutdown stops all health checks.
func (m *Manager) Shutdown() {
	m.stopChans.Range(func(key, value interface{}) bool {
		serviceID := key.(string)
		stopChan := value.(chan struct{})
		close(stopChan)
		m.stopChans.Delete(serviceID)
		log.Info().Str("service_id", serviceID).Msg("Shutdown health check for service")
		return true
	})
}

// runCheck is the main loop for a single service's health check.
func (m *Manager) runCheck(service Checkable, stopChan <-chan struct{}) {
	interval := service.Interval()
	if interval <= 0 {
		log.Warn().Str("service_id", service.ID()).Msg("Health check interval is zero or negative, defaulting to 15 seconds")
		interval = 15 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial check
	m.performCheck(service)

	for {
		select {
		case <-ticker.C:
			m.performCheck(service)
		case <-stopChan:
			return
		}
	}
}

func (m *Manager) performCheck(service Checkable) {
	// Use the interval as the timeout for the check itself.
	ctx, cancel := context.WithTimeout(context.Background(), service.Interval())
	defer cancel()

	previousStatus := m.Status(service.ID())

	if err := service.HealthCheck(ctx); err != nil {
		if previousStatus != StatusUnhealthy {
			log.Warn().Err(err).Str("service_id", service.ID()).Msg("Service health check failed")
			m.statuses.Store(service.ID(), StatusUnhealthy)
		}
	} else {
		if previousStatus != StatusHealthy {
			log.Info().Str("service_id", service.ID()).Msg("Service is now healthy")
			m.statuses.Store(service.ID(), StatusHealthy)
		}
	}
}
