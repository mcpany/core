/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Severity defines the criticality level of an alert.
 */
export type Severity = "critical" | "warning" | "info";

/**
 * AlertStatus represents the current state of an alert.
 */
export type AlertStatus = "active" | "acknowledged" | "resolved";

/**
 * Alert represents a triggered notification event from a service.
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
