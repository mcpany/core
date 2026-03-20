// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
// Summary: Severity represents the severity level of an alert.
//
// Side Effects:
//   - None.
//
// Summary: Status represents the status of an alert.
//
// Side Effects:
//   - None.
//
// Summary: SeverityCritical indicates a critical issue.
//
// Side Effects:
//   - None.
//
// Summary: SeverityWarning indicates a warning.
//
// Side Effects:
//   - None.
//
// Summary: SeverityInfo indicates an informational alert.
//
// Side Effects:
//   - None.
//
// Summary: StatusActive indicates the alert is currently active.
//
// Side Effects:
//   - None.
//
// Summary: StatusAcknowledged indicates the alert has been acknowledged.
//
// Side Effects:
//   - None.
//
// Summary: StatusResolved indicates the alert has been resolved.
//
// Side Effects:
//   - None.
//
// Summary: Alert represents a system alert.
//
// Side Effects:
//   - None.
//
// Summary: AlertRule defines a condition for triggering an alert.
//
// Side Effects:
//   - None.
package alerts

import "time"

type Severity string

type Status string

const (
	SeverityCritical Severity = "critical"

	SeverityWarning Severity = "warning"

	SeverityInfo Severity = "info"

	StatusActive Status = "active"

	StatusAcknowledged Status = "acknowledged"

	StatusResolved Status = "resolved"
)

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

type AlertRule struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Metric    string  `json:"metric"`
	Operator  string  `json:"operator"` // e.g. ">", "<", "="
	Threshold float64 `json:"threshold"`
	Duration  string  `json:"duration"` // e.g. "5m"
	// Summary: AlertStats represents aggregated statistics for alerts.
	//
	// Side Effects:
	//   - None.
	Severity    Severity  `json:"severity"`
	Enabled     bool      `json:"enabled"`
	LastUpdated time.Time `json:"last_updated"`
}

type AlertStats struct {
	ActiveCritical int    `json:"activeCritical"`
	ActiveWarning  int    `json:"activeWarning"`
	MTTR           string `json:"mttr"`
	TotalToday     int    `json:"totalToday"`
}
