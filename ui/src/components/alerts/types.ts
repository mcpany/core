/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Severity levels for alerts.
 */
export type Severity = "critical" | "warning" | "info";
/**
 * AlertStatus type definition.
 */
export type AlertStatus = "active" | "acknowledged" | "resolved";

/**
 * Alert type definition.
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
 * Alert type definition.
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
