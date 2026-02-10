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
import { format } from "date-fns";
import { CalendarIcon, Search, RefreshCw, Eye, AlertTriangle, Play, Download } from "lucide-react";
import { cn } from "@/lib/utils";
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { ActivityChart } from "@/components/audit/activity-chart";

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

    // Filters
    const [toolName, setToolName] = useState("all");
    const [userId, setUserId] = useState("all");
    const [startDate, setStartDate] = useState<Date | undefined>(undefined);
    const [endDate, setEndDate] = useState<Date | undefined>(undefined);

    const [knownTools, setKnownTools] = useState<string[]>([]);
    const [knownUsers, setKnownUsers] = useState<string[]>([]);
    const [isSeeding, setIsSeeding] = useState(false);

    // Fetch filters data
    useEffect(() => {
        const loadFilters = async () => {
            try {
                const toolsRes = await apiClient.listTools();
                const usersRes = await apiClient.listUsers();

                if (toolsRes && toolsRes.tools) {
                    setKnownTools(Array.from(new Set(toolsRes.tools.map((t: any) => t.name))).sort());
                }

                // Users might be array or object depending on API
                const usersList = Array.isArray(usersRes) ? usersRes : (usersRes.users || []);
                setKnownUsers(Array.from(new Set(usersList.map((u: any) => u.id))).sort());
            } catch (e) {
                console.error("Failed to load filters", e);
            }
        };
        loadFilters();
    }, []);

    const fetchLogs = useCallback(async () => {
        setLoading(true);
        try {
            const filters: any = {
                limit: 100, // Increased limit
                offset: 0
            };
            if (toolName && toolName !== "all") filters.tool_name = toolName;
            if (userId && userId !== "all") filters.user_id = userId;
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

    const handleSeed = async () => {
        setIsSeeding(true);
        try {
            await fetch("/api/v1/debug/seed_audit?count=50", { method: "POST" });
            fetchLogs();
        } catch (e) {
            console.error("Seed failed", e);
        } finally {
            setIsSeeding(false);
        }
    };

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
            <ActivityChart data={logs} loading={loading && logs.length === 0} />

            <Card className="flex-none">
                <CardHeader className="pb-3">
                    <div className="flex justify-between items-center">
                        <div>
                            <CardTitle>Filters</CardTitle>
                            <CardDescription>Search audit logs by tool, user, or date.</CardDescription>
                        </div>
                        <div className="flex gap-2">
                             <Button variant="outline" size="sm" onClick={handleSeed} disabled={isSeeding}>
                                <Play className="mr-2 h-4 w-4" />
                                {isSeeding ? "Simulating..." : "Simulate Traffic"}
                            </Button>
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    <div className="flex flex-col md:flex-row gap-4 items-end">
                        <div className="grid gap-2 flex-1 w-full md:w-auto">
                            <label className="text-sm font-medium">Tool Name</label>
                            <Select value={toolName} onValueChange={setToolName}>
                                <SelectTrigger>
                                    <SelectValue placeholder="All Tools" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">All Tools</SelectItem>
                                    {knownTools.map(t => (
                                        <SelectItem key={t} value={t}>{t}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="grid gap-2 flex-1 w-full md:w-auto">
                            <label className="text-sm font-medium">User ID</label>
                            <Select value={userId} onValueChange={setUserId}>
                                <SelectTrigger>
                                    <SelectValue placeholder="All Users" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">All Users</SelectItem>
                                    {knownUsers.map(u => (
                                        <SelectItem key={u} value={u}>{u}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
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

            <Card className="flex-1 flex flex-col overflow-hidden">
                <CardContent className="p-0 flex-1 overflow-auto">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead className="w-[180px]">Timestamp</TableHead>
                                <TableHead>Tool</TableHead>
                                <TableHead>User</TableHead>
                                <TableHead>Duration</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead className="text-right">Details</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {logs.length === 0 && !loading && (
                                <TableRow>
                                    <TableCell colSpan={6} className="text-center h-24 text-muted-foreground">
                                        No logs found.
                                    </TableCell>
                                </TableRow>
                            )}
                            {logs.map((log, i) => (
                                <TableRow key={i}>
                                    <TableCell className="font-mono text-xs">
                                        {new Date(log.timestamp).toLocaleString()}
                                    </TableCell>
                                    <TableCell className="font-medium">{log.toolName}</TableCell>
                                    <TableCell>{log.userId || "-"}</TableCell>
                                    <TableCell>{log.duration}</TableCell>
                                    <TableCell>
                                        {log.error ? (
                                            <Badge variant="destructive" className="gap-1">
                                                <AlertTriangle className="h-3 w-3" /> Error
                                            </Badge>
                                        ) : (
                                            <Badge variant="outline" className="text-green-500 border-green-500/50">
                                                Success
                                            </Badge>
                                        )}
                                    </TableCell>
                                    <TableCell className="text-right">
                                        <Button variant="ghost" size="sm" onClick={() => setSelectedLog(log)}>
                                            <Eye className="h-4 w-4 mr-1" /> View
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>

            <Sheet open={!!selectedLog} onOpenChange={(open) => !open && setSelectedLog(null)}>
                <SheetContent className="sm:max-w-2xl overflow-y-auto">
                    <SheetHeader>
                        <SheetTitle>Audit Log Detail</SheetTitle>
                        <SheetDescription>
                            Execution details for {selectedLog?.toolName}
                        </SheetDescription>
                    </SheetHeader>
                    {selectedLog && (
                        <div className="space-y-6 mt-6">
                            <div className="flex items-center justify-between">
                                <Badge variant={selectedLog.error ? "destructive" : "outline"} className={selectedLog.error ? "" : "text-green-500 border-green-500/50"}>
                                    {selectedLog.error ? "Failed" : "Success"}
                                </Badge>
                                <span className="text-sm text-muted-foreground font-mono">
                                    {new Date(selectedLog.timestamp).toLocaleString()}
                                </span>
                            </div>

                            <div className="grid grid-cols-2 gap-4 text-sm border p-4 rounded-lg bg-muted/20">
                                <div>
                                    <span className="text-xs font-medium text-muted-foreground block mb-1">User ID</span>
                                    <span className="font-mono">{selectedLog.userId || "N/A"}</span>
                                </div>
                                <div>
                                    <span className="text-xs font-medium text-muted-foreground block mb-1">Profile ID</span>
                                    <span className="font-mono">{selectedLog.profileId || "N/A"}</span>
                                </div>
                                <div>
                                    <span className="text-xs font-medium text-muted-foreground block mb-1">Duration</span>
                                    <span className="font-mono">{selectedLog.duration} ({selectedLog.durationMs}ms)</span>
                                </div>
                                <div>
                                    <span className="text-xs font-medium text-muted-foreground block mb-1">Tool</span>
                                    <span className="font-mono truncate block" title={selectedLog.toolName}>{selectedLog.toolName}</span>
                                </div>
                            </div>

                            {selectedLog.error && (
                                <div className="bg-red-900/20 border border-red-900/50 rounded-md p-3 text-red-200 text-sm">
                                    <span className="font-semibold block mb-1 flex items-center gap-2">
                                        <AlertTriangle className="h-4 w-4" /> Error
                                    </span>
                                    <pre className="whitespace-pre-wrap font-mono text-xs mt-2">{selectedLog.error}</pre>
                                </div>
                            )}

                            <div>
                                <h4 className="text-sm font-medium mb-2">Arguments</h4>
                                <div className="rounded-md overflow-hidden border">
                                    <SyntaxHighlighter
                                        language="json"
                                        style={vscDarkPlus}
                                        customStyle={{ margin: 0, fontSize: '12px' }}
                                    >
                                        {formatJson(selectedLog.arguments) || "{}"}
                                    </SyntaxHighlighter>
                                </div>
                            </div>

                            <div>
                                <h4 className="text-sm font-medium mb-2">Result</h4>
                                <div className="rounded-md overflow-hidden border">
                                    <SyntaxHighlighter
                                        language="json"
                                        style={vscDarkPlus}
                                        customStyle={{ margin: 0, fontSize: '12px', maxHeight: '400px' }}
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
