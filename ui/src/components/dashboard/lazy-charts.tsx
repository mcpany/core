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
  import("@/components/dashboard/request-volume-chart").then((mod) => ({ default: mod.RequestVolumeChart }))
);
const _LazyRecentActivityWidget = lazy(() =>
  import("@/components/dashboard/recent-activity-widget").then((mod) => ({ default: mod.RecentActivityWidget }))
);
const _LazyTopToolsWidget = lazy(() =>
  import("@/components/dashboard/top-tools-widget").then((mod) => ({ default: mod.TopToolsWidget }))
);
const _LazyHealthHistoryChart = lazy(() =>
  import("@/components/stats/health-history-chart").then((mod) => ({ default: mod.HealthHistoryChart }))
);
const _LazyAuditLogWidget = lazy(() =>
  import("@/components/audit/audit-log-viewer").then((mod) => ({ default: mod.AuditLogViewer }))
);

/**
 * LazyRequestVolumeChart executes the LazyRequestVolumeChart logic.
 *
 * Summary: Executes the LazyRequestVolumeChart logic.
 *
 * @param props - The props parameter.
 * @returns The result of the operation.
 * @throws An error if the operation fails.
 */
export const LazyRequestVolumeChart = (props: object) => (
  <Suspense fallback={<ChartSkeleton />}><_LazyRequestVolumeChart {...(props as any)} /></Suspense>
);

/**
 * Intent: Document LazyRecentActivityWidget
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
 * LazyRecentActivityWidget with Suspense skeleton.
/**
 * LazyRecentActivityWidget executes the LazyRecentActivityWidget logic.
 *
 * Summary: Executes the LazyRecentActivityWidget logic.
 *
 * @param props - The props parameter.
 * @returns The result of the operation.
 * @throws An error if the operation fails.
 */
export const LazyRecentActivityWidget = (props: object) => (
  <Suspense fallback={<ChartSkeleton />}><_LazyRecentActivityWidget {...(props as any)} /></Suspense>
);

/**
 * Intent: Document LazyTopToolsWidget
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
 * LazyTopToolsWidget with Suspense skeleton.
/**
 * LazyTopToolsWidget executes the LazyTopToolsWidget logic.
 *
 * Summary: Executes the LazyTopToolsWidget logic.
 *
 * @param props - The props parameter.
 * @returns The result of the operation.
 * @throws An error if the operation fails.
 */
export const LazyTopToolsWidget = (props: object) => (
  <Suspense fallback={<ChartSkeleton />}><_LazyTopToolsWidget {...(props as any)} /></Suspense>
);

/**
 * Intent: Document LazyHealthHistoryChart
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
 * LazyHealthHistoryChart with Suspense skeleton.
/**
 * LazyHealthHistoryChart executes the LazyHealthHistoryChart logic.
 *
 * Summary: Executes the LazyHealthHistoryChart logic.
 *
 * @param props - The props parameter.
 * @returns The result of the operation.
 * @throws An error if the operation fails.
 */
export const LazyHealthHistoryChart = (props: object) => (
  <Suspense fallback={<ChartSkeleton />}><_LazyHealthHistoryChart {...(props as any)} /></Suspense>
);

/**
 * Intent: Document LazyAuditLogWidget
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
 * LazyAuditLogWidget with Suspense skeleton.
/**
 * LazyAuditLogWidget executes the LazyAuditLogWidget logic.
 *
 * Summary: Executes the LazyAuditLogWidget logic.
 *
 * @param props - The props parameter.
 * @returns The result of the operation.
 * @throws An error if the operation fails.
 */
export const LazyAuditLogWidget = (props: object) => (
  <Suspense fallback={<ChartSkeleton />}><_LazyAuditLogWidget {...(props as any)} /></Suspense>
);
