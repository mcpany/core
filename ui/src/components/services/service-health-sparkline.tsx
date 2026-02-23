/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { memo, useMemo } from "react";
import { useServiceHealth } from "@/contexts/service-health-context";
import { Sparkline } from "@/components/charts/sparkline";

interface ServiceHealthSparklineProps {
    serviceName: string;
    disabled?: boolean;
}

// ⚡ Bolt Optimization: Isolated health visualization to prevent row re-renders on context updates.
// Randomized Selection from Top 5 High-Impact Targets
/**
 * ServiceHealthSparkline component.
 * @param props - The component props.
 * @param props.serviceName - The name of the service to display health for.
 * @param props.disabled - Whether the service is disabled.
 * @returns The rendered component.
 */
export const ServiceHealthSparkline = memo(function ServiceHealthSparkline({ serviceName, disabled }: ServiceHealthSparklineProps) {
    const { getServiceHistory } = useServiceHealth();
    const history = getServiceHistory(serviceName);

    const latencies = useMemo(() => history.map(h => h.latencyMs), [history]);
    const maxLatency = useMemo(() => Math.max(...latencies, 50), [latencies]); // Minimum max of 50ms for scale

    // Determine color based on latest health
    const healthColor = useMemo(() => {
        if (!history.length) return "#94a3b8"; // slate-400
        const latest = history[history.length - 1];
        if (latest.status === 'NODE_STATUS_ERROR' || latest.errorRate > 0.1) return "#ef4444"; // red-500
        if (latest.latencyMs > 500) return "#eab308"; // yellow-500
        return "#22c55e"; // green-500
    }, [history]);

    if (disabled) return null;

    return (
        <Sparkline
            data={latencies}
            width={80}
            height={24}
            color={healthColor}
            max={maxLatency}
        />
    );
});
