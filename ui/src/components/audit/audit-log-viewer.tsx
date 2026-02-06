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
import { CalendarIcon, Search, RefreshCw, Eye, AlertTriangle, Play, RefreshCcw } from "lucide-react";
import { cn } from "@/lib/utils";
import Editor from "@monaco-editor/react";
import { useRouter } from "next/navigation";

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
    const [startDate, setStartDate] = useState<Date | undefined>(undefined);
    const [endDate, setEndDate] = useState<Date | undefined>(undefined);

    const router = useRouter();

    const fetchLogs = useCallback(async () => {
        setLoading(true);
        try {
            const filters: any = {
                limit: 50,
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

    const formatJson = (jsonStr: string) => {
        if (!jsonStr) return "";
        try {
            const obj = JSON.parse(jsonStr);
            return JSON.stringify(obj, null, 2);
        } catch (e) {
            return jsonStr;
        }
    };

    const handleReplay = (log: AuditLogEntry) => {
        let args = log.arguments;
        // Verify it's valid JSON to avoid breaking URL?
        // Actually, playground expects raw string usually if it's simple, but "arguments" is JSON string.
        // If it fails parsing, we send it as is.
        // But PlaygroundClientPro expects `tool` and `args` in query params.
        const params = new URLSearchParams();
        params.set("tool", log.toolName);
        params.set("args", args);
        router.push(`/playground?${params.toString()}`);
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
                <SheetContent className="sm:max-w-xl md:max-w-2xl w-full flex flex-col h-full p-0 gap-0">
                    <SheetHeader className="p-6 pb-4 border-b">
                        <div className="flex items-start justify-between">
                            <div className="space-y-1">
                                <SheetTitle className="flex items-center gap-2">
                                    {selectedLog?.toolName}
                                    {selectedLog?.error ? (
                                        <Badge variant="destructive" className="text-[10px] h-5 px-1.5">Failed</Badge>
                                    ) : (
                                        <Badge variant="outline" className="text-[10px] h-5 px-1.5 text-green-500 border-green-500/50">Success</Badge>
                                    )}
                                </SheetTitle>
                                <SheetDescription>
                                    Executed at {selectedLog && new Date(selectedLog.timestamp).toLocaleString()}
                                </SheetDescription>
                            </div>
                            {selectedLog && (
                                <Button size="sm" onClick={() => handleReplay(selectedLog)} className="gap-2">
                                    <RefreshCcw className="h-4 w-4" />
                                    Replay
                                </Button>
                            )}
                        </div>
                    </SheetHeader>

                    {selectedLog && (
                        <div className="flex-1 overflow-y-auto p-6 space-y-6">
                            <div className="grid grid-cols-2 gap-4 text-sm bg-muted/20 p-4 rounded-lg border">
                                <div>
                                    <span className="font-semibold block text-muted-foreground text-xs uppercase tracking-wider mb-1">User</span>
                                    <div className="font-mono text-xs">{selectedLog.userId || "N/A"}</div>
                                </div>
                                <div>
                                    <span className="font-semibold block text-muted-foreground text-xs uppercase tracking-wider mb-1">Profile</span>
                                    <div className="font-mono text-xs">{selectedLog.profileId || "N/A"}</div>
                                </div>
                                <div>
                                    <span className="font-semibold block text-muted-foreground text-xs uppercase tracking-wider mb-1">Duration</span>
                                    <div className="font-mono text-xs">{selectedLog.duration} ({selectedLog.durationMs}ms)</div>
                                </div>
                                <div>
                                    <span className="font-semibold block text-muted-foreground text-xs uppercase tracking-wider mb-1">Status</span>
                                    {selectedLog.error ? <span className="text-red-500 font-medium">Failed</span> : <span className="text-green-500 font-medium">Success</span>}
                                </div>
                            </div>

                            {selectedLog.error && (
                                <div className="bg-red-900/10 border border-red-900/30 rounded-md p-4 text-red-400 text-sm">
                                    <span className="font-semibold block mb-2 flex items-center gap-2">
                                        <AlertTriangle className="h-4 w-4" /> Error Details
                                    </span>
                                    <pre className="whitespace-pre-wrap font-mono text-xs">{selectedLog.error}</pre>
                                </div>
                            )}

                            <div className="space-y-2">
                                <h4 className="text-sm font-medium flex items-center gap-2">
                                    Arguments
                                </h4>
                                <div className="h-[200px] border rounded-md overflow-hidden bg-[#1e1e1e]">
                                    <Editor
                                        height="100%"
                                        defaultLanguage="json"
                                        value={formatJson(selectedLog.arguments)}
                                        theme="vs-dark"
                                        options={{
                                            readOnly: true,
                                            minimap: { enabled: false },
                                            fontSize: 12,
                                            scrollBeyondLastLine: false,
                                            folding: true,
                                            lineNumbers: "off",
                                            renderLineHighlight: "none"
                                        }}
                                    />
                                </div>
                            </div>

                            <div className="space-y-2">
                                <h4 className="text-sm font-medium flex items-center gap-2">
                                    Result
                                </h4>
                                <div className="h-[300px] border rounded-md overflow-hidden bg-[#1e1e1e]">
                                     <Editor
                                        height="100%"
                                        defaultLanguage="json"
                                        value={formatJson(selectedLog.result)}
                                        theme="vs-dark"
                                        options={{
                                            readOnly: true,
                                            minimap: { enabled: false },
                                            fontSize: 12,
                                            scrollBeyondLastLine: false,
                                            folding: true,
                                            lineNumbers: "off",
                                            renderLineHighlight: "none"
                                        }}
                                    />
                                </div>
                            </div>
                        </div>
                    )}
                </SheetContent>
            </Sheet>
        </div>
    );
}
