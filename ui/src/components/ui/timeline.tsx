/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { cn } from "@/lib/utils";
import { ChevronDown, ChevronRight, CheckCircle2, AlertTriangle, XCircle, Clock } from "lucide-react";

export type TimelineStatus = 'success' | 'warning' | 'error' | 'pending';

export interface TimelineEvent {
  id: string;
  title: string;
  description?: string;
  timestamp: string;
  durationMs?: number;
  status: TimelineStatus;
  icon?: React.ReactNode;
  metadata?: Record<string, any>;
}

export interface TimelineProps {
  events: TimelineEvent[];
  className?: string;
}

const statusConfig = {
    success: {
        icon: CheckCircle2,
        colorClass: "text-emerald-500",
        bgClass: "bg-emerald-500/10",
        borderClass: "border-emerald-500/20"
    },
    warning: {
        icon: AlertTriangle,
        colorClass: "text-amber-500",
        bgClass: "bg-amber-500/10",
        borderClass: "border-amber-500/20"
    },
    error: {
        icon: XCircle,
        colorClass: "text-rose-500",
        bgClass: "bg-rose-500/10",
        borderClass: "border-rose-500/20"
    },
    pending: {
        icon: Clock,
        colorClass: "text-muted-foreground",
        bgClass: "bg-muted/10",
        borderClass: "border-muted"
    }
};

function TimelineNode({ event, isLast }: { event: TimelineEvent; isLast: boolean }) {
    const [expanded, setExpanded] = useState(false);
    const config = statusConfig[event.status] || statusConfig.pending;
    const StatusIcon = event.icon ? () => <>{event.icon}</> : config.icon;

    return (
        <div className="relative flex w-full pb-8 group">
            {/* Vertical Line */}
            {!isLast && (
                <div className="absolute left-6 top-10 bottom-0 w-px bg-border group-hover:bg-primary/30 transition-colors" />
            )}

            {/* Icon Node */}
            <div className={cn(
                "relative z-10 flex h-12 w-12 items-center justify-center rounded-full border-2 transition-transform duration-300 ease-in-out",
                config.bgClass,
                config.borderClass,
                "group-hover:scale-110 shadow-sm"
            )}>
                <StatusIcon className={cn("h-5 w-5", config.colorClass)} />
            </div>

            {/* Content Card */}
            <div className="ml-6 flex-1 pt-1">
                <div
                    className={cn(
                        "rounded-lg border bg-card p-4 shadow-sm transition-all duration-300 ease-in-out cursor-pointer hover:shadow-md",
                        expanded ? "ring-1 ring-primary/20" : ""
                    )}
                    onClick={() => event.metadata && setExpanded(!expanded)}
                >
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                            <h3 className="font-semibold text-sm text-foreground">{event.title}</h3>
                            {event.durationMs !== undefined && (
                                <span className="rounded bg-muted px-2 py-0.5 text-[10px] font-mono text-muted-foreground">
                                    {event.durationMs}ms
                                </span>
                            )}
                        </div>
                        <div className="flex items-center gap-3">
                            <span className="text-[10px] font-mono text-muted-foreground">{event.timestamp}</span>
                            {event.metadata && (
                                expanded ? <ChevronDown className="h-4 w-4 text-muted-foreground" /> : <ChevronRight className="h-4 w-4 text-muted-foreground" />
                            )}
                        </div>
                    </div>

                    {event.description && (
                        <p className="mt-2 text-xs text-muted-foreground">{event.description}</p>
                    )}

                    {/* Expandable Metadata Area */}
                    <div className={cn(
                        "overflow-hidden transition-all duration-300 ease-in-out",
                        expanded ? "max-h-[500px] mt-4 opacity-100" : "max-h-0 opacity-0"
                    )}>
                        {event.metadata && (
                            <div className="rounded-md bg-muted/50 p-3 border border-border/50">
                                <pre className="text-[10px] font-mono text-muted-foreground overflow-auto">
                                    {JSON.stringify(event.metadata, null, 2)}
                                </pre>
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
}

export function PremiumTimeline({ events, className }: TimelineProps) {
    if (!events || events.length === 0) {
        return (
            <div className="flex h-32 items-center justify-center text-sm text-muted-foreground border border-dashed rounded-lg">
                No execution trace available.
            </div>
        );
    }

    return (
        <div className={cn("relative w-full py-4", className)}>
            <div className="mx-auto max-w-3xl">
                {events.map((event, index) => (
                    <TimelineNode
                        key={event.id}
                        event={event}
                        isLast={index === events.length - 1}
                    />
                ))}
            </div>
        </div>
    );
}
