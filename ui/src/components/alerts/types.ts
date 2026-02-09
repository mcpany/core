/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Defines the severity level of an alert.
 */
export type Severity = "critical" | "warning" | "info";

/**
 * Defines the current status of an alert in its lifecycle.
 */
export type AlertStatus = "active" | "acknowledged" | "resolved";

/**
 * Represents a single alert instance triggered by the system.
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
 * Defines a rule for triggering alerts based on metrics.
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
