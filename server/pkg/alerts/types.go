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
	//
	// Summary: Designates an alert as critical, requiring immediate attention.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - Severity: the critical severity value
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	SeverityCritical Severity = "critical"

	// SeverityWarning indicates a warning.
	//
	// Summary: Designates an alert as a warning, requiring eventual review.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - Severity: the warning severity value
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	SeverityWarning Severity = "warning"

	// SeverityInfo indicates an informational alert.
	//
	// Summary: Designates an alert as informational, requiring no action.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - Severity: the info severity value
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	SeverityInfo Severity = "info"

	// StatusActive indicates the alert is currently active.
	//
	// Summary: Marks an alert as currently firing and unresolved.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - Status: the active status value
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	StatusActive Status = "active"

	// StatusAcknowledged indicates the alert has been acknowledged.
	//
	// Summary: Marks an alert as acknowledged by a user but not yet resolved.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - Status: the acknowledged status value
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	StatusAcknowledged Status = "acknowledged"

	// StatusResolved indicates the alert has been resolved.
	//
	// Summary: Marks an alert as fully resolved and no longer active.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - Status: the resolved status value
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	StatusResolved Status = "resolved"
)

// Alert represents a system alert.
//
// Summary: Represents a Alert.
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
// Summary: Represents a AlertRule.
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
// Summary: Represents a AlertStats.
type AlertStats struct {
	ActiveCritical int    `json:"activeCritical"`
	ActiveWarning  int    `json:"activeWarning"`
	MTTR           string `json:"mttr"`
	TotalToday     int    `json:"totalToday"`
}
