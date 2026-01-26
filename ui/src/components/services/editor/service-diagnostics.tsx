/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { UpstreamServiceConfig, apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Loader2, CheckCircle2, XCircle, AlertTriangle } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";

interface ServiceDiagnosticsProps {
    service: UpstreamServiceConfig;
}

type CheckStatus = "idle" | "running" | "success" | "error" | "warning";

interface DiagnosticResult {
    name: string;
    status: CheckStatus;
    message?: string;
    details?: string;
}

/**
 * A component that runs and displays diagnostic checks for a service.
 * Checks include configuration validation, runtime status, and tool discovery.
 *
 * @param props - The component props.
 * @param props.service - The service configuration to diagnose.
 * @returns The rendered diagnostic component.
 */
export function ServiceDiagnostics({ service }: ServiceDiagnosticsProps) {
    const [running, setRunning] = useState(false);
    const [results, setResults] = useState<DiagnosticResult[]>([]);

    const runDiagnostics = async () => {
        setRunning(true);
        setResults([]);

        const newResults: DiagnosticResult[] = [];

        // 1. Configuration Validation
        const configCheck: DiagnosticResult = { name: "Configuration Validation", status: "running" };
        setResults([configCheck]);

        try {
            const validation = await apiClient.validateService(service);
            if (validation.valid) {
                configCheck.status = "success";
                configCheck.message = "Configuration is valid.";
            } else {
                configCheck.status = "error";
                configCheck.message = "Configuration is invalid.";
                configCheck.details = validation.errors?.join("\n");
            }
        } catch (e: any) {
            configCheck.status = "error";
            configCheck.message = "Validation request failed.";
            configCheck.details = e.message;
        }
        newResults.push(configCheck);
        setResults([...newResults]);

        // Only proceed if saved (has ID/Name)
        if (service.name && service.id) {
            // 2. Runtime Status
            const statusCheck: DiagnosticResult = { name: "Runtime Status", status: "running" };
            setResults([...newResults, statusCheck]);

            try {
                const status = await apiClient.getServiceStatus(service.name);
                // Assume status returns something like { status: "Active", ... }
                // Adjust based on actual API response if needed.
                if (status.status === "Active" || status.status === "Running" || status.status === "OK") {
                    statusCheck.status = "success";
                    statusCheck.message = `Service is ${status.status}.`;
                } else {
                    statusCheck.status = "warning";
                    statusCheck.message = `Service status is ${status.status}.`;
                }

                if (status.lastError) {
                    statusCheck.status = "error";
                    statusCheck.details = status.lastError;
                }
            } catch (e: any) {
                statusCheck.status = "error";
                statusCheck.message = "Failed to fetch service status.";
                statusCheck.details = e.message;
            }
            newResults.push(statusCheck);
            setResults([...newResults]);

            // 3. Tool Discovery
            const toolCheck: DiagnosticResult = { name: "Tool Discovery", status: "running" };
            setResults([...newResults, toolCheck]);

            try {
                const toolsResponse = await apiClient.listTools();
                // Filter tools by serviceId (which might be service name or ID)
                const tools = toolsResponse.tools.filter((t: any) => t.serviceId === service.name || t.serviceId === service.id);

                if (tools.length > 0) {
                    toolCheck.status = "success";
                    toolCheck.message = `Discovered ${tools.length} tool(s).`;
                    toolCheck.details = tools.map((t: any) => `- ${t.name}`).join("\n");
                } else {
                    toolCheck.status = "warning";
                    toolCheck.message = "No tools discovered.";
                    toolCheck.details = "The service might be running but not exposing any tools, or discovery failed.";
                }
            } catch (e: any) {
                toolCheck.status = "error";
                toolCheck.message = "Failed to list tools.";
                toolCheck.details = e.message;
            }
            newResults.push(toolCheck);
            setResults([...newResults]);
        } else {
             // Not saved yet
             newResults.push({
                 name: "Runtime Checks",
                 status: "warning",
                 message: "Skipped (Service not saved yet).",
                 details: "Save the service to run runtime diagnostics."
             });
             setResults([...newResults]);
        }

        setRunning(false);
    };

    const renderIcon = (status: CheckStatus) => {
        switch (status) {
            case "running": return <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />;
            case "success": return <CheckCircle2 className="h-5 w-5 text-green-500" />;
            case "error": return <XCircle className="h-5 w-5 text-destructive" />;
            case "warning": return <AlertTriangle className="h-5 w-5 text-yellow-500" />;
            default: return <div className="h-5 w-5" />;
        }
    };

    return (
        <div className="space-y-4">
            <Card>
                <CardHeader>
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle>Service Diagnostics</CardTitle>
                            <CardDescription>
                                Run checks to verify configuration, connectivity, and capability discovery.
                            </CardDescription>
                        </div>
                        <Button onClick={runDiagnostics} disabled={running}>
                            {running && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Run Diagnostics
                        </Button>
                    </div>
                </CardHeader>
                <CardContent>
                    <div className="space-y-4">
                        {results.length === 0 && !running && (
                            <div className="text-center text-muted-foreground py-8">
                                Click "Run Diagnostics" to start checking the service.
                            </div>
                        )}

                        {results.map((result, index) => (
                            <div key={index} className="border rounded-lg p-4 space-y-2">
                                <div className="flex items-center gap-3">
                                    {renderIcon(result.status)}
                                    <h3 className="font-medium text-sm">{result.name}</h3>
                                    <div className="flex-1" />
                                    <span className="text-xs text-muted-foreground capitalize">{result.status}</span>
                                </div>
                                {result.message && (
                                    <p className="text-sm text-muted-foreground ml-8">{result.message}</p>
                                )}
                                {result.details && (
                                    <ScrollArea className="h-24 w-full rounded-md border bg-muted/50 p-2 mt-2 ml-8 text-xs font-mono">
                                        <pre className="whitespace-pre-wrap break-all">{result.details}</pre>
                                    </ScrollArea>
                                )}
                            </div>
                        ))}
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
