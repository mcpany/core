/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { AnalyticsDashboard } from "@/components/stats/analytics-dashboard";

/**
 * Intent: Document StatsPage
 *
 * Params:
 *   - None
 *
 * Returns:
 *   - Documented below.
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * StatsPage component.
 * @returns The rendered component.
 */
export default function StatsPage() {
  return (
    <div className="flex flex-col h-full">
      <AnalyticsDashboard />
    </div>
  );
}
