/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { MetricsOverview } from "@/components/dashboard/metrics-overview";
import { ServiceHealthWidget } from "@/components/dashboard/service-health-widget";
import { LazyRequestVolumeChart, LazyTopToolsWidget, LazyHealthHistoryChart, LazyRecentActivityWidget } from "@/components/dashboard/lazy-charts";
import { ToolFailureRateWidget } from "@/components/dashboard/tool-failure-rate-widget";
import { Activity, BarChart, Server, AlertTriangle, LineChart, Hash } from "lucide-react";

export type WidgetType = "full" | "half" | "third" | "two-thirds";

export interface WidgetDefinition {
    id: string; // The unique type identifier (e.g., 'metrics', 'recent-activity')
    title: string;
    description: string;
    defaultSize: WidgetType;
    component: React.ComponentType<any>;
    icon: React.ElementType;
}

export const WIDGET_DEFINITIONS: Record<string, WidgetDefinition> = {
    "metrics": {
        id: "metrics",
        title: "Metrics Overview",
        description: "Key performance indicators including total requests, success rate, and active services.",
        defaultSize: "full",
        component: MetricsOverview,
        icon: BarChart
    },
    "recent-activity": {
        id: "recent-activity",
        title: "Recent Activity",
        description: "A live feed of the most recent tool executions and errors.",
        defaultSize: "half",
        component: LazyRecentActivityWidget,
        icon: Activity
    },
    "uptime": {
        id: "uptime",
        title: "System Uptime",
        description: "Historical uptime tracking for connected services.",
        defaultSize: "half",
        component: LazyHealthHistoryChart,
        icon: LineChart
    },
    "failure-rate": {
        id: "failure-rate",
        title: "Tool Failure Rates",
        description: "Top tools ranked by their failure rate.",
        defaultSize: "third",
        component: ToolFailureRateWidget,
        icon: AlertTriangle
    },
    "request-volume": {
        id: "request-volume",
        title: "Request Volume",
        description: "Trend of request volume over time.",
        defaultSize: "half",
        component: LazyRequestVolumeChart,
        icon: BarChart
    },
    "top-tools": {
        id: "top-tools",
        title: "Top Tools",
        description: "Most frequently used tools.",
        defaultSize: "third",
        component: LazyTopToolsWidget,
        icon: Hash
    },
    "service-health": {
        id: "service-health",
        title: "Service Health",
        description: "Real-time health status of all connected services.",
        defaultSize: "third",
        component: ServiceHealthWidget,
        icon: Server
    }
};

export const AVAILABLE_WIDGETS = Object.values(WIDGET_DEFINITIONS);
