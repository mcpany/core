// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package alerts

import "time"

// Severity represents the severity level of an alert.
type Severity string

// Status represents the status of an alert.
type Status string

const (
	// SeverityCritical indicates a critical issue.
	SeverityCritical Severity = "critical"
	// SeverityWarning indicates a warning.
	SeverityWarning Severity = "warning"
	// SeverityInfo indicates an informational alert.
	SeverityInfo Severity = "info"

	// StatusActive indicates the alert is currently active.
	StatusActive Status = "active"
	// StatusAcknowledged indicates the alert has been acknowledged.
	StatusAcknowledged Status = "acknowledged"
	// StatusResolved indicates the alert has been resolved.
	StatusResolved Status = "resolved"
)

// Alert - Auto-generated documentation.
//
// Summary: Alert represents a system alert.
//
// Fields:
//   - Various fields for Alert.
type Alert struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Severity  Severity  `json:"severity"`
	Status    Status    `json:"status"`
	Service   string    `json:"service"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
}

// AlertRule - Auto-generated documentation.
//
// Summary: AlertRule defines a condition for triggering an alert.
//
// Fields:
//   - Various fields for AlertRule.
type AlertRule struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Metric      string    `json:"metric"`
	Operator    string    `json:"operator"` // e.g. ">", "<", "="
	Threshold   float64   `json:"threshold"`
	Duration    string    `json:"duration"` // e.g. "5m"
	Severity    Severity  `json:"severity"`
	Enabled     bool      `json:"enabled"`
	LastUpdated time.Time `json:"last_updated"`
}

// AlertStats - Auto-generated documentation.
//
// Summary: AlertStats represents aggregated statistics for alerts.
//
// Fields:
//   - Various fields for AlertStats.
type AlertStats struct {
	ActiveCritical int    `json:"activeCritical"`
	ActiveWarning  int    `json:"activeWarning"`
	MTTR           string `json:"mttr"`
	TotalToday     int    `json:"totalToday"`
}
