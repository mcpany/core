/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useCallback, useMemo } from "react";
import { apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
} from "@/components/ui/sheet";
import { Badge } from "@/components/ui/badge";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { format, subHours, startOfHour, getHours } from "date-fns";
import { CalendarIcon, Search, RefreshCw, Eye, AlertTriangle, Play, ChevronLeft, ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { BarChart, Bar, XAxis, YAxis, Tooltip as RechartsTooltip, ResponsiveContainer } from 'recharts';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useToast } from "@/hooks/use-toast";

interface AuditLogEntry {
    timestamp: string;
    toolName: string;
    userId: string;
    profileId: string;
    arguments: string;
    result: string;
    error: string;
    duration: string;
    durationMs: number;
}

/**
 * AuditLogViewer component.
 * Displays a table of audit logs with filtering capabilities and detailed view.
 *
 * @returns The rendered AuditLogViewer component.
 */
export function AuditLogViewer() {
    const [logs, setLogs] = useState<AuditLogEntry[]>([]);
    const [loading, setLoading] = useState(true);
    const [selectedLog, setSelectedLog] = useState<AuditLogEntry | null>(null);
    const [seeding, setSeeding] = useState(false);
    const { toast } = useToast();

    // Filters
    const [toolName, setToolName] = useState("all");
    const [userId, setUserId] = useState("all");
    const [status, setStatus] = useState("all"); // all, success, error
    const [startDate, setStartDate] = useState<Date | undefined>(undefined);
    const [endDate, setEndDate] = useState<Date | undefined>(undefined);

    // Pagination
    const [page, setPage] = useState(0);
    const PAGE_SIZE = 50;

    const fetchLogs = useCallback(async () => {
        setLoading(true);
        try {
            const filters: any = {
                limit: PAGE_SIZE,
                offset: page * PAGE_SIZE
            };
            if (toolName && toolName !== "all") filters.tool_name = toolName;
            if (userId && userId !== "all") filters.user_id = userId;
            if (startDate) filters.start_time = startDate.toISOString();
            if (endDate) filters.end_time = endDate.toISOString();

            const res = await apiClient.listAuditLogs(filters);

            // Client-side filtering for Status if API doesn't support it (assuming it doesn't based on proto)
            let entries = res.entries || [];
            if (status !== "all") {
                entries = entries.filter((e: AuditLogEntry) => status === "error" ? !!e.error : !e.error);
            }

            setLogs(entries);
        } catch (e) {
            console.error("Failed to fetch audit logs", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to fetch logs."
            });
        } finally {
            setLoading(false);
        }
    }, [toolName, userId, startDate, endDate, page, status, toast]);

    useEffect(() => {
        fetchLogs();
    }, [fetchLogs]);

    const handleSeed = async () => {
        setSeeding(true);
        try {
            await apiClient.seedAuditLogs(50);
            toast({
                title: "Traffic Simulated",
                description: "Generated 50 audit log entries.",
            });
            fetchLogs();
        } catch (e) {
            console.error("Seed failed", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to seed data."
            });
        } finally {
            setSeeding(false);
        }
    };

    const chartData = useMemo(() => {
        // Aggregate logs by hour for the chart
        // Note: This is only for the *fetched* logs. In a real app, we'd want a separate "metrics" API.
        // But for "Perceived Quality", showing the distribution of the current view is helpful.
        const buckets: Record<string, { time: string; success: number; error: number }> = {};

        // Initialize buckets for last 12 fetch logs range or just map fetch logs
        // Let's just map the logs we have
        logs.forEach(log => {
            const date = new Date(log.timestamp);
            const key = format(startOfHour(date), "HH:mm");
            if (!buckets[key]) {
                buckets[key] = { time: key, success: 0, error: 0 };
            }
            if (log.error) {
                buckets[key].error++;
            } else {
                buckets[key].success++;
            }
        });

        return Object.values(buckets).sort((a, b) => a.time.localeCompare(b.time));
    }, [logs]);

    // Extract unique values for filters from current logs (better than nothing)
    const uniqueTools = useMemo(() => Array.from(new Set(logs.map(l => l.toolName))).sort(), [logs]);
    const uniqueUsers = useMemo(() => Array.from(new Set(logs.map(l => l.userId))).filter(Boolean).sort(), [logs]);

    const formatJson = (jsonStr: string) => {
        if (!jsonStr) return null;
        try {
            const obj = JSON.parse(jsonStr);
            return JSON.stringify(obj, null, 2);
        } catch (e) {
            return jsonStr;
        }
    };

    return (
        <div className="space-y-4 h-full flex flex-col">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <Card className="md:col-span-2">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium text-muted-foreground">Activity Volume</CardTitle>
                    </CardHeader>
                    <CardContent className="h-[200px] w-full">
                        <ResponsiveContainer width="100%" height="100%">
                            <BarChart data={chartData}>
                                <XAxis
                                    dataKey="time"
                                    stroke="#888888"
                                    fontSize={12}
                                    tickLine={false}
                                    axisLine={false}
                                />
                                <YAxis
                                    stroke="#888888"
                                    fontSize={12}
                                    tickLine={false}
                                    axisLine={false}
                                    tickFormatter={(value) => `${value}`}
                                />
                                <RechartsTooltip
                                    contentStyle={{ backgroundColor: 'rgba(0,0,0,0.8)', border: 'none', borderRadius: '4px', color: '#fff' }}
                                    cursor={{ fill: 'rgba(255,255,255,0.1)' }}
                                />
                                <Bar dataKey="success" name="Success" fill="#22c55e" radius={[4, 4, 0, 0]} stackId="a" />
                                <Bar dataKey="error" name="Error" fill="#ef4444" radius={[4, 4, 0, 0]} stackId="a" />
                            </BarChart>
                        </ResponsiveContainer>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium text-muted-foreground">Actions</CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-2">
                        <Button variant="outline" onClick={handleSeed} disabled={seeding} className="w-full justify-start">
                            {seeding ? <RefreshCw className="mr-2 h-4 w-4 animate-spin" /> : <Play className="mr-2 h-4 w-4" />}
                            Simulate Traffic (Demo)
                        </Button>
                        <Button variant="outline" onClick={fetchLogs} disabled={loading} className="w-full justify-start">
                            <RefreshCw className={cn("mr-2 h-4 w-4", loading && "animate-spin")} />
                            Refresh Logs
                        </Button>
                    </CardContent>
                </Card>
            </div>

            <Card className="flex-none">
                <CardContent className="pt-6">
                    <div className="flex flex-col md:flex-row gap-4 items-end">
                        <div className="grid gap-2 flex-1 w-full md:w-auto">
                            <label className="text-xs font-medium uppercase text-muted-foreground">Tool</label>
                            <Select value={toolName} onValueChange={setToolName}>
                                <SelectTrigger>
                                    <SelectValue placeholder="All Tools" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">All Tools</SelectItem>
                                    {uniqueTools.map(t => (
                                        <SelectItem key={t} value={t}>{t}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="grid gap-2 flex-1 w-full md:w-auto">
                            <label className="text-xs font-medium uppercase text-muted-foreground">User</label>
                            <Select value={userId} onValueChange={setUserId}>
                                <SelectTrigger>
                                    <SelectValue placeholder="All Users" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">All Users</SelectItem>
                                    {uniqueUsers.map(u => (
                                        <SelectItem key={u} value={u}>{u}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="grid gap-2 flex-1 w-full md:w-auto">
                            <label className="text-xs font-medium uppercase text-muted-foreground">Status</label>
                            <Select value={status} onValueChange={setStatus}>
                                <SelectTrigger>
                                    <SelectValue placeholder="All Status" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">All Status</SelectItem>
                                    <SelectItem value="success">Success</SelectItem>
                                    <SelectItem value="error">Error</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="grid gap-2 flex-1 w-full md:w-auto">
                            <label className="text-xs font-medium uppercase text-muted-foreground">Date Range</label>
                            <Popover>
                                <PopoverTrigger asChild>
                                    <Button
                                        variant={"outline"}
                                        className={cn(
                                            "w-full justify-start text-left font-normal",
                                            !startDate && "text-muted-foreground"
                                        )}
                                    >
                                        <CalendarIcon className="mr-2 h-4 w-4" />
                                        {startDate ? format(startDate, "PPP") : <span>Pick a date</span>}
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
                        </div>
                    </div>
                </CardContent>
            </Card>

            <Card className="flex-1 flex flex-col overflow-hidden border-t-0 rounded-t-none">
                <CardContent className="p-0 flex-1 overflow-auto">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead className="w-[180px]">Timestamp</TableHead>
                                <TableHead>Tool</TableHead>
                                <TableHead>User</TableHead>
                                <TableHead>Duration</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead className="text-right">Action</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {logs.length === 0 && !loading && (
                                <TableRow>
                                    <TableCell colSpan={6} className="text-center h-24 text-muted-foreground">
                                        No logs found matching your criteria.
                                    </TableCell>
                                </TableRow>
                            )}
                            {logs.map((log, i) => (
                                <TableRow
                                    key={i}
                                    className="cursor-pointer hover:bg-muted/50"
                                    onClick={() => setSelectedLog(log)}
                                >
                                    <TableCell className="font-mono text-xs text-muted-foreground">
                                        {format(new Date(log.timestamp), "MMM dd HH:mm:ss")}
                                    </TableCell>
                                    <TableCell className="font-medium">
                                        <span className="font-mono text-xs bg-muted px-1 py-0.5 rounded text-foreground">
                                            {log.toolName}
                                        </span>
                                    </TableCell>
                                    <TableCell>
                                        {log.userId ? (
                                            <div className="flex items-center gap-1">
                                                <div className="w-5 h-5 rounded-full bg-primary/10 flex items-center justify-center text-[10px] font-bold text-primary">
                                                    {log.userId[0].toUpperCase()}
                                                </div>
                                                <span className="text-xs">{log.userId}</span>
                                            </div>
                                        ) : "-"}
                                    </TableCell>
                                    <TableCell className="text-xs font-mono text-muted-foreground">
                                        {log.durationMs}ms
                                    </TableCell>
                                    <TableCell>
                                        {log.error ? (
                                            <Badge variant="destructive" className="gap-1 text-[10px] px-1.5 py-0">
                                                <AlertTriangle className="h-3 w-3" /> Error
                                            </Badge>
                                        ) : (
                                            <Badge variant="outline" className="text-green-500 border-green-500/30 bg-green-500/5 text-[10px] px-1.5 py-0">
                                                Success
                                            </Badge>
                                        )}
                                    </TableCell>
                                    <TableCell className="text-right">
                                        <Button variant="ghost" size="icon" className="h-8 w-8">
                                            <Eye className="h-4 w-4 opacity-50" />
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
                <div className="p-2 border-t flex items-center justify-between">
                    <span className="text-xs text-muted-foreground">
                        Page {page + 1}
                    </span>
                    <div className="flex gap-1">
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setPage(Math.max(0, page - 1))}
                            disabled={page === 0}
                        >
                            <ChevronLeft className="h-4 w-4" /> Previous
                        </Button>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setPage(page + 1)}
                            disabled={logs.length < PAGE_SIZE}
                        >
                            Next <ChevronRight className="h-4 w-4" />
                        </Button>
                    </div>
                </div>
            </Card>

            <Sheet open={!!selectedLog} onOpenChange={(open) => !open && setSelectedLog(null)}>
                <SheetContent className="sm:max-w-xl overflow-y-auto">
                    <SheetHeader>
                        <SheetTitle>Audit Log Detail</SheetTitle>
                        <SheetDescription>
                            Execution trace for {selectedLog?.toolName}
                        </SheetDescription>
                    </SheetHeader>

                    {selectedLog && (
                        <div className="space-y-6 mt-6">
                            <div className="grid grid-cols-2 gap-4 text-sm bg-muted/30 p-4 rounded-lg">
                                <div>
                                    <span className="text-xs font-medium text-muted-foreground uppercase">Timestamp</span>
                                    <div className="font-mono mt-1">{new Date(selectedLog.timestamp).toLocaleString()}</div>
                                </div>
                                <div>
                                    <span className="text-xs font-medium text-muted-foreground uppercase">Duration</span>
                                    <div className="font-mono mt-1">{selectedLog.duration} ({selectedLog.durationMs}ms)</div>
                                </div>
                                <div>
                                    <span className="text-xs font-medium text-muted-foreground uppercase">User</span>
                                    <div className="mt-1">{selectedLog.userId || "Anonymous"}</div>
                                </div>
                                <div>
                                    <span className="text-xs font-medium text-muted-foreground uppercase">Status</span>
                                    <div className="mt-1">
                                        {selectedLog.error ?
                                            <span className="text-red-500 font-medium">Failed</span> :
                                            <span className="text-green-500 font-medium">Success</span>
                                        }
                                    </div>
                                </div>
                            </div>

                            {selectedLog.error && (
                                <div className="bg-red-500/10 border border-red-500/20 rounded-md p-4 text-sm">
                                    <h4 className="font-semibold text-red-500 mb-1 flex items-center gap-2">
                                        <AlertTriangle className="h-4 w-4" /> Error Details
                                    </h4>
                                    <p className="text-red-700 dark:text-red-300 font-mono text-xs break-all">
                                        {selectedLog.error}
                                    </p>
                                </div>
                            )}

                            <div>
                                <h4 className="text-sm font-medium mb-2 flex items-center gap-2">
                                    Arguments
                                    <Badge variant="outline" className="font-mono text-[10px]">JSON</Badge>
                                </h4>
                                <div className="rounded-md overflow-hidden border bg-[#1e1e1e]">
                                    <SyntaxHighlighter
                                        language="json"
                                        style={vscDarkPlus}
                                        customStyle={{ margin: 0, fontSize: '12px', padding: '1rem' }}
                                        wrapLines={true}
                                        wrapLongLines={true}
                                    >
                                        {formatJson(selectedLog.arguments) || "{}"}
                                    </SyntaxHighlighter>
                                </div>
                            </div>

                            <div>
                                <h4 className="text-sm font-medium mb-2 flex items-center gap-2">
                                    Result
                                    <Badge variant="outline" className="font-mono text-[10px]">JSON</Badge>
                                </h4>
                                <div className="rounded-md overflow-hidden border bg-[#1e1e1e]">
                                    <SyntaxHighlighter
                                        language="json"
                                        style={vscDarkPlus}
                                        customStyle={{ margin: 0, fontSize: '12px', padding: '1rem', maxHeight: '400px' }}
                                        wrapLines={true}
                                        wrapLongLines={true}
                                    >
                                        {formatJson(selectedLog.result) || (selectedLog.error ? "null" : "{}")}
                                    </SyntaxHighlighter>
                                </div>
                            </div>
                        </div>
                    )}
                </SheetContent>
            </Sheet>
        </div>
    );
}
