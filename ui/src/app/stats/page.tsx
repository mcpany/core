/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { AnalyticsDashboard } from "@/components/stats/analytics-dashboard";

/**
 * Summary: StatsPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
export default function StatsPage() {
  return (
    <div className="flex flex-col h-full">
      <AnalyticsDashboard />
    </div>
  );
}
