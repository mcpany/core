/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Defines the severity levels for alerts.
 * - `critical`: Requires immediate attention.
 * - `warning`: Potential issue that should be monitored.
 * - `info`: Informational message.
 */
export type Severity = "critical" | "warning" | "info";

/**
 * Defines the lifecycle status of an alert.
 * - `active`: The alert is currently triggering.
 * - `acknowledged`: The alert has been seen by an operator but not yet fixed.
 * - `resolved`: The issue causing the alert has been fixed.
 */
export type AlertStatus = "active" | "acknowledged" | "resolved";

/**
 * Represents a single alert instance triggered by the system.
 */
export interface Alert {
  /** Unique identifier for the alert. */
  id: string;
  /** Brief summary of the alert. */
  title: string;
  /** Detailed description of the issue. */
  message: string;
  /** Severity level of the alert. */
  severity: Severity;
  /** Current status of the alert. */
  status: AlertStatus;
  /** Name of the service that triggered the alert. */
  service: string;
  /** Timestamp when the alert was triggered (ISO 8601 string). */
  timestamp: string; // ISO string
  /** Source component or system that generated the alert. */
  source: string;
}

/**
 * Defines a rule for triggering alerts based on specific conditions.
 */
export interface AlertRule {
  /** Unique identifier for the alert rule. */
  id: string;
  /** Human-readable name of the rule. */
  name: string;
  /** The condition expression that triggers the alert (e.g., query or threshold). */
  condition: string;
  /** The severity level assigned to alerts triggered by this rule. */
  severity: Severity;
  /** The service to which this rule applies. */
  service: string;
  /** Whether the rule is currently active. */
  enabled: boolean;
}
