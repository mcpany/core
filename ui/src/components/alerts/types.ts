/**
 * Summary: Document Severity
 *
 * Params:
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
 *
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Summary: Defines the severity levels for alerts.
 * Parameters: None.
 * Returns: None.
 * Throws/Errors: None.
 */
export type Severity = "critical" | "warning" | "info";
/**
 * Summary: Defines the status of an alert.
 * Parameters: None.
 * Returns: None.
 * Throws/Errors: None.
 */
export type AlertStatus = "active" | "acknowledged" | "resolved";

/**
 * Summary: Represents an alert.
 * Parameters: None.
 * Returns: None.
 * Throws/Errors: None.
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
 * Summary: Represents a rule for triggering alerts.
 * Parameters: None.
 * Returns: None.
 * Throws/Errors: None.
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
