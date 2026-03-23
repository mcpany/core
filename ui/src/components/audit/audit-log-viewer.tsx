/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect, useCallback, useMemo } from "react";
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
import { CalendarIcon, Search, RefreshCw, Eye, AlertTriangle, Download, Terminal, Code } from "lucide-react";
import { cn } from "@/lib/utils";
import { useToast } from "@/hooks/use-toast";
import { RichResultViewer } from "@/components/tools/rich-result-viewer";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";

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
    const [exporting, setExporting] = useState(false);
    const [selectedLog, setSelectedLog] = useState<AuditLogEntry | null>(null);
    const { toast } = useToast();

    // Filters
    const [toolName, setToolName] = useState("");
    const [userId, setUserId] = useState("");
    const [startDate, setStartDate] = useState<Date | undefined>(undefined);
    const [endDate, setEndDate] = useState<Date | undefined>(undefined);

    const fetchLogs = useCallback(async () => {
        setLoading(true);
        try {
            const filters: Record<string, any> = {
                limit: 50,
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

    const handleExport = async () => {
        setExporting(true);
        try {
            const filters: Record<string, any> = {};
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

    const parseJsonSafely = useCallback((jsonStr: string) => {
        if (!jsonStr) return null;
        try {
            return JSON.parse(jsonStr);
        } catch (_e) {
            return jsonStr;
        }
    }, []);

    const selectedLogArgs = useMemo(() => {
        return selectedLog ? parseJsonSafely(selectedLog.arguments) : null;
    }, [selectedLog, parseJsonSafely]);

    const selectedLogResult = useMemo(() => {
        return selectedLog ? parseJsonSafely(selectedLog.result) : null;
    }, [selectedLog, parseJsonSafely]);

    return (
        <div className="space-y-4 h-full flex flex-col">
            <Card className="flex-none backdrop-blur-sm bg-background/50 border-muted">
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
                                onKeyDown={(e) => e.key === 'Enter' && fetchLogs()}
                            />
                        </div>
                        <div className="grid gap-2 flex-1 w-full md:w-auto">
                            <label className="text-sm font-medium">User ID</label>
                            <Input
                                placeholder="e.g. alice"
                                value={userId}
                                onChange={(e) => setUserId(e.target.value)}
                                onKeyDown={(e) => e.key === 'Enter' && fetchLogs()}
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
                            <Button variant="outline" onClick={handleExport} disabled={exporting}>
                                {exporting ? <RefreshCw className="mr-2 h-4 w-4 animate-spin" /> : <Download className="mr-2 h-4 w-4" />}
                                Export CSV
                            </Button>
                            <Button onClick={fetchLogs} disabled={loading}>
                                {loading ? <RefreshCw className="mr-2 h-4 w-4 animate-spin" /> : <Search className="mr-2 h-4 w-4" />}
                                Filter
                            </Button>
                        </div>
                    </div>
                </CardContent>
            </Card>

            <Card className="flex-1 flex flex-col overflow-hidden backdrop-blur-sm bg-background/50 border-muted">
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
                            {loading ? (
                                Array.from({ length: 5 }).map((_, i) => (
                                    <TableRow key={i}>
                                        <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                                        <TableCell><Skeleton className="h-4 w-24" /></TableCell>
                                        <TableCell><Skeleton className="h-4 w-16" /></TableCell>
                                        <TableCell><Skeleton className="h-4 w-12" /></TableCell>
                                        <TableCell><Skeleton className="h-5 w-16 rounded-full" /></TableCell>
                                        <TableCell className="text-right"><Skeleton className="h-8 w-16 ml-auto" /></TableCell>
                                    </TableRow>
                                ))
                            ) : logs.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={6} className="text-center h-24 text-muted-foreground">
                                        No logs found.
                                    </TableCell>
                                </TableRow>
                            ) : (
                                logs.map((log, i) => (
                                    <TableRow key={i} className="hover:bg-muted/50 transition-colors">
                                        <TableCell className="font-mono text-xs">
                                            {new Date(log.timestamp).toLocaleString()}
                                        </TableCell>
                                        <TableCell className="font-medium">{log.toolName}</TableCell>
                                        <TableCell>{log.userId || "-"}</TableCell>
                                        <TableCell>{log.duration}</TableCell>
                                        <TableCell>
                                            {log.error ? (
                                                <Badge variant="destructive" className="gap-1 shadow-sm">
                                                    <AlertTriangle className="h-3 w-3" /> Error
                                                </Badge>
                                            ) : (
                                                <Badge variant="outline" className="text-green-600 border-green-500/30 bg-green-500/10 shadow-sm">
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
                                ))
                            )}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>

            <Dialog open={!!selectedLog} onOpenChange={(open) => !open && setSelectedLog(null)}>
                <DialogContent className="max-w-4xl max-h-[85vh] overflow-y-hidden flex flex-col p-0 border-muted bg-background/95 backdrop-blur-xl shadow-2xl">
                    <div className="p-6 pb-2 border-b">
                        <DialogHeader>
                            <DialogTitle className="text-xl flex items-center gap-2">
                                Audit Log Detail
                            </DialogTitle>
                            <DialogDescription>
                                Execution details for <span className="font-semibold text-foreground">{selectedLog?.toolName}</span> at {selectedLog && new Date(selectedLog.timestamp).toLocaleString()}
                            </DialogDescription>
                        </DialogHeader>
                        {selectedLog && (
                            <div className="grid grid-cols-4 gap-4 text-sm mt-4 bg-muted/30 p-4 rounded-lg border border-border/50">
                                <div>
                                    <span className="font-semibold block text-muted-foreground text-xs uppercase tracking-wider mb-1">User ID</span>
                                    {selectedLog.userId || "N/A"}
                                </div>
                                <div>
                                    <span className="font-semibold block text-muted-foreground text-xs uppercase tracking-wider mb-1">Profile ID</span>
                                    {selectedLog.profileId || "N/A"}
                                </div>
                                <div>
                                    <span className="font-semibold block text-muted-foreground text-xs uppercase tracking-wider mb-1">Duration</span>
                                    {selectedLog.duration} ({selectedLog.durationMs}ms)
                                </div>
                                <div>
                                    <span className="font-semibold block text-muted-foreground text-xs uppercase tracking-wider mb-1">Status</span>
                                    {selectedLog.error ? <span className="text-destructive font-medium flex items-center gap-1"><AlertTriangle className="h-3 w-3" /> Failed</span> : <span className="text-green-600 font-medium">Success</span>}
                                </div>
                            </div>
                        )}
                    </div>

                    <ScrollArea className="flex-1 p-6">
                        {selectedLog && (
                            <div className="space-y-6">
                                {selectedLog.error && (
                                    <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4 text-destructive text-sm shadow-sm">
                                        <span className="font-semibold flex items-center gap-2 mb-2"><AlertTriangle className="h-4 w-4" /> Error Details</span>
                                        <pre className="whitespace-pre-wrap font-mono text-xs">{selectedLog.error}</pre>
                                    </div>
                                )}

                                <div>
                                    <h4 className="text-sm font-semibold mb-3 flex items-center gap-2 text-primary">
                                        <Code className="h-4 w-4" /> Arguments
                                    </h4>
                                    <div className="rounded-lg overflow-hidden border bg-card shadow-sm">
                                        {selectedLogArgs ? (
                                            <RichResultViewer result={selectedLogArgs} />
                                        ) : (
                                            <div className="p-4 text-sm text-muted-foreground">No arguments provided.</div>
                                        )}
                                    </div>
                                </div>

                                <div>
                                    <h4 className="text-sm font-semibold mb-3 flex items-center gap-2 text-primary">
                                        <Terminal className="h-4 w-4" /> Result
                                    </h4>
                                    <div className="rounded-lg overflow-hidden border bg-card shadow-sm">
                                        {selectedLogResult !== null ? (
                                            <RichResultViewer result={selectedLogResult} />
                                        ) : selectedLog.error ? (
                                            <div className="p-4 text-sm text-muted-foreground italic">Execution failed, no result.</div>
                                        ) : (
                                            <div className="p-4 text-sm text-muted-foreground italic">No result data.</div>
                                        )}
                                    </div>
                                </div>
                            </div>
                        )}
                    </ScrollArea>
                </DialogContent>
            </Dialog>
        </div>
    );
}
