/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useMemo } from "react";
import { cn } from "@/lib/utils";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { HealthHistoryPoint, ServiceStatus } from "@/hooks/use-service-health-history";

interface StatusBlock {
    startTime: number;
    endTime: number;
    status: ServiceStatus | 'empty';
    count: number;
}

const BLOCKS_COUNT = 96; // 24 hours * 4 quarters = 96 blocks (15 min each)

export function StatusTimeline({ history }: { history: HealthHistoryPoint[] }) {
    const blocks = useMemo(() => {
        if (!history || history.length === 0) return Array(BLOCKS_COUNT).fill({ status: 'empty' });

        const now = Date.now();
        const twentyFourHoursAgo = now - 24 * 60 * 60 * 1000;
        const interval = (24 * 60 * 60 * 1000) / BLOCKS_COUNT; // 15 mins

        const result: StatusBlock[] = [];

        for (let i = 0; i < BLOCKS_COUNT; i++) {
            const blockStart = twentyFourHoursAgo + (i * interval);
            const blockEnd = blockStart + interval;

            // Find points in this window
            const pointsInWindow = history.filter(h => h.timestamp >= blockStart && h.timestamp < blockEnd);

            if (pointsInWindow.length === 0) {
                result.push({ startTime: blockStart, endTime: blockEnd, status: 'empty', count: 0 });
                continue;
            }

            // Determine worst status
            // Priority: unhealthy > degraded > healthy
            let worstStatus: ServiceStatus = 'healthy';
            if (pointsInWindow.some(p => p.status === 'unhealthy')) {
                worstStatus = 'unhealthy';
            } else if (pointsInWindow.some(p => p.status === 'degraded')) {
                worstStatus = 'degraded';
            } else if (pointsInWindow.every(p => p.status === 'inactive')) {
                worstStatus = 'inactive';
            }

            result.push({
                startTime: blockStart,
                endTime: blockEnd,
                status: worstStatus,
                count: pointsInWindow.length
            });
        }
        return result;
    }, [history]);

    return (
        <div className="flex items-center gap-[1px] h-8 w-full">
            {blocks.map((block, i) => {
                let colorClass = "bg-muted/30"; // Empty/No Data
                if (block.status === 'healthy') colorClass = "bg-green-500 hover:bg-green-400";
                if (block.status === 'degraded') colorClass = "bg-amber-500 hover:bg-amber-400";
                if (block.status === 'unhealthy') colorClass = "bg-red-500 hover:bg-red-400";
                if (block.status === 'inactive') colorClass = "bg-slate-300 dark:bg-slate-700";

                return (
                    <Tooltip key={i} delayDuration={0}>
                        <TooltipTrigger asChild>
                            <div
                                className={cn("flex-1 h-full first:rounded-l-sm last:rounded-r-sm transition-all hover:opacity-80 cursor-crosshair", colorClass)}
                            />
                        </TooltipTrigger>
                        <TooltipContent className="text-xs p-2">
                            <div className="font-semibold capitalize flex items-center gap-2">
                                <div className={cn("h-2 w-2 rounded-full", colorClass)} />
                                {block.status === 'empty' ? 'No Data' : block.status}
                            </div>
                            <div className="text-muted-foreground mt-1">
                                {new Date(block.startTime).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})} - {new Date(block.endTime).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})}
                            </div>
                            {block.count > 0 && (
                                <div className="mt-1 text-[10px] opacity-70">
                                    {block.count} checks recorded
                                </div>
                            )}
                        </TooltipContent>
                    </Tooltip>
                );
            })}
        </div>
    );
}
