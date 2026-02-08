/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { AuditLogEntry } from "./audit-log-viewer";
import { format, isSameDay, parseISO } from "date-fns";
import { CheckCircle2, XCircle, Clock, User, Terminal } from "lucide-react";
import { cn } from "@/lib/utils";
import { ScrollArea } from "@/components/ui/scroll-area";

interface AuditTimelineProps {
    logs: AuditLogEntry[];
    onSelectLog: (log: AuditLogEntry) => void;
}

/**
 * AuditTimeline component.
 * Displays a vertical timeline of audit logs grouped by day.
 *
 * @param props - The component props.
 * @param props.logs - The list of audit logs to display.
 * @param props.onSelectLog - Callback when a log entry is selected.
 * @returns The rendered component.
 */
export function AuditTimeline({ logs, onSelectLog }: AuditTimelineProps) {
    // Group logs by day
    const groupedLogs = logs.reduce((acc, log) => {
        const date = parseISO(log.timestamp);
        const dayKey = format(date, "yyyy-MM-dd");
        if (!acc[dayKey]) {
            acc[dayKey] = [];
        }
        acc[dayKey].push(log);
        return acc;
    }, {} as Record<string, AuditLogEntry[]>);

    // Sort days descending
    const sortedDays = Object.keys(groupedLogs).sort((a, b) => b.localeCompare(a));

    if (logs.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center h-full text-muted-foreground p-8">
                <Clock className="h-12 w-12 mb-4 opacity-50" />
                <p>No audit activity found.</p>
            </div>
        );
    }

    return (
        <ScrollArea className="h-full pr-4">
            <div className="space-y-8 p-1">
                {sortedDays.map((day) => (
                    <div key={day} className="relative">
                        <div className="sticky top-0 z-10 bg-background/95 backdrop-blur py-2 mb-4 border-b">
                            <h3 className="text-sm font-semibold text-muted-foreground">
                                {format(parseISO(day), "EEEE, MMMM do, yyyy")}
                            </h3>
                        </div>

                        <div className="space-y-6 ml-2 border-l-2 border-muted pl-6 relative">
                            {groupedLogs[day].map((log, index) => (
                                <div
                                    key={index}
                                    className="relative group cursor-pointer transition-all hover:translate-x-1"
                                    onClick={() => onSelectLog(log)}
                                >
                                    {/* Timeline Node */}
                                    <div className={cn(
                                        "absolute -left-[31px] top-1 h-4 w-4 rounded-full border-2 bg-background flex items-center justify-center ring-4 ring-background transition-colors",
                                        log.error ? "border-red-500" : "border-green-500",
                                        "group-hover:scale-110"
                                    )}>
                                        {log.error ? (
                                            <div className="h-1.5 w-1.5 rounded-full bg-red-500" />
                                        ) : (
                                            <div className="h-1.5 w-1.5 rounded-full bg-green-500" />
                                        )}
                                    </div>

                                    {/* Content Card */}
                                    <div className={cn(
                                        "rounded-lg border bg-card text-card-foreground shadow-sm p-4 hover:shadow-md hover:border-primary/50 transition-all",
                                        log.error && "border-red-200 dark:border-red-900 bg-red-50/10"
                                    )}>
                                        <div className="flex justify-between items-start mb-2">
                                            <div className="flex items-center gap-2">
                                                <span className={cn(
                                                    "font-mono text-sm font-bold",
                                                    log.error ? "text-red-600 dark:text-red-400" : "text-primary"
                                                )}>
                                                    {log.tool_name}
                                                </span>
                                                {log.error && (
                                                    <span className="text-[10px] bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300 px-1.5 py-0.5 rounded-full font-medium">
                                                        Failed
                                                    </span>
                                                )}
                                            </div>
                                            <span className="text-xs text-muted-foreground font-mono">
                                                {format(parseISO(log.timestamp), "HH:mm:ss")}
                                            </span>
                                        </div>

                                        <div className="flex flex-wrap gap-x-4 gap-y-2 text-xs text-muted-foreground">
                                            <div className="flex items-center gap-1">
                                                <User className="h-3 w-3" />
                                                <span>{log.user_id || "System"}</span>
                                            </div>
                                            <div className="flex items-center gap-1">
                                                <Clock className="h-3 w-3" />
                                                <span>{log.duration_ms}ms</span>
                                            </div>
                                            {log.profile_id && (
                                                <div className="flex items-center gap-1">
                                                    <Terminal className="h-3 w-3" />
                                                    <span>{log.profile_id}</span>
                                                </div>
                                            )}
                                        </div>

                                        {log.error && (
                                            <div className="mt-3 text-xs text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/10 p-2 rounded border border-red-100 dark:border-red-900/20 font-mono break-all">
                                                {log.error}
                                            </div>
                                        )}
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                ))}
            </div>
        </ScrollArea>
    );
}
