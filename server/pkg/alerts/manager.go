// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package alerts manages system alerts and incidents.
package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mcpany/core/server/pkg/logging"
)

// ManagerInterface defines the interface for managing alerts.
//
// Summary: defines the interface for managing alerts.
type ManagerInterface interface {
	// ListAlerts returns a list of all alerts.
	//
	// Summary: returns a list of all alerts.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   - []*Alert: The []*Alert.
	ListAlerts() []*Alert
	// GetAlert retrieves an alert by its ID.
	//
	// Summary: retrieves an alert by its ID.
	//
	// Parameters:
	//   - id: string. The unique identifier.
	//
	// Returns:
	//   - *Alert: The *Alert.
	GetAlert(id string) *Alert
	// CreateAlert creates a new alert.
	//
	// Summary: creates a new alert.
	//
	// Parameters:
	//   - alert: *Alert. The alert.
	//
	// Returns:
	//   - *Alert: The *Alert.
	CreateAlert(alert *Alert) *Alert
	// UpdateAlert updates an existing alert.
	//
	// Summary: updates an existing alert.
	//
	// Parameters:
	//   - id: string. The unique identifier.
	//   - alert: *Alert. The alert.
	//
	// Returns:
	//   - *Alert: The *Alert.
	UpdateAlert(id string, alert *Alert) *Alert

	// Webhooks

	// GetWebhookURL returns the configured global webhook URL.
	//
	// Summary: returns the configured global webhook URL.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   - string: The string.
	GetWebhookURL() string
	// SetWebhookURL sets the configured global webhook URL.
	//
	// Summary: sets the configured global webhook URL.
	//
	// Parameters:
	//   - url: string. The URL.
	//
	// Returns:
	//   None.
	SetWebhookURL(url string)

	// Rules

	// ListRules returns a list of all alert rules.
	//
	// Summary: returns a list of all alert rules.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   - []*AlertRule: The []*AlertRule.
	ListRules() []*AlertRule
	// GetRule retrieves an alert rule by its ID.
	//
	// Summary: retrieves an alert rule by its ID.
	//
	// Parameters:
	//   - id: string. The unique identifier.
	//
	// Returns:
	//   - *AlertRule: The *AlertRule.
	GetRule(id string) *AlertRule
	// CreateRule creates a new alert rule.
	//
	// Summary: creates a new alert rule.
	//
	// Parameters:
	//   - rule: *AlertRule. The alert rule.
	//
	// Returns:
	//   - *AlertRule: The *AlertRule.
	CreateRule(rule *AlertRule) *AlertRule
	// UpdateRule updates an existing alert rule.
	//
	// Summary: updates an existing alert rule.
	//
	// Parameters:
	//   - id: string. The unique identifier.
	//   - rule: *AlertRule. The alert rule.
	//
	// Returns:
	//   - *AlertRule: The *AlertRule.
	UpdateRule(id string, rule *AlertRule) *AlertRule
	// DeleteRule deletes an alert rule by its ID.
	//
	// Summary: deletes an alert rule by its ID.
	//
	// Parameters:
	//   - id: string. The unique identifier.
	//
	// Returns:
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	DeleteRule(id string) error
}

// Manager implements ManagerInterface using in-memory storage.
//
// Summary: implements ManagerInterface using in-memory storage.
type Manager struct {
	mu         sync.RWMutex
	alerts     map[string]*Alert
	rules      map[string]*AlertRule
	webhookURL string
}

// NewManager creates a new Manager and seeds it with initial data.
//
// Summary: creates a new Manager and seeds it with initial data.
//
// Parameters:
//   None.
//
// Returns:
//   - *Manager: The *Manager.
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

// ListAlerts returns all alerts sorted by timestamp descending.
//
// Summary: returns all alerts sorted by timestamp descending.
//
// Parameters:
//   None.
//
// Returns:
//   - []*Alert: The []*Alert.
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
//
// Summary: returns an alert by ID, or nil if not found.
//
// Parameters:
//   - id: string. The id.
//
// Returns:
//   - *Alert: The *Alert.
func (m *Manager) GetAlert(id string) *Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.alerts[id]
}

