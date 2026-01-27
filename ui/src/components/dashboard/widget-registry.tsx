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

export type WidgetSize = "full" | "half" | "third" | "two-thirds";

export interface WidgetDefinition {
    type: string;
    title: string;
    description: string;
    defaultSize: WidgetSize;
    component: React.ComponentType<any>;
    icon: React.ElementType;
}

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

export const getWidgetDefinition = (type: string): WidgetDefinition | undefined => {
    return WIDGET_DEFINITIONS.find(w => w.type === type);
};
