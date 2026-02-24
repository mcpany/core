/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { RefreshCw, Scan, Radio } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

interface ProviderStatus {
    name: string;
    status: string;
    lastError: string;
    lastRunAt: string;
    discoveredCount: number;
}

/**
 * DiscoveryStatus displays the status of auto-discovery providers.
 */
export function DiscoveryStatus() {
    const [statuses, setStatuses] = useState<ProviderStatus[]>([]);
    const [loading, setLoading] = useState(true);
    const [scanning, setScanning] = useState(false);
    const { toast } = useToast();

    const fetchStatus = async () => {
        try {
            const data = await apiClient.getDiscoveryStatus();
            setStatuses(data);
        } catch (e) {
            console.error("Failed to fetch discovery status", e);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchStatus();
        const interval = setInterval(fetchStatus, 5000);
        return () => clearInterval(interval);
    }, []);

    const handleScan = async () => {
        setScanning(true);
        try {
            await apiClient.triggerDiscovery();
            toast({
                title: "Scan Started",
                description: "Auto-discovery process has been triggered in the background.",
            });
            fetchStatus();
        } catch (e) {
            toast({
                title: "Scan Failed",
                description: "Failed to trigger discovery process.",
                variant: "destructive",
            });
        } finally {
            // Keep spinning for a bit to show activity
            setTimeout(() => setScanning(false), 1000);
        }
    };

    return (
        <Card className="border-l-4 border-l-blue-500 shadow-sm">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-xl font-bold flex items-center gap-2">
                    <Radio className="h-5 w-5 text-blue-500" />
                    Auto-Discovery Status
                </CardTitle>
                <Button onClick={handleScan} disabled={scanning} size="sm" variant="outline">
                    {scanning ? (
                        <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                    ) : (
                        <Scan className="mr-2 h-4 w-4" />
                    )}
                    Scan Network
                </Button>
            </CardHeader>
            <CardContent>
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Provider</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead>Discovered</TableHead>
                            <TableHead>Last Run</TableHead>
                            <TableHead>Message</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {loading && statuses.length === 0 && (
                            <TableRow>
                                <TableCell colSpan={5} className="text-center text-muted-foreground">Loading status...</TableCell>
                            </TableRow>
                        )}
                        {!loading && statuses.length === 0 && (
                            <TableRow>
                                <TableCell colSpan={5} className="text-center text-muted-foreground">No discovery providers configured.</TableCell>
                            </TableRow>
                        )}
                        {statuses.map((status) => (
                            <TableRow key={status.name}>
                                <TableCell className="font-medium">{status.name}</TableCell>
                                <TableCell>
                                    <Badge
                                        variant={status.status === "OK" ? "default" : status.status === "ERROR" ? "destructive" : "outline"}
                                        className={status.status === "OK" ? "bg-green-600" : ""}
                                    >
                                        {status.status}
                                    </Badge>
                                </TableCell>
                                <TableCell>{status.discoveredCount}</TableCell>
                                <TableCell className="text-xs text-muted-foreground">
                                    {status.lastRunAt ? new Date(status.lastRunAt).toLocaleString() : "Never"}
                                </TableCell>
                                <TableCell className="max-w-[200px] truncate text-xs text-red-500" title={status.lastError}>
                                    {status.lastError}
                                </TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </CardContent>
        </Card>
    );
}