// CreateAlert creates a new alert.
//
// Summary: creates a new alert.
//
// Parameters:
//   - alert: *Alert. The alert.
//
// Returns:
//   - *Alert: The *Alert.
func (m *Manager) CreateAlert(alert *Alert) *Alert {
	m.mu.Lock()
	if alert.ID == "" {
		alert.ID = "AL-" + uuid.New().String()[:8]
	}
	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now()
	}
	m.alerts[alert.ID] = alert
	webhookURL := m.webhookURL
	m.mu.Unlock()

	// Trigger webhook asynchronously
	if webhookURL != "" {
		go func(url string, a *Alert) {
			body, err := json.Marshal(a)
			if err != nil {
				logging.GetLogger().Error("failed to marshal alert for webhook", "error", err)
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
			if err != nil {
				logging.GetLogger().Error("failed to create webhook request", "error", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				logging.GetLogger().Error("failed to call webhook", "url", url, "error", err)
				return
			}
			defer func() { _ = resp.Body.Close() }()
		}(webhookURL, alert)
	}

	return alert
}

// UpdateAlert updates an existing alert.
//
// Summary: updates an existing alert.
//
// Parameters:
//   - id: string. The id.
//   - alert: *Alert. The alert.
//
// Returns:
//   - *Alert: The *Alert.
func (m *Manager) UpdateAlert(id string, alert *Alert) *Alert {
	m.mu.Lock()
	existing, ok := m.alerts[id]
	if !ok {
		m.mu.Unlock()
		return nil
	}
	// Update fields
	if alert.Status != "" {
		existing.Status = alert.Status
	}
	// Can add more updatable fields here
	webhookURL := m.webhookURL
	m.mu.Unlock()

	// Trigger webhook asynchronously if status changed (or just on update?)
	// For now trigger on any update as "Incident Response" implies updates are important.
	if webhookURL != "" {
		go func(url string, a *Alert) {
			body, err := json.Marshal(a)
			if err != nil {
				logging.GetLogger().Error("failed to marshal alert for webhook", "error", err)
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
			if err != nil {
				logging.GetLogger().Error("failed to create webhook request", "error", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				logging.GetLogger().Error("failed to call webhook", "url", url, "error", err)
				return
			}
			defer func() { _ = resp.Body.Close() }()
		}(webhookURL, existing)
	}

	return existing
}

// GetWebhookURL returns the configured global webhook URL.
//
// Summary: returns the configured global webhook URL.
//
// Parameters:
//   None.
//
// Returns:
//   - string: The string.
func (m *Manager) GetWebhookURL() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.webhookURL
}

// SetWebhookURL sets the configured global webhook URL.
//
// Summary: sets the configured global webhook URL.
//
// Parameters:
//   - url: string. The url.
//
// Returns:
//   None.
func (m *Manager) SetWebhookURL(url string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.webhookURL = url
}

// ListRules returns all rules.
//
// Summary: returns all rules.
//
// Parameters:
//   None.
//
// Returns:
//   - []*AlertRule: The []*AlertRule.
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

// GetRule returns a rule by ID.
//
// Summary: returns a rule by ID.
//
// Parameters:
//   - id: string. The id.
//
// Returns:
//   - *AlertRule: The *AlertRule.
func (m *Manager) GetRule(id string) *AlertRule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rules[id]
}

// CreateRule creates a new rule.
//
// Summary: creates a new rule.
//
// Parameters:
//   - rule: *AlertRule. The rule.
//
// Returns:
//   - *AlertRule: The *AlertRule.
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

// UpdateRule updates a rule.
//
// Summary: updates a rule.
//
// Parameters:
//   - id: string. The id.
//   - rule: *AlertRule. The rule.
//
// Returns:
//   - *AlertRule: The *AlertRule.
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

// DeleteRule deletes a rule.
//
// Summary: deletes a rule.
//
// Parameters:
//   - id: string. The id.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (m *Manager) DeleteRule(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rules, id)
	return nil
}
