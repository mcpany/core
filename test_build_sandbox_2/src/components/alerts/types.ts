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
  resolvedAt?: string; // ISO string, optional
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
  lastUpdated?: string;
}
