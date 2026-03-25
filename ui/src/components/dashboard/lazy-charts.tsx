/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { lazy, Suspense } from "react";

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

const _LazyRequestVolumeChart = lazy(() =>
  import("@/components/dashboard/request-volume-chart").then((mod) => ({
    default: mod.RequestVolumeChart,
  })),
);
const _LazyRecentActivityWidget = lazy(() =>
  import("@/components/dashboard/recent-activity-widget").then((mod) => ({
    default: mod.RecentActivityWidget,
  })),
);
const _LazyTopToolsWidget = lazy(() =>
  import("@/components/dashboard/top-tools-widget").then((mod) => ({
    default: mod.TopToolsWidget,
  })),
);
const _LazyHealthHistoryChart = lazy(() =>
  import("@/components/stats/health-history-chart").then((mod) => ({
    default: mod.HealthHistoryChart,
  })),
);
const _LazyAuditLogWidget = lazy(() =>
  import("@/components/audit/audit-log-viewer").then((mod) => ({
    default: mod.AuditLogViewer,
  })),
);

/** LazyRequestVolumeChart with Suspense skeleton. */
export const LazyRequestVolumeChart = (props: object) => (
  <Suspense fallback={<ChartSkeleton />}>
    <_LazyRequestVolumeChart {...(props as any)} />
  </Suspense>
);

/** LazyRecentActivityWidget with Suspense skeleton. */
export const LazyRecentActivityWidget = (props: object) => (
  <Suspense fallback={<ChartSkeleton />}>
    <_LazyRecentActivityWidget {...(props as any)} />
  </Suspense>
);

/** LazyTopToolsWidget with Suspense skeleton. */
export const LazyTopToolsWidget = (props: object) => (
  <Suspense fallback={<ChartSkeleton />}>
    <_LazyTopToolsWidget {...(props as any)} />
  </Suspense>
);

/** LazyHealthHistoryChart with Suspense skeleton. */
export const LazyHealthHistoryChart = (props: object) => (
  <Suspense fallback={<ChartSkeleton />}>
    <_LazyHealthHistoryChart {...(props as any)} />
  </Suspense>
);

/** LazyAuditLogWidget with Suspense skeleton. */
export const LazyAuditLogWidget = (props: object) => (
  <Suspense fallback={<ChartSkeleton />}>
    <_LazyAuditLogWidget {...(props as any)} />
  </Suspense>
);
