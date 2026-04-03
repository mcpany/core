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
 * Severity represents the public Severity entity.
 *
 * Summary: Defines the structured data model representing a .
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - None.
 *
 * Throws/Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
 */
export type Severity = "critical" | "warning" | "info";
/**
 * AlertStatus represents the public AlertStatus entity.
 *
 * Summary: Defines the structured data model representing a status.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - None.
 *
 * Throws/Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
 */
export type AlertStatus = "active" | "acknowledged" | "resolved";

/**
 * Alert represents the public Alert entity.
 *
 * Summary: Defines the structured data model representing a .
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - None.
 *
 * Throws/Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
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
 * Summary: Document AlertRule
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
 * Alert type definition.
 */

/**
 * AlertRule represents the public AlertRule entity.
 *
 * Summary: Defines the structured data model representing a rule.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - None.
 *
 * Throws/Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
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
