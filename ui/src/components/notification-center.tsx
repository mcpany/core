/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, useCallback } from "react";
import { apiClient, Alert } from "@/lib/client";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Bell, Check, Trash2, Info, AlertTriangle, XCircle, AlertOctagon } from "lucide-react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { formatDistanceToNow } from "date-fns";

/**
 * NotificationCenter component.
 * Displays a bell icon with a badge for unread alerts and a popover list of alerts.
 * Allows dismissing individual alerts or all at once.
 */
export function NotificationCenter() {
    const [alerts, setAlerts] = useState<Alert[]>([]);
    const [isOpen, setIsOpen] = useState(false);
    const [loading, setLoading] = useState(false);

    const fetchAlerts = useCallback(async () => {
        try {
            // Fetch alerts using apiClient
            const data = await apiClient.listAlerts();
            // Ensure we have an array
            const list = Array.isArray(data) ? data : [];
            // Sort by timestamp desc (newest first) - backend does it but good to ensure
            list.sort((a: Alert, b: Alert) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
            setAlerts(list);
        } catch (error) {
            console.error("Failed to fetch alerts", error);
        }
    }, []);

    useEffect(() => {
        fetchAlerts();
        // Poll every 30 seconds
        const interval = setInterval(fetchAlerts, 30000);
        return () => clearInterval(interval);
    }, [fetchAlerts]);

    const handleDismiss = async (id: string, e: React.MouseEvent) => {
        e.stopPropagation();
        try {
            // Optimistic update
            setAlerts(prev => prev.map(a => a.id === id ? { ...a, status: "resolved" } : a));
            await apiClient.updateAlertStatus(id, "resolved");
        } catch (error) {
            console.error("Failed to dismiss alert", error);
            fetchAlerts(); // Revert
        }
    };

    const handleDismissAll = async () => {
        try {
            const activeAlerts = alerts.filter(a => a.status === "active");
            // Parallel update (limit concurrency in real app, but ok for now)
            await Promise.all(activeAlerts.map(a => apiClient.updateAlertStatus(a.id, "resolved")));
            fetchAlerts();
        } catch (error) {
            console.error("Failed to dismiss all alerts", error);
        }
    };

    const activeCount = alerts.filter(a => a.status === "active").length;

    const getSeverityIcon = (severity: string) => {
        switch (severity.toLowerCase()) {
            case "critical":
            case "error":
                return <XCircle className="h-4 w-4 text-red-500" />;
            case "warning":
                return <AlertTriangle className="h-4 w-4 text-yellow-500" />;
            case "info":
            default:
                return <Info className="h-4 w-4 text-blue-500" />;
        }
    };

    const getSeverityColor = (severity: string) => {
        switch (severity.toLowerCase()) {
            case "critical":
            case "error":
                return "border-l-red-500 bg-red-50/50 dark:bg-red-900/10";
            case "warning":
                return "border-l-yellow-500 bg-yellow-50/50 dark:bg-yellow-900/10";
            case "info":
            default:
                return "border-l-blue-500 bg-blue-50/50 dark:bg-blue-900/10";
        }
    };

    return (
        <Popover open={isOpen} onOpenChange={setIsOpen}>
            <PopoverTrigger asChild>
                <Button variant="ghost" size="icon" className="relative h-9 w-9">
                    <Bell className="h-4 w-4" />
                    {activeCount > 0 && (
                        <span className="absolute top-1.5 right-1.5 h-2 w-2 rounded-full bg-red-600 ring-2 ring-background animate-pulse" />
                    )}
                    <span className="sr-only">Notifications</span>
                </Button>
            </PopoverTrigger>
            <PopoverContent className="w-80 p-0" align="end">
                <div className="flex items-center justify-between px-4 py-2 border-b bg-muted/40">
                    <h4 className="font-semibold text-sm">Notifications</h4>
                    {activeCount > 0 && (
                        <Button variant="ghost" size="xs" className="h-auto px-1.5 text-xs text-muted-foreground hover:text-foreground" onClick={handleDismissAll}>
                            <Check className="h-3 w-3 mr-1" />
                            Mark all read
                        </Button>
                    )}
                </div>
                <ScrollArea className="h-[300px]">
                    {alerts.length === 0 ? (
                        <div className="flex flex-col items-center justify-center h-full text-muted-foreground p-8 text-center">
                            <Bell className="h-8 w-8 mb-2 opacity-20" />
                            <p className="text-xs">No notifications yet.</p>
                        </div>
                    ) : (
                        <div className="flex flex-col">
                            {alerts.map((alert) => (
                                <div
                                    key={alert.id}
                                    className={cn(
                                        "flex flex-col gap-1 p-4 border-b last:border-0 hover:bg-muted/50 transition-colors border-l-4",
                                        getSeverityColor(alert.severity),
                                        alert.status === "resolved" && "opacity-50 border-l-transparent bg-transparent"
                                    )}
                                >
                                    <div className="flex items-start justify-between gap-2">
                                        <div className="flex items-center gap-2 font-medium text-sm">
                                            {getSeverityIcon(alert.severity)}
                                            {alert.title}
                                        </div>
                                        <div className="text-[10px] text-muted-foreground whitespace-nowrap">
                                            {formatDistanceToNow(new Date(alert.timestamp), { addSuffix: true })}
                                        </div>
                                    </div>
                                    <p className="text-xs text-muted-foreground line-clamp-2 pl-6">
                                        {alert.message}
                                    </p>
                                    <div className="flex items-center justify-between mt-1 pl-6">
                                        <Badge variant="outline" className="text-[10px] h-4 px-1 font-normal bg-background/50">
                                            {alert.service}
                                        </Badge>
                                        {alert.status !== "resolved" && (
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                className="h-5 w-5 text-muted-foreground hover:text-foreground"
                                                onClick={(e) => handleDismiss(alert.id, e)}
                                                title="Dismiss"
                                            >
                                                <Check className="h-3 w-3" />
                                            </Button>
                                        )}
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </ScrollArea>
                <div className="p-2 border-t bg-muted/40 text-center">
                    <Button variant="link" size="sm" className="text-xs h-auto p-0" asChild>
                       {/* Link to full alerts page if we want, but for now just text */}
                       <a href="/alerts">View all alerts</a>
                    </Button>
                </div>
            </PopoverContent>
        </Popover>
    );
}
