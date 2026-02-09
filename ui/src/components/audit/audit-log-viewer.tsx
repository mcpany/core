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
import { Bar, BarChart, ResponsiveContainer, XAxis, YAxis, Tooltip } from "recharts";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
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
import { CalendarIcon, Search, RefreshCw, Eye, AlertTriangle, Zap } from "lucide-react";
import { cn } from "@/lib/utils";
import { useToast } from "@/hooks/use-toast";
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';

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
    const { toast } = useToast();
    const [logs, setLogs] = useState<AuditLogEntry[]>([]);
    const [loading, setLoading] = useState(true);
    const [selectedLog, setSelectedLog] = useState<AuditLogEntry | null>(null);

    // Filters
    const [toolName, setToolName] = useState("all");
    const [userId, setUserId] = useState("all");
    const [startDate, setStartDate] = useState<Date | undefined>(undefined);
    const [endDate, setEndDate] = useState<Date | undefined>(undefined);

    const [availableTools, setAvailableTools] = useState<string[]>([]);
    const [availableUsers, setAvailableUsers] = useState<string[]>([]);

    useEffect(() => {
        apiClient.listTools().then(res => setAvailableTools(res.tools.map((t: any) => t.name))).catch(console.error);
        apiClient.listUsers().then(res => {
             const users = Array.isArray(res) ? res : res.users || [];
             setAvailableUsers(users.map((u: any) => u.id));
        }).catch(console.error);
    }, []);

    const fetchLogs = useCallback(async () => {
        setLoading(true);
        try {
            const filters: any = {
                limit: 100, // Increased limit for better chart data
                offset: 0
            };
            if (toolName && toolName !== "all") filters.tool_name = toolName;
            if (userId && userId !== "all") filters.user_id = userId;
            if (startDate) filters.start_time = startDate.toISOString();
            if (endDate) filters.end_time = endDate.toISOString();

            const res = await apiClient.listAuditLogs(filters);
            // Map snake_case to camelCase manually if needed, but assuming client returns what server sends.
            // Server sends protobuf JSON which is camelCase by default for fields?
            // Actually, grpc-gateway default uses snake_case for JSON unless configured otherwise.
            // But I implemented manual marshalling in `server.go` using `AuditLogEntry` struct?
            // No, I used `pb.AuditLogEntry`. Protobuf JSON serialization uses camelCase by default in Go (protojson).
            // Let's assume camelCase.
            // Wait, looking at `admin.proto`:
            // string tool_name = 2;
            // In JSON it will be `toolName`.
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

    const chartData = useMemo(() => {
        const counts: Record<string, number> = {};
        logs.forEach(log => {
            const date = new Date(log.timestamp);
            const key = date.getHours() + ":00";
            counts[key] = (counts[key] || 0) + 1;
        });
        return Object.entries(counts)
            .map(([time, count]) => ({ time, count }))
            .sort((a, b) => parseInt(a.time) - parseInt(b.time));
    }, [logs]);

    const handleSeed = async () => {
        setLoading(true);
        try {
            await apiClient.seedAuditLogs(50);
            toast({ title: "Traffic Simulated", description: "Generated 50 audit logs." });
            await fetchLogs();
        } catch (e) {
            console.error(e);
            toast({ title: "Simulation Failed", variant: "destructive", description: String(e) });
        } finally {
            setLoading(false);
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
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <Card className="md:col-span-2">
                    <CardHeader className="pb-2">
                        <CardTitle>Activity Volume (Last 24h)</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="h-[120px] w-full">
                            <ResponsiveContainer width="100%" height="100%">
                                <BarChart data={chartData}>
                                    <XAxis dataKey="time" fontSize={12} tickLine={false} axisLine={false} />
                                    <YAxis fontSize={12} tickLine={false} axisLine={false} />
                                    <Tooltip
                                        contentStyle={{ backgroundColor: 'var(--background)', borderRadius: '8px', border: '1px solid var(--border)' }}
                                        labelStyle={{ color: 'var(--muted-foreground)' }}
                                    />
                                    <Bar dataKey="count" fill="currentColor" radius={[4, 4, 0, 0]} className="fill-primary" />
                                </BarChart>
                            </ResponsiveContainer>
                        </div>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="pb-2">
                        <CardTitle>Filters</CardTitle>
                        <CardDescription>Refine your view.</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                         <div className="space-y-2">
                            <label className="text-xs font-medium">Tool Name</label>
                            <Select value={toolName} onValueChange={setToolName}>
                                <SelectTrigger>
                                    <SelectValue placeholder="All Tools" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">All Tools</SelectItem>
                                    {availableTools.map(t => (
                                        <SelectItem key={t} value={t}>{t}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="space-y-2">
                            <label className="text-xs font-medium">User ID</label>
                            <Select value={userId} onValueChange={setUserId}>
                                <SelectTrigger>
                                    <SelectValue placeholder="All Users" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">All Users</SelectItem>
                                    {availableUsers.map(u => (
                                        <SelectItem key={u} value={u}>{u}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                    </CardContent>
                </Card>
            </div>

            <Card className="flex-none">
                <CardContent className="pt-6">
                    <div className="flex flex-col md:flex-row gap-4 items-end">
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
                        <Button variant="outline" onClick={handleSeed} disabled={loading}>
                            <Zap className="mr-2 h-4 w-4" /> Simulate Traffic
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
                <SheetContent className="sm:max-w-xl w-full overflow-y-auto">
                    <SheetHeader>
                        <SheetTitle>Audit Log Detail</SheetTitle>
                        <SheetDescription>
                            Execution details for {selectedLog?.toolName} at {selectedLog && new Date(selectedLog.timestamp).toLocaleString()}
                        </SheetDescription>
                    </SheetHeader>
                    {selectedLog && (
                        <div className="space-y-6 mt-6">
                            <div className="grid grid-cols-2 gap-4 text-sm">
                                <div>
                                    <span className="font-semibold block text-muted-foreground">User ID</span>
                                    {selectedLog.userId || "N/A"}
                                </div>
                                <div>
                                    <span className="font-semibold block text-muted-foreground">Profile ID</span>
                                    {selectedLog.profileId || "N/A"}
                                </div>
                                <div>
                                    <span className="font-semibold block text-muted-foreground">Duration</span>
                                    {selectedLog.duration} ({selectedLog.durationMs}ms)
                                </div>
                                <div>
                                    <span className="font-semibold block text-muted-foreground">Status</span>
                                    {selectedLog.error ? <span className="text-red-500">Failed</span> : <span className="text-green-500">Success</span>}
                                </div>
                            </div>

                            {selectedLog.error && (
                                <div className="bg-red-900/20 border border-red-900/50 rounded-md p-3 text-red-200 text-sm">
                                    <span className="font-semibold block mb-1">Error:</span>
                                    {selectedLog.error}
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
                                        customStyle={{ margin: 0, fontSize: '12px', maxHeight: '300px' }}
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
