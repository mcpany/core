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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { format } from "date-fns";
import { CalendarIcon, Search, RefreshCw, Eye, AlertTriangle, Zap } from "lucide-react";
import { cn } from "@/lib/utils";
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { toast } from "sonner";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

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
    const [toolName, setToolName] = useState("");
    const [userId, setUserId] = useState("");
    const [status, setStatus] = useState<"all" | "success" | "error">("all");
    const [startDate, setStartDate] = useState<Date | undefined>(undefined);
    const [endDate, setEndDate] = useState<Date | undefined>(undefined);

    const fetchLogs = useCallback(async () => {
        setLoading(true);
        try {
            const filters: any = {
                limit: 100, // Increased limit for client-side filtering
                offset: 0
            };
            if (toolName) filters.tool_name = toolName;
            if (userId) filters.user_id = userId;
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
            let entries = res.entries || [];

            // Client-side status filtering (until API supports it)
            if (status === "success") {
                entries = entries.filter((e: AuditLogEntry) => !e.error);
            } else if (status === "error") {
                entries = entries.filter((e: AuditLogEntry) => !!e.error);
            }

            setLogs(entries);
        } catch (e) {
            console.error("Failed to fetch audit logs", e);
        } finally {
            setLoading(false);
        }
    }, [toolName, userId, status, startDate, endDate]);

    useEffect(() => {
        fetchLogs();
    }, [fetchLogs]);

    const handleSeed = async () => {
        setLoading(true);
        try {
            await apiClient.seedAuditLogs(50);
            toast.success("Seeded 50 random audit logs");
            fetchLogs();
        } catch (e) {
            console.error("Failed to seed logs", e);
            toast.error("Failed to seed logs");
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

    const chartData = useMemo(() => {
        const counts: Record<string, number> = {};
        logs.forEach(log => {
            // Group by hour
            const date = new Date(log.timestamp);
            const key = format(date, "HH:00");
            counts[key] = (counts[key] || 0) + 1;
        });

        // Ensure chronological order (simplified for last 24h view)
        return Object.entries(counts)
            .map(([time, count]) => ({ time, count }))
            .sort((a, b) => a.time.localeCompare(b.time));
    }, [logs]);

    return (
        <div className="space-y-4 h-full flex flex-col">
            <div className="grid gap-4 md:grid-cols-3">
                <Card className="md:col-span-3">
                    <CardHeader className="pb-2">
                        <CardTitle>Activity Volume</CardTitle>
                        <CardDescription>Requests per hour over the selected period.</CardDescription>
                    </CardHeader>
                    <CardContent className="h-[150px]">
                        <ResponsiveContainer width="100%" height="100%">
                            <BarChart data={chartData}>
                                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#333" />
                                <XAxis dataKey="time" stroke="#888888" fontSize={12} tickLine={false} axisLine={false} />
                                <YAxis stroke="#888888" fontSize={12} tickLine={false} axisLine={false} tickFormatter={(value) => `${value}`} />
                                <Tooltip
                                    contentStyle={{ backgroundColor: "#1f2937", border: "none", borderRadius: "8px", color: "#f3f4f6" }}
                                    cursor={{ fill: "#374151", opacity: 0.4 }}
                                />
                                <Bar dataKey="count" fill="#3b82f6" radius={[4, 4, 0, 0]} />
                            </BarChart>
                        </ResponsiveContainer>
                    </CardContent>
                </Card>
            </div>

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
                            <label className="text-sm font-medium">Status</label>
                            <Select value={status} onValueChange={(v: any) => setStatus(v)}>
                                <SelectTrigger>
                                    <SelectValue placeholder="Status" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">All</SelectItem>
                                    <SelectItem value="success">Success</SelectItem>
                                    <SelectItem value="error">Error</SelectItem>
                                </SelectContent>
                            </Select>
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
                        <Button variant="outline" onClick={handleSeed} disabled={loading} className="ml-auto">
                            <Zap className="mr-2 h-4 w-4 text-yellow-500" />
                            Simulate Traffic
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
                <SheetContent className="sm:max-w-2xl w-full overflow-y-auto">
                    <SheetHeader>
                        <SheetTitle>Audit Log Detail</SheetTitle>
                        <SheetDescription>
                            Execution details for {selectedLog?.toolName} at {selectedLog && new Date(selectedLog.timestamp).toLocaleString()}
                        </SheetDescription>
                    </SheetHeader>
                    {selectedLog && (
                        <div className="space-y-6 mt-6">
                            <div className="grid grid-cols-2 gap-4 text-sm border-b pb-4">
                                <div>
                                    <span className="font-semibold block text-muted-foreground text-xs uppercase tracking-wider">User ID</span>
                                    <div className="mt-1 font-mono">{selectedLog.userId || "N/A"}</div>
                                </div>
                                <div>
                                    <span className="font-semibold block text-muted-foreground text-xs uppercase tracking-wider">Profile ID</span>
                                    <div className="mt-1 font-mono">{selectedLog.profileId || "N/A"}</div>
                                </div>
                                <div>
                                    <span className="font-semibold block text-muted-foreground text-xs uppercase tracking-wider">Duration</span>
                                    <div className="mt-1 font-mono">{selectedLog.duration} ({selectedLog.durationMs}ms)</div>
                                </div>
                                <div>
                                    <span className="font-semibold block text-muted-foreground text-xs uppercase tracking-wider">Status</span>
                                    <div className="mt-1">
                                        {selectedLog.error ?
                                            <Badge variant="destructive">Failed</Badge> :
                                            <Badge variant="outline" className="text-green-500 border-green-500/50">Success</Badge>
                                        }
                                    </div>
                                </div>
                            </div>

                            {selectedLog.error && (
                                <div className="bg-red-900/20 border border-red-900/50 rounded-md p-4 text-red-200 text-sm">
                                    <span className="font-semibold block mb-2 text-red-100 flex items-center gap-2">
                                        <AlertTriangle className="h-4 w-4" /> Error Details
                                    </span>
                                    <code className="whitespace-pre-wrap">{selectedLog.error}</code>
                                </div>
                            )}

                            <div>
                                <h4 className="text-sm font-semibold mb-2 flex items-center gap-2">
                                    Arguments
                                </h4>
                                <div className="rounded-md overflow-hidden border bg-[#1e1e1e]">
                                    <SyntaxHighlighter
                                        language="json"
                                        style={vscDarkPlus}
                                        customStyle={{ margin: 0, fontSize: '12px', padding: '1rem' }}
                                    >
                                        {formatJson(selectedLog.arguments) || "{}"}
                                    </SyntaxHighlighter>
                                </div>
                            </div>

                            <div>
                                <h4 className="text-sm font-semibold mb-2 flex items-center gap-2">
                                    Result
                                </h4>
                                <div className="rounded-md overflow-hidden border bg-[#1e1e1e]">
                                    <SyntaxHighlighter
                                        language="json"
                                        style={vscDarkPlus}
                                        customStyle={{ margin: 0, fontSize: '12px', padding: '1rem', maxHeight: '400px' }}
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
