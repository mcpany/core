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
	// ListAlerts returns a list of all alerts.
	ListAlerts() []*Alert
	// GetAlert retrieves an alert by its ID.
	GetAlert(id string) *Alert
	// CreateAlert creates a new alert.
	CreateAlert(alert *Alert) *Alert
	// UpdateAlert updates an existing alert.
	UpdateAlert(id string, alert *Alert) *Alert

	// Rules

	// ListRules returns a list of all alert rules.
	ListRules() []*AlertRule
	// GetRule retrieves an alert rule by its ID.
	GetRule(id string) *AlertRule
	// CreateRule creates a new alert rule.
	CreateRule(rule *AlertRule) *AlertRule
	// UpdateRule updates an existing alert rule.
	UpdateRule(id string, rule *AlertRule) *AlertRule
	// DeleteRule deletes an alert rule by its ID.
	DeleteRule(id string) error
}

// Manager implements ManagerInterface using in-memory storage.
type Manager struct {
	mu     sync.RWMutex
	alerts map[string]*Alert
	rules  map[string]*AlertRule
}

// NewManager creates a new Manager and seeds it with initial data.
func NewManager() *Manager {
	m := &Manager{
		alerts: make(map[string]*Alert),
		rules:  make(map[string]*AlertRule),
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

	// Seed Rules
	m.CreateRule(&AlertRule{ID: "rule-1", Name: "High CPU", Metric: "cpu_usage", Operator: ">", Threshold: 90, Duration: "5m", Severity: SeverityCritical, Enabled: true, LastUpdated: now})
	m.CreateRule(&AlertRule{ID: "rule-2", Name: "High Latency", Metric: "http_latency_p99", Operator: ">", Threshold: 1000, Duration: "1m", Severity: SeverityWarning, Enabled: true, LastUpdated: now})
}

// ListAlerts retrieves all active and resolved alerts from the manager.
// It returns a slice of Alert pointers, sorted by timestamp in descending order.
//
// Returns:
//   - []*Alert: A slice of all alerts.
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

// GetAlert retrieves a specific alert by its unique identifier.
//
// Parameters:
//   - id: string. The unique identifier of the alert.
//
// Returns:
//   - *Alert: The alert if found, or nil if not.
func (m *Manager) GetAlert(id string) *Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.alerts[id]
}

// CreateAlert adds a new alert to the system.
// If the alert ID is missing, a new unique ID is generated.
// If the timestamp is zero, it defaults to the current time.
//
// Parameters:
//   - alert: *Alert. The alert object to create.
//
// Returns:
//   - *Alert: The created alert with populated fields (ID, Timestamp).
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

// UpdateAlert updates the details of an existing alert.
// Only fields that are allowed to be updated (e.g., Status) are modified.
//
// Parameters:
//   - id: string. The ID of the alert to update.
//   - alert: *Alert. The object containing the updated fields.
//
// Returns:
//   - *Alert: The updated alert, or nil if the alert was not found.
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

// ListRules retrieves all configured alert rules.
// It returns a slice of AlertRule pointers, sorted by name.
//
// Returns:
//   - []*AlertRule: A slice of all alert rules.
func (m *Manager) ListRules() []*AlertRule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*AlertRule, 0, len(m.rules))
	for _, r := range m.rules {
		list = append(list, r)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})
	return list
}

// GetRule retrieves a specific alert rule by its unique identifier.
//
// Parameters:
//   - id: string. The unique identifier of the rule.
//
// Returns:
//   - *AlertRule: The alert rule if found, or nil if not.
func (m *Manager) GetRule(id string) *AlertRule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rules[id]
}

// CreateRule adds a new alert rule to the system.
// A unique ID is generated if one is not provided, and the last updated timestamp is set.
//
// Parameters:
//   - rule: *AlertRule. The rule object to create.
//
// Returns:
//   - *AlertRule: The created alert rule.
func (m *Manager) CreateRule(rule *AlertRule) *AlertRule {
	m.mu.Lock()
	defer m.mu.Unlock()
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	rule.LastUpdated = time.Now()
	m.rules[rule.ID] = rule
	return rule
}

// UpdateRule updates the configuration of an existing alert rule.
// It updates the definition and the last updated timestamp.
//
// Parameters:
//   - id: string. The ID of the rule to update.
//   - rule: *AlertRule. The new rule configuration.
//
// Returns:
//   - *AlertRule: The updated alert rule, or nil if the rule was not found.
func (m *Manager) UpdateRule(id string, rule *AlertRule) *AlertRule {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.rules[id]
	if !ok {
		return nil
	}
	existing.Name = rule.Name
	existing.Metric = rule.Metric
	existing.Operator = rule.Operator
	existing.Threshold = rule.Threshold
	existing.Duration = rule.Duration
	existing.Severity = rule.Severity
	existing.Enabled = rule.Enabled
	existing.LastUpdated = time.Now()
	return existing
}

// DeleteRule removes an alert rule from the system by its ID.
//
// Parameters:
//   - id: string. The unique identifier of the rule to delete.
//
// Returns:
//   - error: nil if successful (even if the rule didn't exist).
func (m *Manager) DeleteRule(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rules, id)
	return nil
}
