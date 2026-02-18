/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { HttpCallDefinition, HttpCallDefinition_HttpMethod } from "@proto/config/v1/call";
import { ToolDefinition } from "@proto/config/v1/tool";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Eye } from "lucide-react";

interface RequestPreviewProps {
    call: HttpCallDefinition;
    tool: ToolDefinition;
    serviceName: string;
    args: Record<string, any>;
}

/**
 * Renders a preview of the HTTP request based on the tool definition and arguments.
 */
export function RequestPreview({ call, tool, serviceName, args }: RequestPreviewProps) {
    let url = call.endpointPath || "/";
    const queryParams = new URLSearchParams();
    let body: any = null;

    if (args) {
        // Path replacement
        Object.entries(args).forEach(([key, value]) => {
            const placeholder = `{${key}}`;
            // Simple string replacement for path params
            if (url.includes(placeholder)) {
                url = url.split(placeholder).join(String(value));
            } else {
                // Default to query if GET/DELETE, body if POST/PUT/PATCH
                const isBodyMethod =
                    call.method === HttpCallDefinition_HttpMethod.HTTP_METHOD_POST ||
                    call.method === HttpCallDefinition_HttpMethod.HTTP_METHOD_PUT ||
                    call.method === HttpCallDefinition_HttpMethod.HTTP_METHOD_PATCH;

                if (isBodyMethod) {
                    if (!body) body = {};
                    body[key] = value;
                } else {
                    queryParams.append(key, String(value));
                }
            }
        });
    }

    const fullQuery = queryParams.toString();
    const fullUrl = fullQuery ? `${url}?${fullQuery}` : url;
    const method = getMethodName(call.method);

    return (
        <Card className="h-full border-dashed bg-muted/30">
            <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium flex items-center gap-2">
                    <Eye className="h-4 w-4" /> Request Preview
                </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
                <div className="space-y-2">
                    <div className="flex items-center gap-2 font-mono text-sm break-all">
                        <Badge variant="outline" className={getMethodColor(call.method)}>{method}</Badge>
                        <span title={fullUrl}>{fullUrl}</span>
                    </div>
                </div>

                {body && (
                    <div className="space-y-1">
                        <div className="text-xs text-muted-foreground uppercase font-semibold">Body (JSON)</div>
                        <pre className="text-xs bg-muted p-2 rounded overflow-auto max-h-[200px] font-mono">
                            {JSON.stringify(body, null, 2)}
                        </pre>
                    </div>
                )}
                 {!body && (call.method === HttpCallDefinition_HttpMethod.HTTP_METHOD_POST || call.method === HttpCallDefinition_HttpMethod.HTTP_METHOD_PUT) && (
                    <div className="text-xs text-muted-foreground italic">No body content (no arguments provided or all used in path).</div>
                )}
            </CardContent>
        </Card>
    );
}

// Helpers
const getMethodName = (method: number) => {
    switch (method) {
        case HttpCallDefinition_HttpMethod.HTTP_METHOD_GET: return "GET";
        case HttpCallDefinition_HttpMethod.HTTP_METHOD_POST: return "POST";
        case HttpCallDefinition_HttpMethod.HTTP_METHOD_PUT: return "PUT";
        case HttpCallDefinition_HttpMethod.HTTP_METHOD_DELETE: return "DELETE";
        case HttpCallDefinition_HttpMethod.HTTP_METHOD_PATCH: return "PATCH";
        default: return "UNK";
    }
};

const getMethodColor = (method: number) => {
    switch (method) {
        case HttpCallDefinition_HttpMethod.HTTP_METHOD_GET: return "bg-blue-500/10 text-blue-500 border-blue-500/20";
        case HttpCallDefinition_HttpMethod.HTTP_METHOD_POST: return "bg-green-500/10 text-green-500 border-green-500/20";
        case HttpCallDefinition_HttpMethod.HTTP_METHOD_PUT: return "bg-orange-500/10 text-orange-500 border-orange-500/20";
        case HttpCallDefinition_HttpMethod.HTTP_METHOD_DELETE: return "bg-red-500/10 text-red-500 border-red-500/20";
        default: return "bg-gray-500/10 text-gray-500";
    }
};
