// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package alerts

import "time"

// Severity represents the public Severity entity.
//
// Summary: Defines the structured data model representing a .
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

// Status represents the public Status entity.
//
// Summary: Defines the structured data model representing a .
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

// Alert represents the public Alert entity.
//
// Summary: Defines the structured data model representing a .
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

// AlertRule represents the public AlertRule entity.
//
// Summary: Defines the structured data model representing a rule.
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

// AlertStats represents the public AlertStats entity.
//
// Summary: Defines the structured data model representing a stats.
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
