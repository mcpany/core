/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { MetricsOverview } from "@/components/dashboard/metrics-overview";
import { ServiceHealthWidget } from "@/components/dashboard/service-health-widget";
import { LazyRequestVolumeChart, LazyTopToolsWidget, LazyHealthHistoryChart, LazyRecentActivityWidget } from "@/components/dashboard/lazy-charts";
import { ToolFailureRateWidget } from "@/components/dashboard/tool-failure-rate-widget";
import { Activity, BarChart, Server, AlertTriangle, TrendingUp, Hash, HeartPulse } from "lucide-react";

/**
 * Defines the possible sizes for a widget in the dashboard grid.
 */
export type WidgetSize = "full" | "half" | "third" | "two-thirds";

/**
 * Defines the configuration and metadata for a dashboard widget.
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
    /** The React component that renders the widget. */
    component: React.ComponentType<any>;
    /** The icon to display for the widget in the registry. */
    icon: React.ElementType;
}

/**
 * A registry of all available dashboard widgets.
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
        type: "failure-rate",
        title: "Tool Failure Rates",
        description: "Top failing tools with error counts.",
        defaultSize: "third",
        component: ToolFailureRateWidget,
        icon: AlertTriangle
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
    {
        type: "service-health",
        title: "Service Health",
        description: "Status and health checks for connected services.",
        defaultSize: "third",
        component: ServiceHealthWidget,
        icon: HeartPulse
    },
];

/**
 * Retrieves a widget definition by its type.
 * @param type - The unique type identifier of the widget.
 * @returns The widget definition if found, otherwise undefined.
 */
export const getWidgetDefinition = (type: string): WidgetDefinition | undefined => {
    return WIDGET_DEFINITIONS.find(w => w.type === type);
};
