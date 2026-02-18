/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo, useEffect } from "react";
import { HttpCallDefinition, HttpCallDefinition_HttpMethod, ParameterType } from "@proto/config/v1/call";
import { ToolDefinition } from "@proto/config/v1/tool";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { Play, Loader2, Save } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { JsonView } from "@/components/ui/json-view";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

interface RequestPreviewProps {
    call: HttpCallDefinition;
    tool: ToolDefinition;
    serviceName?: string;
    onExecute?: (args: any) => void;
    executionResult?: any;
    isExecuting?: boolean;
}

const getMethodName = (method: HttpCallDefinition_HttpMethod) => {
    switch (method) {
        case HttpCallDefinition_HttpMethod.HTTP_METHOD_GET: return "GET";
        case HttpCallDefinition_HttpMethod.HTTP_METHOD_POST: return "POST";
        case HttpCallDefinition_HttpMethod.HTTP_METHOD_PUT: return "PUT";
        case HttpCallDefinition_HttpMethod.HTTP_METHOD_DELETE: return "DELETE";
        case HttpCallDefinition_HttpMethod.HTTP_METHOD_PATCH: return "PATCH";
        default: return "GET";
    }
};

const getMethodColor = (method: HttpCallDefinition_HttpMethod) => {
    switch (method) {
        case HttpCallDefinition_HttpMethod.HTTP_METHOD_GET: return "text-blue-500 border-blue-500/20 bg-blue-500/10";
        case HttpCallDefinition_HttpMethod.HTTP_METHOD_POST: return "text-green-500 border-green-500/20 bg-green-500/10";
        case HttpCallDefinition_HttpMethod.HTTP_METHOD_PUT: return "text-orange-500 border-orange-500/20 bg-orange-500/10";
        case HttpCallDefinition_HttpMethod.HTTP_METHOD_DELETE: return "text-red-500 border-red-500/20 bg-red-500/10";
        default: return "text-gray-500 border-gray-500/20 bg-gray-500/10";
    }
};

export function RequestPreview({ call, tool, serviceName, onExecute, executionResult, isExecuting }: RequestPreviewProps) {
    const [argsJson, setArgsJson] = useState("{}");
    const [argsError, setArgsError] = useState<string | null>(null);

    // Initialize args with defaults if empty
    useEffect(() => {
        if (argsJson === "{}" && call.parameters && call.parameters.length > 0) {
            const defaults: Record<string, any> = {};
            call.parameters.forEach(p => {
                if (p.schema?.name) {
                    if (p.schema.type === ParameterType.STRING) defaults[p.schema.name] = "value";
                    else if (p.schema.type === ParameterType.NUMBER || p.schema.type === ParameterType.INTEGER) defaults[p.schema.name] = 0;
                    else if (p.schema.type === ParameterType.BOOLEAN) defaults[p.schema.name] = false;
                }
            });
            if (Object.keys(defaults).length > 0) {
                setArgsJson(JSON.stringify(defaults, null, 2));
            }
        }
    }, [call.parameters]);

    const preview = useMemo(() => {
        let args: any = {};
        try {
            args = JSON.parse(argsJson);
            setArgsError(null);
        } catch (e) {
            // Don't update error state immediately while typing partial JSON?
            // Actually, for preview, we need valid JSON.
            // We'll just return null or previous valid state?
            // Let's show error in UI but try to proceed if possible or just stop.
            setArgsError("Invalid JSON");
            return null;
        }

        // 1. Path Substitution
        let path = call.endpointPath || "/";
        path = path.replace(/{(\w+)}/g, (_, key) => {
            return args[key] !== undefined ? String(args[key]) : `{${key}}`;
        });

        // 2. Query Params (Simulation)
        // Note: The actual mapping logic is in backend. This is a simplified simulation.
        // Assuming unmapped args go to query for GET, body for POST?
        // Or strictly following `call.parameters`?
        // For now, we simulate that path params are consumed, rest might be query or body.

        // This is a visual aid, so accurate mapping requires duplicating backend logic.
        // For "MCP Any", the HTTP adapter uses Templates or direct mapping.
        // If no explicit mapping, it might dump args to body or query.

        return {
            method: getMethodName(call.method),
            url: path,
            body: args, // Simplified: Assume body receives all args for non-GET
        };
    }, [call, argsJson]);

    const handleExecute = () => {
        if (!onExecute) return;
        try {
            const args = JSON.parse(argsJson);
            onExecute(args);
        } catch {
            setArgsError("Invalid JSON");
        }
    };

    return (
        <div className="flex flex-col gap-4 h-full">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 h-full">
                <div className="flex flex-col gap-4">
                    <div className="space-y-2">
                        <Label>Test Arguments (JSON)</Label>
                        <Textarea
                            className="font-mono text-xs h-[200px]"
                            value={argsJson}
                            onChange={(e) => setArgsJson(e.target.value)}
                        />
                        {argsError && <p className="text-xs text-red-500">{argsError}</p>}
                    </div>

                    <Card>
                        <CardHeader className="py-3 px-4 bg-muted/20 border-b">
                            <CardTitle className="text-xs font-medium uppercase tracking-wider text-muted-foreground">Request Preview</CardTitle>
                        </CardHeader>
                        <CardContent className="p-4 space-y-3">
                            {preview ? (
                                <>
                                    <div className="flex items-center gap-2 font-mono text-sm break-all">
                                        <Badge variant="outline" className={getMethodColor(call.method)}>
                                            {preview.method}
                                        </Badge>
                                        <span className="text-muted-foreground">{preview.url}</span>
                                    </div>
                                    <div className="text-xs text-muted-foreground">
                                        <span className="font-semibold">Note:</span> This is a client-side simulation. Actual headers and transformation may vary.
                                    </div>
                                </>
                            ) : (
                                <div className="text-sm text-muted-foreground italic">Fix JSON arguments to see preview.</div>
                            )}
                        </CardContent>
                    </Card>

                    <div className="pt-2 space-y-3">
                        <Alert className="py-2 bg-yellow-500/10 border-yellow-500/20 text-yellow-600 dark:text-yellow-400">
                            <AlertTitle className="text-xs font-semibold flex items-center gap-1">
                                <Save className="h-3 w-3" />
                                Important
                            </AlertTitle>
                            <AlertDescription className="text-[10px] leading-tight">
                                Execution runs the version <strong>saved on the server</strong>.
                                Unsaved changes (shown in Preview) will not be used.
                            </AlertDescription>
                        </Alert>
                        <Button
                            className="w-full"
                            onClick={handleExecute}
                            disabled={isExecuting || !!argsError || !onExecute}
                            title="Execute Tool"
                        >
                            {isExecuting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Play className="mr-2 h-4 w-4" />}
                            Execute
                        </Button>
                    </div>
                </div>

                <div className="flex flex-col h-full overflow-hidden border rounded-md bg-muted/10">
                    <div className="p-2 border-b bg-muted/20 text-xs font-medium flex justify-between items-center">
                        <span>Execution Result</span>
                        {executionResult && (
                            <Badge variant={executionResult.isError ? "destructive" : "secondary"} className="text-[10px] h-5">
                                {executionResult.isError ? "Error" : "Success"}
                            </Badge>
                        )}
                    </div>
                    <div className="flex-1 overflow-auto p-0">
                        {executionResult ? (
                            <JsonView data={executionResult} className="border-0 bg-transparent" />
                        ) : (
                            <div className="h-full flex items-center justify-center text-muted-foreground text-xs italic p-4 text-center">
                                {isExecuting ? "Executing..." : "Run the tool to see results here."}
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
}
