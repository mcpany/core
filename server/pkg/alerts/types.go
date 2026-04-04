// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package alerts

import "time"

// Severity represents the severity level of an alert.
//
// Summary: Represents the severity level of an alert.
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

type Severity string

// Status represents the status of an alert.
//
// Summary: Represents the status of an alert.
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

type Status string

const (
	// SeverityCritical indicates a critical issue.
//
// Summary: Indicates a critical issue.
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

	SeverityCritical Severity = "critical"
	// SeverityWarning indicates a warning.
//
// Summary: Indicates a warning.
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

	SeverityWarning Severity = "warning"
	// SeverityInfo indicates an informational alert.
//
// Summary: Indicates an informational alert.
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

	SeverityInfo Severity = "info"

	// StatusActive indicates the alert is currently active.
//
// Summary: Indicates the alert is currently active.
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

	StatusActive Status = "active"
	// StatusAcknowledged indicates the alert has been acknowledged.
//
// Summary: Indicates the alert has been acknowledged.
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

	StatusAcknowledged Status = "acknowledged"
	// StatusResolved indicates the alert has been resolved.
//
// Summary: Indicates the alert has been resolved.
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

	StatusResolved Status = "resolved"
)

// Alert represents a system alert.
//
// Summary: Represents a system alert.
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
// Summary: Defines a condition for triggering an alert.
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
// Summary: Represents aggregated statistics for alerts.
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

type AlertStats struct {
	ActiveCritical      int    `json:"activeCritical"`
	ActiveCriticalTrend string `json:"activeCriticalTrend"`
	ActiveWarning       int    `json:"activeWarning"`
	ActiveWarningTrend  string `json:"activeWarningTrend"`
	MTTR                string `json:"mttr"`
	MTTRTrend           string `json:"mttrTrend"`
	TotalToday          int    `json:"totalToday"`
	TotalTodayTrend     string `json:"totalTodayTrend"`
}
