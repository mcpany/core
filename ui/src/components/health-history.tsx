/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Activity } from "lucide-react";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

interface HealthPoint {
    time: number;
    status: string;
}

interface Segment {
    status: string;
    duration: number;
    startTime: number;
    endTime: number;
}

export function HealthHistoryCard({ serviceId }: { serviceId: string }) {
    const [history, setHistory] = useState<HealthPoint[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        setLoading(true);
        apiClient.getHealthHistory(serviceId)
            .then(data => setHistory(data))
            .catch(() => setHistory([]))
            .finally(() => setLoading(false));
    }, [serviceId]);

    if (loading) {
        return (
            <Card>
                <CardHeader>
                    <CardTitle className="text-xl flex items-center gap-2"><Activity /> Health History (24h)</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="h-8 bg-muted animate-pulse rounded" />
                </CardContent>
            </Card>
        );
    }

    const now = Date.now() / 1000;
    const start = now - 24 * 3600;
    const sorted = [...history].sort((a, b) => a.time - b.time);

    // Calculate segments
    const segments: Segment[] = [];
    let currentTime = start;
    let currentStatus = 'unknown';

    // Find status at start time (last event before start)
    const lastBeforeStart = sorted.filter(p => p.time < start).pop();
    if (lastBeforeStart) {
        currentStatus = lastBeforeStart.status;
    } else if (sorted.length > 0 && sorted[0].time > start) {
        // If we have history but it starts AFTER the window, assume unknown before that
        currentStatus = 'unknown';
    } else if (sorted.length > 0) {
        // Should be covered by first case if we have events before start
        currentStatus = sorted[0].status;
    }

    const inRange = sorted.filter(p => p.time >= start);
    const pointsToProcess = [...inRange, { time: now, status: 'end' }];

    for (const point of pointsToProcess) {
        // Clamp point time to now (in case of clock skew or future points)
        const pointTime = Math.min(point.time, now);
        const duration = pointTime - currentTime;

        if (duration > 0) {
            segments.push({
                status: currentStatus,
                duration: duration,
                startTime: currentTime,
                endTime: pointTime
            });
        }
        currentTime = pointTime;
        if (point.status !== 'end') currentStatus = point.status;
    }

    if (segments.length === 0) {
         return (
            <Card>
                <CardHeader>
                    <CardTitle className="text-xl flex items-center gap-2"><Activity /> Health History (24h)</CardTitle>
                </CardHeader>
                <CardContent>
                    <p className="text-muted-foreground text-sm">No health history available.</p>
                </CardContent>
            </Card>
        );
    }

    return (
        <Card>
            <CardHeader>
                <CardTitle className="text-xl flex items-center gap-2"><Activity /> Health History (24h)</CardTitle>
            </CardHeader>
            <CardContent>
                <div className="space-y-4">
                    <div className="flex w-full h-8 overflow-hidden rounded-md bg-muted/20 border border-border/50">
                        {segments.map((seg, i) => (
                            <TooltipProvider key={i}>
                                <Tooltip>
                                    <TooltipTrigger asChild>
                                        <div
                                            className={`h-full transition-all hover:brightness-110 ${
                                                seg.status === 'up' ? 'bg-green-500' :
                                                seg.status === 'down' ? 'bg-red-500' :
                                                seg.status === 'degraded' ? 'bg-yellow-500' :
                                                'bg-gray-300 dark:bg-gray-700'
                                            }`}
                                            style={{ width: `${(seg.duration / (24 * 3600)) * 100}%` }}
                                        />
                                    </TooltipTrigger>
                                    <TooltipContent>
                                        <div className="text-xs">
                                            <p className="font-semibold capitalize mb-1">{seg.status}</p>
                                            <p>{new Date(seg.startTime * 1000).toLocaleTimeString()} - {new Date(seg.endTime * 1000).toLocaleTimeString()}</p>
                                            <p className="text-muted-foreground">Duration: {Math.round(seg.duration / 60)} mins</p>
                                        </div>
                                    </TooltipContent>
                                </Tooltip>
                            </TooltipProvider>
                        ))}
                    </div>
                    <div className="flex justify-between text-xs text-muted-foreground">
                        <span>24 hours ago</span>
                        <span>Now</span>
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}
