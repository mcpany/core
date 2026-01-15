/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

export interface CheckResult {
  status: string;
  message?: string;
  latency?: string;
}

export interface DoctorReport {
  status: string;
  timestamp: string;
  checks: Record<string, CheckResult>;
}
