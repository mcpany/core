/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useDashboard } from "@/components/dashboard/dashboard-context";
import { Clock } from "lucide-react";

/**
 * A dropdown component to filter dashboard views by a specific time range.
 * Updates the global dashboard context when a range is selected.
 *
 * @returns The rendered time range picker component.
 */
export function TimeRangePicker() {
  const { timeRange, setTimeRange } = useDashboard();

  return (
    <div className="flex items-center space-x-2">
      <Clock className="h-4 w-4 text-muted-foreground" />
      <Select
        value={timeRange}
        onValueChange={setTimeRange}
      >
        <SelectTrigger className="w-[120px] h-8">
          <SelectValue placeholder="Time Range" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="1h">Last 1 Hour</SelectItem>
          <SelectItem value="6h">Last 6 Hours</SelectItem>
          <SelectItem value="12h">Last 12 Hours</SelectItem>
          <SelectItem value="24h">Last 24 Hours</SelectItem>
        </SelectContent>
      </Select>
    </div>
  );
}
