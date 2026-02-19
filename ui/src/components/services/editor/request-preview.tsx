/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useMemo } from "react";
import { HttpCallDefinition, HttpCallDefinition_HttpMethod } from "@proto/config/v1/call";
import { ToolDefinition } from "@proto/config/v1/tool";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface RequestPreviewProps {
    call: HttpCallDefinition;
    tool: ToolDefinition;
    args: Record<string, any>;
    baseUrl?: string;
}

export function RequestPreview({ call, tool, args, baseUrl }: RequestPreviewProps) {
    const preview = useMemo(() => {
        let method = "GET";
        switch (call.method) {
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_GET: method = "GET"; break;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_POST: method = "POST"; break;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_PUT: method = "PUT"; break;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_DELETE: method = "DELETE"; break;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_PATCH: method = "PATCH"; break;
        }

        let url = call.endpointPath || "/";
        const queryParams = new URLSearchParams();
        const headers: Record<string, string> = {};
        let body: any = null;

        // Path Parameter Substitution
        const processedArgs = { ...args };

        // 1. Path Params
        for (const key of Object.keys(processedArgs)) {
            const placeholder = `{${key}}`;
            if (url.includes(placeholder)) {
                url = url.replace(placeholder, encodeURIComponent(String(processedArgs[key])));
                delete processedArgs[key]; // Consumed
            }
        }

        // 2. Query Params vs Body
        if (method === "GET" || method === "DELETE") {
            for (const key of Object.keys(processedArgs)) {
                queryParams.append(key, String(processedArgs[key]));
            }
        } else {
            // For body, we assume JSON by default for now
            if (Object.keys(processedArgs).length > 0) {
                body = processedArgs;
                headers["Content-Type"] = "application/json";
            }
        }

        const fullUrl = (baseUrl ? baseUrl.replace(/\/$/, "") : "") + url + (queryParams.toString() ? `?${queryParams.toString()}` : "");

        return {
            method,
            url: fullUrl,
            headers,
            body: body ? JSON.stringify(body, null, 2) : null
        };
    }, [call, args, baseUrl]);

    return (
        <Card className="h-full flex flex-col">
            <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Request Preview</CardTitle>
            </CardHeader>
            <CardContent className="flex-1 overflow-auto">
                <div className="space-y-4">
                    <div className="font-mono text-sm break-all">
                        <span className={cn("font-bold mr-2",
                            preview.method === "GET" ? "text-blue-500" :
                            preview.method === "POST" ? "text-green-500" :
                            preview.method === "DELETE" ? "text-red-500" : "text-yellow-500"
                        )}>{preview.method}</span>
                        <span className="text-foreground">{preview.url}</span>
                    </div>

                    {Object.keys(preview.headers).length > 0 && (
                        <div className="space-y-1">
                            <div className="text-xs text-muted-foreground font-semibold">Headers</div>
                            <pre className="text-xs bg-muted/50 p-2 rounded overflow-x-auto">
                                {Object.entries(preview.headers).map(([k, v]) => `${k}: ${v}`).join("\n")}
                            </pre>
                        </div>
                    )}

                    {preview.body && (
                        <div className="space-y-1">
                            <div className="text-xs text-muted-foreground font-semibold">Body</div>
                            <pre className="text-xs bg-muted/50 p-2 rounded overflow-x-auto font-mono">
                                {preview.body}
                            </pre>
                        </div>
                    )}

                    {!preview.body && Object.keys(preview.headers).length === 0 && (
                        <div className="text-xs text-muted-foreground italic">No headers or body.</div>
                    )}
                </div>
            </CardContent>
        </Card>
    );
}
