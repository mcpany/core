/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Severity levels for alerts.
 */
export type Severity = "critical" | "warning" | "info";

/**
 * AlertStatus represents the current state of an alert.
 */
export type AlertStatus = "active" | "acknowledged" | "resolved";

/**
 * Alert represents a system alert event.
 *
 * @property id - Unique identifier for the alert.
 * @property title - Short summary of the alert.
 * @property message - Detailed description of the alert condition.
 * @property severity - The severity level of the alert.
 * @property status - The current status of the alert.
 * @property service - The name of the service that generated the alert.
 * @property timestamp - The time when the alert was triggered (ISO string).
 * @property source - The source of the alert (e.g., "system", "user").
 */
export interface Alert {
  id: string;
  title: string;
  message: string;
  severity: Severity;
  status: AlertStatus;
  service: string;
  timestamp: string; // ISO string
  source: string;
}

/**
 * AlertRule defines a rule for triggering alerts based on metrics.
 *
 * @property id - Unique identifier for the rule.
 * @property name - Human-readable name of the rule.
 * @property metric - The metric to monitor (e.g., "cpu_usage").
 * @property operator - The comparison operator (e.g., ">", "<").
 * @property threshold - The value to compare against.
 * @property duration - The duration the condition must persist before triggering.
 * @property severity - The severity level assigned to alerts triggered by this rule.
 * @property enabled - Whether the rule is currently active.
 * @property last_updated - Timestamp of the last update to the rule.
 */
export interface AlertRule {
  id: string;
  name: string;
  metric: string;
  operator: string;
  threshold: number;
  duration: string;
  severity: Severity;
  enabled: boolean;
  last_updated?: string;
}
