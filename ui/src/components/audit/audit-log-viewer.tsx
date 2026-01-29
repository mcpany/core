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
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { format } from "date-fns";
import { CalendarIcon, Search, RefreshCw, Eye, AlertTriangle } from "lucide-react";
import { cn } from "@/lib/utils";
import { AuditLogDetail, AuditLogEntry } from "./audit-log-detail";

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

            <Dialog open={!!selectedLog} onOpenChange={(open) => !open && setSelectedLog(null)}>
                <DialogContent className="max-w-4xl h-[80vh] flex flex-col p-0 overflow-hidden gap-0">
                    <DialogHeader className="sr-only">
                        <DialogTitle>Audit Log Detail</DialogTitle>
                    </DialogHeader>
                    {selectedLog && <AuditLogDetail entry={selectedLog} />}
                </DialogContent>
            </Dialog>
        </div>
    );
}
