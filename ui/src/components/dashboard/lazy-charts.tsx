/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import dynamic from "next/dynamic";

const ChartSkeleton = () => (
  <div className="col-span-3 rounded-lg border bg-card text-card-foreground shadow-sm h-full backdrop-blur-sm bg-background/50">
      <div className="p-6 flex flex-col space-y-1.5">
          <div className="h-6 w-1/3 bg-muted animate-pulse rounded" />
          <div className="h-4 w-2/3 bg-muted animate-pulse rounded" />
      </div>
      <div className="p-6 pt-0 h-[300px] flex items-center justify-center">
          <div className="h-full w-full bg-muted/20 animate-pulse rounded" />
      </div>
  </div>
);

// ⚡ Bolt Optimization: Lazy load heavy chart components to reduce initial bundle size
// and improve Time to Interactive. 'recharts' is a large dependency.
export const LazyRequestVolumeChart = dynamic(
  () => import("@/components/dashboard/request-volume-chart").then((mod) => mod.RequestVolumeChart),
  {
    ssr: false,
    loading: () => <ChartSkeleton />,
  }
);

export const LazyTopToolsWidget = dynamic(
  () => import("@/components/dashboard/top-tools-widget").then((mod) => mod.TopToolsWidget),
  {
    ssr: false,
    loading: () => <ChartSkeleton />,
  }
);

export const LazyHealthHistoryChart = dynamic(
  () => import("@/components/stats/health-history-chart").then((mod) => mod.HealthHistoryChart),
  {
    ssr: false,
    loading: () => <ChartSkeleton />,
  }
);
