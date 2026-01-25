/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { MetricsOverview } from "@/components/dashboard/metrics-overview";
import { ServiceHealthWidget } from "@/components/dashboard/service-health-widget";
import { LazyRequestVolumeChart, LazyTopToolsWidget, LazyHealthHistoryChart, LazyRecentActivityWidget } from "@/components/dashboard/lazy-charts";
import { ToolFailureRateWidget } from "@/components/dashboard/tool-failure-rate-widget";
import { QuickActionsWidget } from "@/components/dashboard/quick-actions-widget";
import { Activity, BarChart, Server, AlertTriangle, TrendingUp, Hash, HeartPulse, Zap } from "lucide-react";

/**
 * Defines the available sizes for a dashboard widget.
 * - full: Spans the entire width (12 columns).
 * - two-thirds: Spans two-thirds of the width (8 columns).
 * - half: Spans half the width (6 columns).
 * - third: Spans one-third of the width (4 columns).
 */
export type WidgetSize = "full" | "half" | "third" | "two-thirds";

/**
 * Defines the structure and metadata for a dashboard widget type.
 */
export interface WidgetDefinition {
    /** The unique type identifier for the widget. */
    type: string;
    /** The display title of the widget. */
    title: string;
    /** A brief description of what the widget displays. */
    description: string;
    /** The default size of the widget when added to the dashboard. */
    defaultSize: WidgetSize;
    /** The React component responsible for rendering the widget. */
    component: React.ComponentType<any>;
    /** The icon associated with the widget. */
    icon: React.ElementType;
}

/**
 * A registry of all available dashboard widgets.
 * This list is used to populate the "Add Widget" menu and to look up widget definitions.
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
 * Retrieves the widget definition for a given widget type.
 *
 * @param type - The unique type identifier of the widget to find.
 * @returns The definition of the widget if found, otherwise undefined.
 */
export const getWidgetDefinition = (type: string): WidgetDefinition | undefined => {
    return WIDGET_DEFINITIONS.find(w => w.type === type);
};
