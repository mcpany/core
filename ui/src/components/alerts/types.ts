/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

export type Severity = "critical" | "warning" | "info";
/**
 * AlertStatus type definition.
 */
export type AlertStatus = "active" | "acknowledged" | "resolved";

/**
 * Represents a system alert.
 */
export interface Alert {
  /** Unique identifier for the alert. */
  id: string;
  /** Title of the alert. */
  title: string;
  /** Detailed message of the alert. */
  message: string;
  /** Severity level of the alert. */
  severity: Severity;
  /** Current status of the alert. */
  status: AlertStatus;
  /** Service associated with the alert. */
  service: string;
  /** Timestamp when the alert was triggered (ISO string). */
  timestamp: string;
  /** Source of the alert (e.g., system, user). */
  source: string;
}

/**
 * Represents a rule for triggering alerts.
 */
export interface AlertRule {
  /** Unique identifier for the rule. */
  id: string;
  /** Name of the rule. */
  name: string;
  /** Metric to monitor. */
  metric: string;
  /** Operator for comparison (e.g., >, <, =). */
  operator: string;
  /** Threshold value for the metric. */
  threshold: number;
  /** Duration for which the condition must be true. */
  duration: string;
  /** Severity level to assign to triggered alerts. */
  severity: Severity;
  /** Whether the rule is enabled. */
  enabled: boolean;
  /** Timestamp of the last update to the rule. */
  last_updated?: string;
}
