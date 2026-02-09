/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useCallback } from "react";
import { apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { format, formatDistanceToNow } from "date-fns";
import { CalendarIcon, Search, RefreshCw, Eye, AlertTriangle, CheckCircle2, Clock, User, Terminal } from "lucide-react";
import { cn } from "@/lib/utils";
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';

interface AuditLogEntry {
    timestamp: string;
    tool_name: string;
    user_id?: string;
    profile_id?: string;
    arguments?: any;
    result?: any;
    error?: string;
    duration: string;
    duration_ms: number;
}

/**
 * AuditLogViewer component.
 * Displays a timeline of audit logs with filtering capabilities and detailed view.
 *
 * @returns The rendered AuditLogViewer component.
 */
export function AuditLogViewer() {
    const [logs, setLogs] = useState<AuditLogEntry[]>([]);
    const [loading, setLoading] = useState(true);
    const [selectedLog, setSelectedLog] = useState<AuditLogEntry | null>(null);

    // Filters
    const [toolName, setToolName] = useState("");
    const [userId, setUserId] = useState("");
    const [startDate, setStartDate] = useState<Date | undefined>(undefined);
    const [endDate, setEndDate] = useState<Date | undefined>(undefined);

    const fetchLogs = useCallback(async () => {
        setLoading(true);
        try {
            const filters: any = {
                limit: 100, // Increased limit for timeline
                offset: 0
            };
            if (toolName) filters.tool_name = toolName;
            if (userId) filters.user_id = userId;
            if (startDate) filters.start_time = startDate.toISOString();
            if (endDate) filters.end_time = endDate.toISOString();

            const res = await apiClient.listAuditLogs(filters);
            setLogs(res.entries || []);
        } catch (e) {
            console.error("Failed to fetch audit logs", e);
        } finally {
            setLoading(false);
        }
    }, [toolName, userId, startDate, endDate]);

    useEffect(() => {
        fetchLogs();
    }, [fetchLogs]);

    const formatJson = (data: any) => {
        if (!data) return "{}";
        try {
            return JSON.stringify(data, null, 2);
        } catch (e) {
            return String(data);
        }
    };

    return (
        <div className="space-y-4 h-full flex flex-col">
            <Card className="flex-none">
                <CardHeader className="pb-3">
                    <CardTitle>Filters</CardTitle>
                    <CardDescription>Search audit logs by tool, user, or date.</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="flex flex-col md:flex-row gap-4 items-end">
                        <div className="grid gap-2 flex-1 w-full md:w-auto">
                            <label className="text-sm font-medium">Tool Name</label>
                            <Input
                                placeholder="e.g. weather_get"
                                value={toolName}
                                onChange={(e) => setToolName(e.target.value)}
                            />
                        </div>
                        <div className="grid gap-2 flex-1 w-full md:w-auto">
                            <label className="text-sm font-medium">User ID</label>
                            <Input
                                placeholder="e.g. alice"
                                value={userId}
                                onChange={(e) => setUserId(e.target.value)}
                            />
                        </div>
                        <div className="grid gap-2 flex-1 w-full md:w-auto">
                            <label className="text-sm font-medium">Date Range</label>
                            <div className="flex gap-2">
                                <Popover>
                                    <PopoverTrigger asChild>
                                        <Button
                                            variant={"outline"}
                                            className={cn(
                                                "w-[140px] justify-start text-left font-normal",
                                                !startDate && "text-muted-foreground"
                                            )}
                                        >
                                            <CalendarIcon className="mr-2 h-4 w-4" />
                                            {startDate ? format(startDate, "PPP") : <span>Start Date</span>}
                                        </Button>
                                    </PopoverTrigger>
                                    <PopoverContent className="w-auto p-0" align="start">
                                        <Calendar
                                            mode="single"
                                            selected={startDate}
                                            onSelect={setStartDate}
                                            initialFocus
                                        />
                                    </PopoverContent>
                                </Popover>
                                <Popover>
                                    <PopoverTrigger asChild>
                                        <Button
                                            variant={"outline"}
                                            className={cn(
                                                "w-[140px] justify-start text-left font-normal",
                                                !endDate && "text-muted-foreground"
                                            )}
                                        >
                                            <CalendarIcon className="mr-2 h-4 w-4" />
                                            {endDate ? format(endDate, "PPP") : <span>End Date</span>}
                                        </Button>
                                    </PopoverTrigger>
                                    <PopoverContent className="w-auto p-0" align="start">
                                        <Calendar
                                            mode="single"
                                            selected={endDate}
                                            onSelect={setEndDate}
                                            initialFocus
                                        />
                                    </PopoverContent>
                                </Popover>
                            </div>
                        </div>
                        <Button onClick={fetchLogs} disabled={loading}>
                            {loading ? <RefreshCw className="mr-2 h-4 w-4 animate-spin" /> : <Search className="mr-2 h-4 w-4" />}
                            Filter
                        </Button>
                    </div>
                </CardContent>
            </Card>

            <div className="flex-1 flex flex-col overflow-hidden space-y-4">
                {logs.length === 0 && !loading && (
                    <div className="flex-1 flex flex-col items-center justify-center text-muted-foreground border-2 border-dashed rounded-lg bg-muted/10">
                        <Terminal className="h-10 w-10 mb-4 opacity-50" />
                        <p>No audit logs found matching your criteria.</p>
                    </div>
                )}

                <div className="flex-1 overflow-y-auto space-y-4 pr-2">
                    {logs.map((log, i) => (
                        <Card key={i} className="group hover:border-primary/50 transition-colors cursor-pointer" onClick={() => setSelectedLog(log)}>
                            <CardContent className="p-4 flex flex-col md:flex-row gap-4 items-start md:items-center">
                                <div className="flex-none flex flex-col items-center gap-1 w-16 text-xs text-muted-foreground">
                                    <span className="font-mono">{format(new Date(log.timestamp), "HH:mm:ss")}</span>
                                    <span className="text-[10px]">{formatDistanceToNow(new Date(log.timestamp), { addSuffix: true })}</span>
                                </div>

                                <div className="flex-none">
                                    {log.error ? (
                                        <div className="h-10 w-10 rounded-full bg-red-500/10 flex items-center justify-center text-red-500">
                                            <AlertTriangle className="h-5 w-5" />
                                        </div>
                                    ) : (
                                        <div className="h-10 w-10 rounded-full bg-green-500/10 flex items-center justify-center text-green-500">
                                            <CheckCircle2 className="h-5 w-5" />
                                        </div>
                                    )}
                                </div>

                                <div className="flex-1 min-w-0">
                                    <div className="flex items-center gap-2 mb-1">
                                        <span className="font-semibold text-lg">{log.tool_name}</span>
                                        {log.error && <Badge variant="destructive">Error</Badge>}
                                    </div>
                                    <div className="flex items-center gap-4 text-sm text-muted-foreground">
                                        <div className="flex items-center gap-1">
                                            <User className="h-3 w-3" />
                                            {log.user_id || "Anonymous"}
                                        </div>
                                        <div className="flex items-center gap-1">
                                            <Clock className="h-3 w-3" />
                                            {log.duration}
                                        </div>
                                        {log.profile_id && (
                                            <div className="flex items-center gap-1">
                                                <Badge variant="outline" className="text-[10px] h-4 px-1 py-0">{log.profile_id}</Badge>
                                            </div>
                                        )}
                                    </div>
                                </div>

                                <div className="flex-none opacity-0 group-hover:opacity-100 transition-opacity">
                                    <Button variant="ghost" size="sm">View Details</Button>
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            </div>

            <Dialog open={!!selectedLog} onOpenChange={(open) => !open && setSelectedLog(null)}>
                <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
                    <DialogHeader>
                        <DialogTitle>Audit Log Detail</DialogTitle>
                        <DialogDescription>
                            Execution details for {selectedLog?.tool_name}
                        </DialogDescription>
                    </DialogHeader>
                    {selectedLog && (
                        <div className="space-y-6">
                            <div className="grid grid-cols-2 gap-4 text-sm bg-muted/30 p-4 rounded-lg">
                                <div>
                                    <span className="font-semibold block text-muted-foreground text-xs uppercase tracking-wider mb-1">Timestamp</span>
                                    {format(new Date(selectedLog.timestamp), "PPP pp")}
                                </div>
                                <div>
                                    <span className="font-semibold block text-muted-foreground text-xs uppercase tracking-wider mb-1">User ID</span>
                                    {selectedLog.user_id || "N/A"}
                                </div>
                                <div>
                                    <span className="font-semibold block text-muted-foreground text-xs uppercase tracking-wider mb-1">Duration</span>
                                    {selectedLog.duration} ({selectedLog.duration_ms}ms)
                                </div>
                                <div>
                                    <span className="font-semibold block text-muted-foreground text-xs uppercase tracking-wider mb-1">Profile ID</span>
                                    {selectedLog.profile_id || "N/A"}
                                </div>
                            </div>

                            {selectedLog.error && (
                                <div className="bg-red-900/20 border border-red-900/50 rounded-md p-4 text-red-200 text-sm">
                                    <span className="font-semibold block mb-1 flex items-center gap-2">
                                        <AlertTriangle className="h-4 w-4" /> Execution Error
                                    </span>
                                    <div className="font-mono mt-2 whitespace-pre-wrap">{selectedLog.error}</div>
                                </div>
                            )}

                            <div>
                                <h4 className="text-sm font-medium mb-2 flex items-center gap-2">
                                    <Terminal className="h-4 w-4" /> Arguments
                                </h4>
                                <div className="rounded-md overflow-hidden border">
                                    <SyntaxHighlighter
                                        language="json"
                                        style={vscDarkPlus}
                                        customStyle={{ margin: 0, fontSize: '12px', maxHeight: '300px' }}
                                    >
                                        {formatJson(selectedLog.arguments)}
                                    </SyntaxHighlighter>
                                </div>
                            </div>

                            <div>
                                <h4 className="text-sm font-medium mb-2 flex items-center gap-2">
                                    <CheckCircle2 className="h-4 w-4" /> Result
                                </h4>
                                <div className="rounded-md overflow-hidden border">
                                    <SyntaxHighlighter
                                        language="json"
                                        style={vscDarkPlus}
                                        customStyle={{ margin: 0, fontSize: '12px', maxHeight: '400px' }}
                                    >
                                        {formatJson(selectedLog.result)}
                                    </SyntaxHighlighter>
                                </div>
                            </div>
                        </div>
                    )}
                </DialogContent>
            </Dialog>
        </div>
    );
}
