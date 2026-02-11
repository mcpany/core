/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import dynamic from "next/dynamic";
import { Activity, BarChart, Server, AlertTriangle, TrendingUp, Hash, HeartPulse, Zap, Share2 } from "lucide-react";
import { LazyRequestVolumeChart, LazyTopToolsWidget, LazyHealthHistoryChart, LazyRecentActivityWidget } from "@/components/dashboard/lazy-charts";
import { WidgetSkeleton } from "@/components/dashboard/widget-skeleton";

// ⚡ Bolt Optimization: Lazy load heavy widget components to reduce initial bundle size.
// Randomized Selection from Top 5 High-Impact Targets.
const MetricsOverview = dynamic(() => import("@/components/dashboard/metrics-overview").then(mod => mod.MetricsOverview), { loading: () => <WidgetSkeleton /> });
const ServiceHealthWidget = dynamic(() => import("@/components/dashboard/service-health-widget").then(mod => mod.ServiceHealthWidget), { loading: () => <WidgetSkeleton /> });
const ToolFailureRateWidget = dynamic(() => import("@/components/dashboard/tool-failure-rate-widget").then(mod => mod.ToolFailureRateWidget), { loading: () => <WidgetSkeleton /> });
const QuickActionsWidget = dynamic(() => import("@/components/dashboard/quick-actions-widget").then(mod => mod.QuickActionsWidget), { loading: () => <WidgetSkeleton /> });
const NetworkGraphWidget = dynamic(() => import("@/components/dashboard/network-graph-widget").then(mod => mod.NetworkGraphWidget), { loading: () => <WidgetSkeleton /> });

/**
 * Defines the possible sizes for a dashboard widget.
 * - full: Takes up the full width (12 columns).
 * - two-thirds: Takes up 2/3 of the width (8 columns).
 * - half: Takes up 1/2 of the width (6 columns).
 * - third: Takes up 1/3 of the width (4 columns).
 */
export type WidgetSize = "full" | "half" | "third" | "two-thirds";

/**
 * Defines the metadata and component for a dashboard widget.
 */
export interface WidgetDefinition {
    /** Unique identifier for the widget type. */
    type: string;
    /** Display title of the widget. */
    title: string;
    /** Brief description of what the widget does. */
    description: string;
    /** The default size when the widget is first added. */
    defaultSize: WidgetSize;
    /** The React component to render. */
    component: React.ComponentType<any>;
    /** Icon to display in the widget picker. */
    icon: React.ElementType;
}

/**
 * Registry of all available dashboard widgets.
 * This list determines what widgets are available to add to the dashboard.
 */
export const WIDGET_DEFINITIONS: WidgetDefinition[] = [
    {
        type: "metrics",
        title: "Metrics Overview",
        description: "Key performance indicators including RPS, Latency, and Error Rate.",
        defaultSize: "full",
        component: MetricsOverview,
        icon: Activity
    },
    {
        type: "network-topology",
        title: "Network Topology",
        description: "Visual graph of connected services and tools.",
        defaultSize: "full",
        component: NetworkGraphWidget,
        icon: Share2
    },
    {
        type: "quick-actions",
        title: "Quick Actions",
        description: "Shortcuts to common management tasks.",
        defaultSize: "third",
        component: QuickActionsWidget,
        icon: Zap
    },
    {
        type: "service-health",
        title: "Service Health",
        description: "Status and health checks for connected services.",
        defaultSize: "third",
        component: ServiceHealthWidget,
        icon: HeartPulse
    },
    {
        type: "failure-rate",
        title: "Tool Failure Rates",
        description: "Top failing tools with error counts.",
        defaultSize: "third",
        component: ToolFailureRateWidget,
        icon: AlertTriangle
    },
    {
        type: "recent-activity",
        title: "Recent Activity",
        description: "Real-time log of tool executions and their status.",
        defaultSize: "half",
        component: LazyRecentActivityWidget,
        icon: TrendingUp
    },
    {
        type: "uptime",
        title: "System Uptime",
        description: "Historical uptime and availability chart.",
        defaultSize: "half",
        component: LazyHealthHistoryChart,
        icon: Server
    },
    {
        type: "request-volume",
        title: "Request Volume",
        description: "Request volume trends over time.",
        defaultSize: "half",
        component: LazyRequestVolumeChart,
        icon: BarChart
    },
    {
        type: "top-tools",
        title: "Top Tools",
        description: "Most frequently used tools.",
        defaultSize: "third",
        component: LazyTopToolsWidget,
        icon: Hash
    },
];

/**
 * Retrieves a widget definition by its type.
 *
 * @param type - The widget type identifier.
 * @returns The widget definition if found, otherwise undefined.
 */
export const getWidgetDefinition = (type: string): WidgetDefinition | undefined => {
    return WIDGET_DEFINITIONS.find(w => w.type === type);
};
