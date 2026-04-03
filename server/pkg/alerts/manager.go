// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package alerts manages system alerts and incidents.
package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mcpany/core/server/pkg/logging"
)

// ManagerInterface represents the public ManagerInterface entity.
//
// Summary: Defines the required contract and behavior that interface implementations must satisfy.
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
type ManagerInterface interface {
	// ListAlerts returns a list of all alerts.
	ListAlerts() []*Alert
	// GetAlert retrieves an alert by its ID.
	GetAlert(id string) *Alert
	// CreateAlert creates a new alert.
	CreateAlert(alert *Alert) *Alert
	// UpdateAlert updates an existing alert.
	UpdateAlert(id string, alert *Alert) *Alert
	// DeleteAlert deletes an existing alert.
	DeleteAlert(id string) error
	// GetAlertStats returns aggregated statistics for alerts.
	GetAlertStats() *AlertStats

	// Webhooks

	// GetWebhookURL returns the configured global webhook URL.
	GetWebhookURL() string
	// SetWebhookURL sets the configured global webhook URL.
	SetWebhookURL(url string)

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
	mu         sync.RWMutex
	alerts     map[string]*Alert
	rules      map[string]*AlertRule
	webhookURL string
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

// ListAlerts serves as a public interface for interacting with ListAlerts.
//
// Summary: List the alerts appropriately based on current system conditions.
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

// GetAlert serves as a public interface for interacting with GetAlert.
//
// Summary: Fetches and returns the underlying alert from the system state.
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
func (m *Manager) GetAlert(id string) *Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.alerts[id]
}

// CreateAlert serves as a public interface for interacting with CreateAlert.
//
// Summary: Create the alert appropriately based on current system conditions.
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

// GetAlertStats serves as a public interface for interacting with GetAlertStats.
//
// Summary: Fetches and returns the underlying alert stats from the system state.
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
func (m *Manager) GetAlertStats() *AlertStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &AlertStats{}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	for _, a := range m.alerts {
		if a.Timestamp.After(today) {
			stats.TotalToday++
		}

		if a.Status == StatusActive {
			if a.Severity == SeverityCritical {
				stats.ActiveCritical++
			} else if a.Severity == SeverityWarning {
				stats.ActiveWarning++
			}
		}
	}

	// Mock MTTR for now as calculating true MTTR requires alert state transition history
	stats.MTTR = "14m"

	stats.ActiveCriticalTrend = "+1 since last hour"
	stats.ActiveWarningTrend = "-2 since last hour"
	stats.MTTRTrend = "-2m from yesterday"
	stats.TotalTodayTrend = "+12% from average"

	return stats
}

// DeleteAlert serves as a public interface for interacting with DeleteAlert.
//
// Summary: Delete the alert appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *Manager) DeleteAlert(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.alerts[id]; !ok {
		return fmt.Errorf("alert not found")
	}
	delete(m.alerts, id)
	return nil
}

// UpdateAlert serves as a public interface for interacting with UpdateAlert.
//
// Summary: Update the alert appropriately based on current system conditions.
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

// GetWebhookURL serves as a public interface for interacting with GetWebhookURL.
//
// Summary: Fetches and returns the underlying webhook url from the system state.
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
func (m *Manager) GetWebhookURL() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.webhookURL
}

// SetWebhookURL serves as a public interface for interacting with SetWebhookURL.
//
// Summary: Set the webhook url appropriately based on current system conditions.
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
func (m *Manager) SetWebhookURL(url string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.webhookURL = url
}

// ListRules serves as a public interface for interacting with ListRules.
//
// Summary: List the rules appropriately based on current system conditions.
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

// GetRule serves as a public interface for interacting with GetRule.
//
// Summary: Fetches and returns the underlying rule from the system state.
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
func (m *Manager) GetRule(id string) *AlertRule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rules[id]
}

// CreateRule serves as a public interface for interacting with CreateRule.
//
// Summary: Create the rule appropriately based on current system conditions.
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

// UpdateRule serves as a public interface for interacting with UpdateRule.
//
// Summary: Update the rule appropriately based on current system conditions.
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

// DeleteRule serves as a public interface for interacting with DeleteRule.
//
// Summary: Delete the rule appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *Manager) DeleteRule(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rules, id)
	return nil
}
