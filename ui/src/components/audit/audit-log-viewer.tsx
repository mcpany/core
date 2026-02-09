/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useCallback, useMemo } from "react";
import { apiClient, ToolDefinition } from "@/lib/client";
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
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { format } from "date-fns";
import { CalendarIcon, Search, RefreshCw, Eye, AlertTriangle, Zap, Activity } from "lucide-react";
import { cn } from "@/lib/utils";
import { useToast } from "@/hooks/use-toast";
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import {
    BarChart,
    Bar,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    ResponsiveContainer,
    Legend
} from 'recharts';

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
    const { toast } = useToast();

    // Suggestion Lists
    const [availableTools, setAvailableTools] = useState<string[]>([]);
    const [availableUsers, setAvailableUsers] = useState<string[]>([]);

    // Filters
    const [toolName, setToolName] = useState("ALL");
    const [userId, setUserId] = useState("ALL");
    const [startDate, setStartDate] = useState<Date | undefined>(undefined);
    const [endDate, setEndDate] = useState<Date | undefined>(undefined);

    // Initial Load
    useEffect(() => {
        const loadMetadata = async () => {
            try {
                const [toolsRes, usersRes] = await Promise.all([
                    apiClient.listTools(),
                    apiClient.listUsers()
                ]);

                // Extract tool names
                const toolNames = (toolsRes.tools || []).map((t: ToolDefinition) => t.name).sort();
                setAvailableTools(toolNames);

                // Extract user IDs
                const users = Array.isArray(usersRes) ? usersRes : (usersRes.users || []);
                const userIds = users.map((u: any) => u.id).sort();
                setAvailableUsers(userIds);

            } catch (e) {
                console.error("Failed to load metadata", e);
            }
        };
        loadMetadata();
        fetchLogs();
    }, []);

    const fetchLogs = useCallback(async () => {
        setLoading(true);
        try {
            const filters: any = {
                limit: 100, // Increase limit for better charts
                offset: 0
            };
            if (toolName && toolName !== "ALL") filters.tool_name = toolName;
            if (userId && userId !== "ALL") filters.user_id = userId;
            if (startDate) filters.start_time = startDate.toISOString();
            if (endDate) filters.end_time = endDate.toISOString();

            const res = await apiClient.listAuditLogs(filters);
            setLogs(res.entries || []);
        } catch (e) {
            console.error("Failed to fetch audit logs", e);
            toast({ variant: "destructive", title: "Error", description: "Failed to fetch logs." });
        } finally {
            setLoading(false);
        }
    }, [toolName, userId, startDate, endDate, toast]);

    const handleSeed = async () => {
        setLoading(true);
        try {
            await apiClient.seedAuditLogs(50);
            toast({ title: "Traffic Simulated", description: "Generated 50 random audit logs." });
            // Wait a bit for indexing if needed, then fetch
            setTimeout(fetchLogs, 500);
        } catch (e) {
            console.error("Failed to seed logs", e);
             toast({ variant: "destructive", title: "Seeding Failed", description: "Could not generate traffic." });
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

    // Aggregate data for chart
    const chartData = useMemo(() => {
        const data: Record<string, { time: string; success: number; error: number }> = {};

        // Sort logs by time ascending for the chart
        const sortedLogs = [...logs].sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());

        if (sortedLogs.length === 0) return [];

        sortedLogs.forEach(log => {
            const date = new Date(log.timestamp);
            // Group by Hour if span > 24h, else by Minute?
            // For simplicity/demo: Group by Hour:Minute (every 10 mins?) or just Hour.
            // Let's do Hour:Minute for granular view if few logs, or Hour if many.
            const key = format(date, "HH:mm");

            if (!data[key]) {
                data[key] = { time: key, success: 0, error: 0 };
            }
            if (log.error) {
                data[key].error++;
            } else {
                data[key].success++;
            }
        });

        // Limit to last 20 data points for readability
        const values = Object.values(data);
        return values.slice(Math.max(values.length - 20, 0));
    }, [logs]);

    return (
        <div className="space-y-4 h-full flex flex-col">
            {/* Activity Chart */}
            <Card className="flex-none bg-background/50 backdrop-blur-sm">
                <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                        <Activity className="h-4 w-4" /> Activity Volume
                    </CardTitle>
                </CardHeader>
                <CardContent className="h-[150px]">
                    <ResponsiveContainer width="100%" height="100%">
                        <BarChart data={chartData}>
                            <CartesianGrid strokeDasharray="3 3" opacity={0.1} vertical={false} />
                            <XAxis
                                dataKey="time"
                                stroke="#888888"
                                fontSize={12}
                                tickLine={false}
                                axisLine={false}
                            />
                            <Tooltip
                                contentStyle={{ backgroundColor: '#1f2937', borderColor: '#374151', color: '#f3f4f6' }}
                                itemStyle={{ color: '#f3f4f6' }}
                                cursor={{ fill: 'transparent' }}
                            />
                            <Legend />
                            <Bar dataKey="success" name="Success" fill="#22c55e" radius={[4, 4, 0, 0]} stackId="a" />
                            <Bar dataKey="error" name="Error" fill="#ef4444" radius={[4, 4, 0, 0]} stackId="a" />
                        </BarChart>
                    </ResponsiveContainer>
                </CardContent>
            </Card>

            {/* Filters */}
            <Card className="flex-none">
                <CardContent className="p-4">
                    <div className="flex flex-col lg:flex-row gap-4 items-end">
                        <div className="grid gap-2 flex-1 w-full">
                            <label className="text-xs font-medium text-muted-foreground">Tool Name</label>
                            <Select value={toolName} onValueChange={setToolName}>
                                <SelectTrigger>
                                    <SelectValue placeholder="All Tools" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="ALL">All Tools</SelectItem>
                                    {availableTools.map(t => (
                                        <SelectItem key={t} value={t}>{t}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="grid gap-2 flex-1 w-full">
                            <label className="text-xs font-medium text-muted-foreground">User ID</label>
                            <Select value={userId} onValueChange={setUserId}>
                                <SelectTrigger>
                                    <SelectValue placeholder="All Users" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="ALL">All Users</SelectItem>
                                    {availableUsers.map(u => (
                                        <SelectItem key={u} value={u}>{u}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="grid gap-2 flex-1 w-full min-w-[200px]">
                            <label className="text-xs font-medium text-muted-foreground">Date Range</label>
                            <div className="flex gap-2">
                                <Popover>
                                    <PopoverTrigger asChild>
                                        <Button
                                            variant={"outline"}
                                            className={cn(
                                                "w-full justify-start text-left font-normal text-xs h-10",
                                                !startDate && "text-muted-foreground"
                                            )}
                                        >
                                            <CalendarIcon className="mr-2 h-3 w-3" />
                                            {startDate ? format(startDate, "P") : <span>Start</span>}
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
                                                "w-full justify-start text-left font-normal text-xs h-10",
                                                !endDate && "text-muted-foreground"
                                            )}
                                        >
                                            <CalendarIcon className="mr-2 h-3 w-3" />
                                            {endDate ? format(endDate, "P") : <span>End</span>}
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
                        <div className="flex gap-2">
                            <Button onClick={fetchLogs} disabled={loading} size="sm" className="h-10">
                                {loading ? <RefreshCw className="mr-2 h-4 w-4 animate-spin" /> : <Search className="mr-2 h-4 w-4" />}
                                Filter
                            </Button>
                            <Button variant="outline" onClick={handleSeed} disabled={loading} size="sm" className="h-10" title="Simulate traffic">
                                <Zap className="mr-2 h-4 w-4 text-yellow-500" />
                                Seed
                            </Button>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Results Table */}
            <Card className="flex-1 flex flex-col overflow-hidden border-muted/50 shadow-sm">
                <CardContent className="p-0 flex-1 overflow-auto">
                    <Table>
                        <TableHeader>
                            <TableRow className="bg-muted/20 hover:bg-muted/20">
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
                                    <TableCell colSpan={6} className="text-center h-32 text-muted-foreground">
                                        <div className="flex flex-col items-center gap-2">
                                            <Search className="h-8 w-8 opacity-20" />
                                            <p>No audit logs found matching your filters.</p>
                                            <Button variant="link" onClick={handleSeed} className="text-xs text-primary">
                                                Generate sample data?
                                            </Button>
                                        </div>
                                    </TableCell>
                                </TableRow>
                            )}
                            {logs.map((log, i) => (
                                <TableRow key={i} className="group hover:bg-muted/50 cursor-pointer transition-colors" onClick={() => setSelectedLog(log)}>
                                    <TableCell className="font-mono text-xs text-muted-foreground">
                                        {format(new Date(log.timestamp), "MMM dd HH:mm:ss")}
                                    </TableCell>
                                    <TableCell className="font-medium text-foreground">
                                        {log.toolName}
                                    </TableCell>
                                    <TableCell>
                                        <Badge variant="secondary" className="font-normal text-xs">
                                            {log.userId || "anon"}
                                        </Badge>
                                    </TableCell>
                                    <TableCell className="font-mono text-xs">
                                        {log.durationMs}ms
                                    </TableCell>
                                    <TableCell>
                                        {log.error ? (
                                            <Badge variant="destructive" className="gap-1 font-normal text-xs px-2 py-0.5">
                                                <AlertTriangle className="h-3 w-3" /> Error
                                            </Badge>
                                        ) : (
                                            <Badge variant="outline" className="text-green-500 border-green-500/30 bg-green-500/5 font-normal text-xs px-2 py-0.5">
                                                Success
                                            </Badge>
                                        )}
                                    </TableCell>
                                    <TableCell className="text-right">
                                        <Button variant="ghost" size="sm" className="opacity-0 group-hover:opacity-100 transition-opacity h-8 w-8 p-0">
                                            <Eye className="h-4 w-4" />
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>

            {/* Details Sheet */}
            <Sheet open={!!selectedLog} onOpenChange={(open) => !open && setSelectedLog(null)}>
                <SheetContent className="min-w-[50vw] sm:max-w-[600px] overflow-y-auto">
                    <SheetHeader className="mb-6 border-b pb-4">
                        <SheetTitle className="flex items-center gap-2 text-xl">
                            {selectedLog?.toolName}
                            {selectedLog?.error ? (
                                <Badge variant="destructive" className="ml-2">Failed</Badge>
                            ) : (
                                <Badge variant="outline" className="ml-2 text-green-500 border-green-500">Success</Badge>
                            )}
                        </SheetTitle>
                        <SheetDescription className="font-mono text-xs">
                            {selectedLog && format(new Date(selectedLog.timestamp), "PPP pp")} • {selectedLog?.duration}
                        </SheetDescription>
                    </SheetHeader>

                    {selectedLog && (
                        <div className="space-y-6">
                            <div className="grid grid-cols-2 gap-4 p-4 bg-muted/20 rounded-lg border">
                                <div>
                                    <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">User</span>
                                    <div className="mt-1 font-medium">{selectedLog.userId || "N/A"}</div>
                                </div>
                                <div>
                                    <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Profile</span>
                                    <div className="mt-1 font-medium">{selectedLog.profileId || "N/A"}</div>
                                </div>
                                <div>
                                    <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Latency</span>
                                    <div className="mt-1 font-mono">{selectedLog.durationMs}ms</div>
                                </div>
                                <div>
                                    <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Status</span>
                                    <div className={cn("mt-1 font-medium", selectedLog.error ? "text-red-500" : "text-green-500")}>
                                        {selectedLog.error ? "Error" : "OK"}
                                    </div>
                                </div>
                            </div>

                            {selectedLog.error && (
                                <div className="space-y-2">
                                    <h4 className="text-sm font-semibold flex items-center text-red-500">
                                        <AlertTriangle className="mr-2 h-4 w-4" />
                                        Error Details
                                    </h4>
                                    <div className="p-3 bg-red-950/20 border border-red-900/50 rounded-md text-sm font-mono text-red-400 break-words">
                                        {selectedLog.error}
                                    </div>
                                </div>
                            )}

                            <div className="space-y-2">
                                <h4 className="text-sm font-semibold">Arguments</h4>
                                <div className="rounded-md overflow-hidden border bg-zinc-950">
                                    <SyntaxHighlighter
                                        language="json"
                                        style={vscDarkPlus}
                                        customStyle={{ margin: 0, padding: '1rem', fontSize: '12px', background: 'transparent' }}
                                        wrapLongLines={true}
                                    >
                                        {formatJson(selectedLog.arguments) || "{}"}
                                    </SyntaxHighlighter>
                                </div>
                            </div>

                            <div className="space-y-2">
                                <h4 className="text-sm font-semibold">Result</h4>
                                <div className="rounded-md overflow-hidden border bg-zinc-950">
                                    <SyntaxHighlighter
                                        language="json"
                                        style={vscDarkPlus}
                                        customStyle={{ margin: 0, padding: '1rem', fontSize: '12px', background: 'transparent', maxHeight: '400px' }}
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
