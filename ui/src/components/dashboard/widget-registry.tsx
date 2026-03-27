/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { MetricsOverview } from "@/components/dashboard/metrics-overview";
import { ServiceHealthWidget } from "@/components/dashboard/service-health-widget";
import { LazyRequestVolumeChart, LazyTopToolsWidget, LazyHealthHistoryChart, LazyRecentActivityWidget, LazyAuditLogWidget } from "@/components/dashboard/lazy-charts";
import { ToolFailureRateWidget } from "@/components/dashboard/tool-failure-rate-widget";
import { QuickActionsWidget } from "@/components/dashboard/quick-actions-widget";
import { NetworkGraphWidget } from "@/components/dashboard/network-graph-widget";
import { ActiveIntentAlignmentWidget } from "@/components/dashboard/active-intent-alignment-widget";
import { Activity, BarChart, Server, AlertTriangle, TrendingUp, Hash, HeartPulse, Zap, Share2, ClipboardCheck } from "lucide-react";

/**
 * Intent: Document WidgetSize
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
 * Defines the possible sizes for a dashboard widget.
 * - full: Takes up the full width (12 columns).
 * - two-thirds: Takes up 2/3 of the width (8 columns).
 * - half: Takes up 1/2 of the width (6 columns).
 * - third: Takes up 1/3 of the width (4 columns).
 */
export type WidgetSize = "full" | "half" | "third" | "two-thirds";

/**
 * Intent: Document WidgetDefinition
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
 * Intent: Document WIDGET_DEFINITIONS
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
 * Registry of all available dashboard widgets.
 * This list determines what widgets are available to add to the dashboard.
 */
export const WIDGET_DEFINITIONS: WidgetDefinition[] = [
    {
        type: "intent-alignment",
        title: "Active Intent Alignment Monitor",
        description: "Visual indicator for AIA heartbeat status and semantic drift alerts.",
        defaultSize: "third",
        component: ActiveIntentAlignmentWidget,
        icon: Activity
    },
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
    {
        type: "audit-logs",
        title: "Audit Logs",
        description: "Track and inspect tool executions for compliance.",
        defaultSize: "full",
        component: LazyAuditLogWidget,
        icon: ClipboardCheck
    },
    {
        type: "swarm-topology",
        title: "Swarm Topology",
        description: "Multi-Agent Swarm Topology & Sovereignty Monitor.",
        defaultSize: "two-thirds",
        component: React.lazy(() => import('./swarm-topology-widget').then(m => ({ default: m.SwarmTopologyWidget }))),
        icon: Share2
    },
];

/**
 * Intent: Document getWidgetDefinition
 *
 * Params:
 *   - Documented below.
 *
 * Returns:
 *   - Documented below.
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * Retrieves a widget definition by its type.
 *
 * @param type - The widget type identifier.
 * @returns The widget definition if found, otherwise undefined.
 */
export const getWidgetDefinition = (type: string): WidgetDefinition | undefined => {
    return WIDGET_DEFINITIONS.find(w => w.type === type);
};
