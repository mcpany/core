/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useMemo } from "react";
import { HttpCallDefinition, HttpCallDefinition_HttpMethod, ParameterType } from "@proto/config/v1/call";
import { ToolDefinition } from "@proto/config/v1/tool";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";

interface RequestPreviewProps {
    tool: ToolDefinition;
    call: HttpCallDefinition;
    args: Record<string, unknown>;
    baseUrl?: string;
}

/**
 * RequestPreview component.
 * Renders a preview of the HTTP request based on the tool configuration and arguments.
 * @param props - The component props.
 * @returns The rendered preview card.
 */
export function RequestPreview({ tool, call, args, baseUrl }: RequestPreviewProps) {
    const preview = useMemo(() => {
        let method = "GET";
        switch (call.method) {
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_GET: method = "GET"; break;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_POST: method = "POST"; break;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_PUT: method = "PUT"; break;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_DELETE: method = "DELETE"; break;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_PATCH: method = "PATCH"; break;
        }

        let path = call.endpointPath || "/";
        const queryParams = new URLSearchParams();
        const headers: Record<string, string> = {};
        let body: unknown = undefined;

        // Apply parameters
        if (call.parameters) {
            call.parameters.forEach(param => {
                const paramName = param.schema?.name;
                if (!paramName) return;

                const value = args[paramName];
                if (value === undefined || value === null || value === "") return;

                const strValue = String(value);

                // Simple heuristic: if path contains {name}, replace it.
                // Otherwise, add to query or body depending on method.
                if (path.includes(`{${paramName}}`)) {
                    path = path.replace(`{${paramName}}`, encodeURIComponent(strValue));
                } else {
                    if (method === "GET" || method === "DELETE") {
                        queryParams.append(paramName, strValue);
                    } else {
                        // For POST/PUT/PATCH, default to JSON body unless specified otherwise
                        if (typeof body !== 'object') body = {};
                        (body as Record<string, unknown>)[paramName] = value;
                    }
                }
            });
        }

        let fullUrl = path;
        if (baseUrl) {
            // Ensure proper join
            const base = baseUrl.endsWith('/') ? baseUrl.slice(0, -1) : baseUrl;
            const p = path.startsWith('/') ? path : '/' + path;
            fullUrl = base + p;
        }

        const queryString = queryParams.toString();
        if (queryString) {
            fullUrl += `?${queryString}`;
        }

        return {
            method,
            url: fullUrl,
            headers,
            body: body ? JSON.stringify(body, null, 2) : undefined
        };
    }, [call, args, baseUrl]);

    const methodColor = (m: string) => {
        switch (m) {
            case "GET": return "bg-blue-500/10 text-blue-600 border-blue-200 dark:border-blue-900";
            case "POST": return "bg-green-500/10 text-green-600 border-green-200 dark:border-green-900";
            case "PUT": return "bg-orange-500/10 text-orange-600 border-orange-200 dark:border-orange-900";
            case "DELETE": return "bg-red-500/10 text-red-600 border-red-200 dark:border-red-900";
            default: return "bg-gray-500/10 text-gray-600";
        }
    };

    return (
        <Card className="h-full flex flex-col shadow-sm border-muted">
            <CardHeader className="py-3 px-4 border-b bg-muted/20">
                <CardTitle className="text-sm font-medium flex items-center gap-2">
                    Request Preview
                </CardTitle>
            </CardHeader>
            <CardContent className="flex-1 p-0 font-mono text-xs">
                <ScrollArea className="h-[300px] w-full">
                    <div className="p-4 space-y-4">
                        <div className="space-y-1">
                            <div className="text-muted-foreground text-[10px] uppercase tracking-wider font-semibold">Endpoint</div>
                            <div className="flex items-center gap-2 p-2 bg-muted/30 rounded border border-muted">
                                <Badge variant="outline" className={cn("font-bold", methodColor(preview.method))}>
                                    {preview.method}
                                </Badge>
                                <span className="break-all">{preview.url}</span>
                            </div>
                        </div>

                        {preview.body && (
                            <div className="space-y-1">
                                <div className="text-muted-foreground text-[10px] uppercase tracking-wider font-semibold">Body</div>
                                <div className="p-3 bg-muted/30 rounded border border-muted overflow-x-auto">
                                    <pre>{preview.body}</pre>
                                </div>
                            </div>
                        )}

                        <div className="text-[10px] text-muted-foreground italic pt-2 border-t">
                            * Parameters are mapped based on simple heuristics. Complex transformations (Input Transformers) happen on the server.
                        </div>
                    </div>
                </ScrollArea>
            </CardContent>
        </Card>
    );
}
