// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package alerts manages system alerts and incidents.
package alerts

import (
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ManagerInterface defines the interface for managing alerts.
type ManagerInterface interface {
	ListAlerts() []*Alert
	GetAlert(id string) *Alert
	CreateAlert(alert *Alert) *Alert
	UpdateAlert(id string, alert *Alert) *Alert
}

// Manager implements ManagerInterface using in-memory storage.
type Manager struct {
	mu     sync.RWMutex
	alerts map[string]*Alert
}

// NewManager creates a new Manager and seeds it with initial data.
func NewManager() *Manager {
	m := &Manager{
		alerts: make(map[string]*Alert),
	}
	m.seedData()
	return m
}

func (m *Manager) seedData() {
	now := time.Now()
	// Mock data from frontend
	m.CreateAlert(&Alert{ID: "AL-1024", Title: "High CPU Usage", Message: "CPU usage > 90% for 5m", Severity: SeverityCritical, Status: StatusActive, Service: "weather-service", Source: "System Monitor", Timestamp: now.Add(-5 * time.Minute)})
	m.CreateAlert(&Alert{ID: "AL-1023", Title: "API Latency Spike", Message: "P99 Latency > 2000ms", Severity: SeverityWarning, Status: StatusActive, Service: "api-gateway", Source: "Latency Watchdog", Timestamp: now.Add(-15 * time.Minute)})
	m.CreateAlert(&Alert{ID: "AL-1022", Title: "Disk Space Low", Message: "Volume /data at 85%", Severity: SeverityWarning, Status: StatusAcknowledged, Service: "database-primary", Source: "Disk Monitor", Timestamp: now.Add(-45 * time.Minute)})
	m.CreateAlert(&Alert{ID: "AL-1021", Title: "Connection Refused", Message: "Upstream connection failed", Severity: SeverityCritical, Status: StatusResolved, Service: "payment-provider", Source: "Connectivity Check", Timestamp: now.Add(-2 * time.Hour)})
	m.CreateAlert(&Alert{ID: "AL-1020", Title: "New Service Deployed", Message: "Service 'search-v2' detected", Severity: SeverityInfo, Status: StatusResolved, Service: "discovery", Source: "Orchestrator", Timestamp: now.Add(-5 * time.Hour)})
}

// ListAlerts returns all alerts sorted by timestamp descending.
func (m *Manager) ListAlerts() []*Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*Alert, 0, len(m.alerts))
	for _, a := range m.alerts {
		list = append(list, a)
	}
	// Sort by timestamp desc
	sort.Slice(list, func(i, j int) bool {
		return list[i].Timestamp.After(list[j].Timestamp)
	})
	return list
}

// GetAlert returns an alert by ID, or nil if not found.
func (m *Manager) GetAlert(id string) *Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.alerts[id]
}

// CreateAlert creates a new alert.
func (m *Manager) CreateAlert(alert *Alert) *Alert {
	m.mu.Lock()
	defer m.mu.Unlock()
	if alert.ID == "" {
		alert.ID = "AL-" + uuid.New().String()[:8]
	}
	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now()
	}
	m.alerts[alert.ID] = alert
	return alert
}

// UpdateAlert updates an existing alert.
func (m *Manager) UpdateAlert(id string, alert *Alert) *Alert {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.alerts[id]
	if !ok {
		return nil
	}
	// Update fields
	if alert.Status != "" {
		existing.Status = alert.Status
	}
	// Can add more updatable fields here
	return existing
}
