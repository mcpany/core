/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import { useState, useEffect, useCallback } from "react";
import { apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
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
import { format } from "date-fns";
import { CalendarIcon, Search, RefreshCw, Eye, AlertTriangle, Download } from "lucide-react";
import { cn } from "@/lib/utils";
import { useToast } from "@/hooks/use-toast";
import SyntaxHighlighter from 'react-syntax-highlighter/dist/esm/light';
import json from 'react-syntax-highlighter/dist/esm/languages/hljs/json';
import vs2015 from 'react-syntax-highlighter/dist/esm/styles/hljs/vs2015';
import { RichResultViewer } from "@/components/tools/rich-result-viewer";

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
 * Summary: AuditLogViewer component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
export function AuditLogViewer() {
    SyntaxHighlighter.registerLanguage('json', json);
    const [logs, setLogs] = useState<AuditLogEntry[]>([]);
    const [loading, setLoading] = useState(true);
    const [exporting, setExporting] = useState(false);
    const [selectedLog, setSelectedLog] = useState<AuditLogEntry | null>(null);
    const { toast } = useToast();

    // Filters
    const [toolName, setToolName] = useState("");
    const [userId, setUserId] = useState("");
    const [startDate, setStartDate] = useState<Date | undefined>(undefined);
    const [endDate, setEndDate] = useState<Date | undefined>(undefined);

    // Pagination
    const [page, setPage] = useState(0);
    const limit = 50;
    const [hasMore, setHasMore] = useState(false);

    const fetchLogs = useCallback(async () => {
        setLoading(true);
        try {
            const filters: any = {
                limit: limit,
                offset: page * limit
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
            const newLogs = res.entries || [];
            setLogs(newLogs);
            setHasMore(newLogs.length === limit);
        } catch (e) {
            console.error("Failed to fetch audit logs", e);
        } finally {
            setLoading(false);
        }
    }, [toolName, userId, startDate, endDate, page]);

    useEffect(() => {
        fetchLogs();
    }, [fetchLogs]);

    const handleFilter = () => {
        if (page === 0) {
            fetchLogs();
        } else {
            setPage(0);
        }
    };

    const handleExport = async () => {
        setExporting(true);
        try {
            const filters: any = {};
            if (toolName) filters.tool_name = toolName;
            if (userId) filters.user_id = userId;
            if (startDate) filters.start_time = startDate.toISOString();
            if (endDate) filters.end_time = endDate.toISOString();

            await apiClient.exportAuditLogs(filters);
            toast({
                title: "Export Successful",
                description: "Audit logs have been exported.",
            });
        } catch (e: any) {
            console.error("Failed to export audit logs", e);
            toast({
                title: "Export Failed",
                description: e.message || "Failed to export audit logs.",
                variant: "destructive",
            });
        } finally {
            setExporting(false);
        }
    };


    const safeParse = (str: string | undefined | null) => {
        if (!str) return null;
        try {
            return JSON.parse(str);
        } catch (e) {
            return str;
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
                        <div className="flex gap-2 w-full md:w-auto mt-4 md:mt-0">
                            <Button variant="outline" onClick={handleExport} disabled={exporting || loading}>
                                {exporting ? <RefreshCw className="mr-2 h-4 w-4 animate-spin" /> : <Download className="mr-2 h-4 w-4" />}
                                Export CSV
                            </Button>
                            <Button onClick={handleFilter} disabled={loading}>
                                {loading ? <RefreshCw className="mr-2 h-4 w-4 animate-spin" /> : <Search className="mr-2 h-4 w-4" />}
                                Filter
                            </Button>
                        </div>
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
                <div className="flex items-center justify-between px-4 py-3 border-t bg-muted/10">
                    <div className="text-sm text-muted-foreground">
                        Showing {logs.length > 0 ? page * limit + 1 : 0} to {page * limit + logs.length} entries
                    </div>
                    <div className="flex items-center space-x-2">
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setPage(p => Math.max(0, p - 1))}
                            disabled={page === 0 || loading}
                        >
                            Previous
                        </Button>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setPage(p => p + 1)}
                            disabled={!hasMore || loading}
                        >
                            Next
                        </Button>
                    </div>
                </div>
            </Card>

            <Dialog open={!!selectedLog} onOpenChange={(open) => !open && setSelectedLog(null)}>
                <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
                    <DialogHeader>
                        <DialogTitle>Audit Log Detail</DialogTitle>
                        <DialogDescription>
                            Execution details for {selectedLog?.toolName} at {selectedLog && new Date(selectedLog.timestamp).toLocaleString()}
                        </DialogDescription>
                    </DialogHeader>
                    {selectedLog && (
                        <div className="space-y-4">
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
                                    <RichResultViewer result={safeParse(selectedLog.arguments) || {}} />
                                </div>
                            </div>

                            <div>
                                <h4 className="text-sm font-medium mb-2">Result</h4>
                                <div className="rounded-md overflow-hidden border">
                                    <RichResultViewer result={safeParse(selectedLog.result) || (selectedLog.error ? null : {})} />
                                </div>
                            </div>
                        </div>
                    )}
                </DialogContent>
            </Dialog>
        </div>
    );
}
