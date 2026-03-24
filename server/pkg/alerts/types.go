// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package alerts

import "time"

// Severity represents the severity level of an alert.
//
// Summary: Type representing the severity level of an alert (critical, warning, info).
type Severity string

// Status represents the status of an alert.
//
// Summary: Type representing the lifecycle status of an alert (active, acknowledged, resolved).
type Status string

const (
	// SeverityCritical indicates a critical issue.
	// Summary: Severity level for critical alerts requiring immediate attention.
	SeverityCritical Severity = "critical"
	// SeverityWarning indicates a warning.
	// Summary: Severity level for warning alerts.
	SeverityWarning Severity = "warning"
	// SeverityInfo indicates an informational alert.
	// Summary: Severity level for informational alerts.
	SeverityInfo Severity = "info"

	// StatusActive indicates the alert is currently active.
	// Summary: Status indicating the alert is currently triggered and active.
	StatusActive Status = "active"
	// StatusAcknowledged indicates the alert has been acknowledged.
	// Summary: Status indicating the alert has been acknowledged by a user.
	StatusAcknowledged Status = "acknowledged"
	// StatusResolved indicates the alert has been resolved.
	// Summary: Status indicating the alert has been resolved.
	StatusResolved Status = "resolved"
)

// Alert represents a system alert.
//
// Summary: Data structure representing a system-level alert or incident.
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

// AlertRule defines a condition for triggering an alert.
//
// Summary: Configuration for a metric-based alert triggering rule.
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

// AlertStats represents aggregated statistics for alerts.
//
// Summary: Aggregated metrics and statistics for system alerts.
type AlertStats struct {
	ActiveCritical int    `json:"activeCritical"`
	ActiveWarning  int    `json:"activeWarning"`
	MTTR           string `json:"mttr"`
	TotalToday     int    `json:"totalToday"`
}
