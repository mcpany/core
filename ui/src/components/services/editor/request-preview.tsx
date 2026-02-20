/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { HttpCallDefinition, HttpCallDefinition_HttpMethod } from "@proto/config/v1/call";
import { ToolDefinition } from "@proto/config/v1/tool";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Copy, Terminal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";

interface RequestPreviewProps {
    call: HttpCallDefinition;
    _tool: ToolDefinition;
    args: Record<string, unknown>;
    baseUrl?: string;
}

export function RequestPreview({ call, _tool, args, baseUrl = "https://api.example.com" }: RequestPreviewProps) {
    const { toast } = useToast();

    // Helper to get method string
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

    // Calculate Request
    const calculateRequest = () => {
        let path = call.endpointPath || "/";
        const queryParams = new URLSearchParams();
        const headers: Record<string, string> = {};
        let body: unknown = null;

        // Apply parameters
        call.parameters?.forEach(param => {
            const schema = param.schema;
            if (!schema) return;

            const value = args[schema.name];
            if (value === undefined || value === null) return;

            // Simple substitution logic mirroring backend (simplified)
            // Backend logic handles complex recursive mappings, but for preview we handle direct mapping.
            // If parameter name matches a path segment {name}, substitute it.
            // Else, if GET, add to query.
            // Else (POST/PUT/PATCH), add to body (unless specifically mapped elsewhere, which proto allows but simple UI assumes direct).

            // Note: Proto HttpParameterMapping has `location` field? No, it seems it infers from path.
            // Let's check proto definition if possible.
            // Based on `http-tool-editor.tsx`, we don't set location explicitly.
            // Standard MCP Any behavior:
            // 1. Path substitution
            // 2. Query param (if GET or not in path)
            // 3. Body (if POST/PUT and not in path)

            const strValue = String(value);

            if (path.includes(`{${schema.name}}`)) {
                path = path.replace(`{${schema.name}}`, encodeURIComponent(strValue));
            } else {
                const method = call.method;
                const isBodyMethod = method === HttpCallDefinition_HttpMethod.HTTP_METHOD_POST ||
                                     method === HttpCallDefinition_HttpMethod.HTTP_METHOD_PUT ||
                                     method === HttpCallDefinition_HttpMethod.HTTP_METHOD_PATCH;

                if (isBodyMethod) {
                    if (body === null) body = {};
                    (body as Record<string, unknown>)[schema.name] = value;
                } else {
                    queryParams.append(schema.name, strValue);
                }
            }
        });

        // Append query string
        const queryString = queryParams.toString();
        const fullUrl = `${baseUrl.replace(/\/$/, "")}${path}${queryString ? `?${queryString}` : ""}`;

        return {
            method: getMethodName(call.method),
            url: fullUrl,
            headers: {
                "Content-Type": body ? "application/json" : "text/plain",
                ...headers
            },
            body: body ? JSON.stringify(body, null, 2) : undefined
        };
    };

    const req = calculateRequest();

    const copyToClipboard = () => {
        const curl = `curl -X ${req.method} "${req.url}" \\
${Object.entries(req.headers).map(([k, v]) => `  -H "${k}: ${v}"`).join(" \\\n")}${req.body ? ` \\
  -d '${req.body}'` : ""}`;
        navigator.clipboard.writeText(curl);
        toast({ title: "Copied to clipboard", description: "cURL command copied." });
    };

    return (
        <Card className="h-full flex flex-col bg-muted/20 border-l-4 border-l-primary/20" data-testid="request-preview-card">
            <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium uppercase tracking-wider text-muted-foreground flex items-center gap-2">
                    <Terminal className="h-4 w-4" /> Request Preview
                </CardTitle>
            </CardHeader>
            <CardContent className="flex-1 space-y-4 font-mono text-sm overflow-auto">
                <div className="flex flex-col gap-2">
                    <div className="flex items-center gap-2">
                        <Badge variant="outline" className="font-bold">{req.method}</Badge>
                        <span className="break-all text-primary">{req.url}</span>
                    </div>
                </div>

                {Object.keys(req.headers).length > 0 && (
                     <div className="space-y-1">
                        <p className="text-xs text-muted-foreground">Headers</p>
                        {Object.entries(req.headers).map(([k, v]) => (
                            <div key={k} className="text-xs">
                                <span className="text-muted-foreground">{k}:</span> <span className="text-foreground">{v}</span>
                            </div>
                        ))}
                    </div>
                )}

                {req.body && (
                    <div className="space-y-1">
                         <p className="text-xs text-muted-foreground">Body</p>
                         <div className="bg-background/50 p-2 rounded border text-xs overflow-x-auto">
                            <pre>{req.body}</pre>
                         </div>
                    </div>
                )}

                <div className="pt-4 flex justify-end">
                    <Button variant="ghost" size="sm" onClick={copyToClipboard} className="text-xs h-7">
                        <Copy className="mr-2 h-3 w-3" /> Copy cURL
                    </Button>
                </div>
            </CardContent>
        </Card>
    );
}
