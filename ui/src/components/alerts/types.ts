/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Summary: Defines the severity levels for alerts.
 *
 * Parameters:
 *   - None
 *
 * Returns:
 *   - None
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 */
export type Severity = "critical" | "warning" | "info";

/**
 * Summary: Defines the lifecycle status states for an alert.
 *
 * Parameters:
 *   - None
 *
 * Returns:
 *   - None
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 */
export type AlertStatus = "active" | "acknowledged" | "resolved";

/**
 * Summary: Represents a system alert with its metadata and state.
 *
 * Parameters:
 *   - None
 *
 * Returns:
 *   - None
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
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
 * Summary: Defines a rule for triggering alerts based on metric thresholds.
 *
 * Parameters:
 *   - None
 *
 * Returns:
 *   - None
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
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
