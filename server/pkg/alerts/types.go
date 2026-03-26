// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package alerts

import "time"

// Severity represents the severity level of an alert.
//
// Summary: Represents a Severity.
type Severity string

// Status represents the status of an alert.
//
// Summary: Represents a Status.
type Status string

const (
	// SeverityCritical indicates a critical issue.
	// Summary: Defines SeverityCritica.
	SeverityCritical Severity = "critical"
	// SeverityWarning indicates a warning.
	// Summary: Defines SeverityWarnin.
	SeverityWarning Severity = "warning"
	// SeverityInfo indicates an informational alert.
	// Summary: Defines SeverityInf.
	SeverityInfo Severity = "info"

	// StatusActive indicates the alert is currently active.
	// Summary: Defines StatusActiv.
	StatusActive Status = "active"
	// StatusAcknowledged indicates the alert has been acknowledged.
	// Summary: Defines StatusAcknowledge.
	StatusAcknowledged Status = "acknowledged"
	// StatusResolved indicates the alert has been resolved.
	// Summary: Defines StatusResolve.
	StatusResolved Status = "resolved"
)

// Alert alert represents a alert.
//
// Summary: Alert represents a alert.
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

// AlertRule alertRule represents a alert rule.
//
// Summary: AlertRule represents a alert rule.
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

// AlertStats alertStats represents a alert stats.
//
// Summary: AlertStats represents a alert stats.
type AlertStats struct {
	ActiveCritical int    `json:"activeCritical"`
	ActiveWarning  int    `json:"activeWarning"`
	MTTR           string `json:"mttr"`
	TotalToday     int    `json:"totalToday"`
}
