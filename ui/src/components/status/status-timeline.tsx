/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useMemo } from "react";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import { HealthHistoryPoint, ServiceStatus } from "@/hooks/use-service-health-history";

interface StatusTimelineProps {
    history: HealthHistoryPoint[];
    totalDurationMs?: number; // Defaults to 24h
    bucketCount?: number; // Defaults to 96 (15 min intervals for 24h)
}

type BucketStatus = ServiceStatus | "no-data";

interface Bucket {
    startTime: number;
    endTime: number;
    status: BucketStatus;
    count: number;
    errorCount: number;
}

const statusPriority = (status: BucketStatus): number => {
    switch (status) {
        case "unhealthy": return 4;
        case "degraded": return 3;
        case "healthy": return 2;
        case "inactive": return 1;
        default: return 0;
    }
};

const getStatusColor = (status: BucketStatus) => {
    switch (status) {
        case "healthy": return "bg-green-500 hover:bg-green-400";
        case "degraded": return "bg-amber-500 hover:bg-amber-400";
        case "unhealthy": return "bg-red-500 hover:bg-red-400";
        case "inactive": return "bg-slate-300 dark:bg-slate-700 opacity-50";
        default: return "bg-slate-200 dark:bg-slate-800";
    }
};

export function StatusTimeline({ history, totalDurationMs = 24 * 60 * 60 * 1000, bucketCount = 96 }: StatusTimelineProps) {
    const buckets = useMemo(() => {
        const now = Date.now();
        const start = now - totalDurationMs;
        const bucketDuration = totalDurationMs / bucketCount;

        const result: Bucket[] = [];

        // Initialize buckets
        for (let i = 0; i < bucketCount; i++) {
            result.push({
                startTime: start + i * bucketDuration,
                endTime: start + (i + 1) * bucketDuration,
                status: "no-data",
                count: 0,
                errorCount: 0,
            });
        }

        // Fill buckets
        if (history && history.length > 0) {
            history.forEach(point => {
                if (point.timestamp < start) return; // Too old
                const bucketIndex = Math.floor((point.timestamp - start) / bucketDuration);
                if (bucketIndex >= 0 && bucketIndex < bucketCount) {
                    const bucket = result[bucketIndex];
                    bucket.count++;

                    // Update worst status logic
                    const currentPriority = statusPriority(bucket.status);
                    const newPriority = statusPriority(point.status);

                    if (newPriority > currentPriority) {
                        bucket.status = point.status;
                    } else if (bucket.status === "no-data") {
                        bucket.status = point.status;
                    }

                    if (point.status === "unhealthy") {
                        bucket.errorCount++;
                    }
                }
            });
        }

        return result;
    }, [history, totalDurationMs, bucketCount]);

    return (
        <div className="flex items-center gap-[2px] w-full h-8 select-none">
            {buckets.map((bucket, i) => (
                <Tooltip key={i} delayDuration={0}>
                    <TooltipTrigger asChild>
                        <div
                            className={cn(
                                "flex-1 h-full rounded-[1px] transition-all",
                                getStatusColor(bucket.status),
                                bucket.status === "no-data" && "opacity-30"
                            )}
                        />
                    </TooltipTrigger>
                    <TooltipContent className="text-xs p-2">
                        <div className="font-semibold mb-1">
                            {new Date(bucket.startTime).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})} -
                            {new Date(bucket.endTime).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})}
                        </div>
                        <div className="capitalize">
                            Status: <span className={cn(
                                "font-bold",
                                bucket.status === "healthy" ? "text-green-500" :
                                bucket.status === "unhealthy" ? "text-red-500" :
                                bucket.status === "degraded" ? "text-amber-500" : "text-muted-foreground"
                            )}>{bucket.status}</span>
                        </div>
                        {bucket.count > 0 && (
                            <div className="text-[10px] text-muted-foreground mt-1">
                                Samples: {bucket.count} {bucket.errorCount > 0 && `(${bucket.errorCount} errors)`}
                            </div>
                        )}
                    </TooltipContent>
                </Tooltip>
            ))}
        </div>
    );
}
