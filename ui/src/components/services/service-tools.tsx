/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { UpstreamServiceConfig, apiClient, ToolDefinition } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Play } from "lucide-react";
import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';

interface ServiceToolsProps {
    service: UpstreamServiceConfig;
}

export function ServiceTools({ service }: ServiceToolsProps) {
    const [tools, setTools] = useState<ToolDefinition[]>([]);
    const [loading, setLoading] = useState(true);

    const fetchTools = useCallback(async () => {
        setLoading(true);
        try {
            // Get dynamic tools from status
            let dynamicTools: ToolDefinition[] = [];
            try {
                const status = await apiClient.getServiceStatus(service.name);
                if (status && status.tools) {
                    dynamicTools = status.tools;
                }
            } catch (e) {
                // Ignore status fetch error, might not be running
            }

            // Get static tools from config
            let staticTools: ToolDefinition[] = [];
            if (service.httpService?.tools) staticTools = service.httpService.tools;
            else if (service.grpcService?.tools) staticTools = service.grpcService.tools;
            else if (service.commandLineService?.tools) staticTools = service.commandLineService.tools;
            else if (service.mcpService?.tools) staticTools = service.mcpService.tools;
            else if (service.openapiService?.tools) staticTools = service.openapiService.tools;

            // Merge tools, preferring dynamic ones (runtime truth)
            const toolMap = new Map<string, ToolDefinition>();
            staticTools.forEach(t => toolMap.set(t.name, t));
            dynamicTools.forEach(t => toolMap.set(t.name, t));

            setTools(Array.from(toolMap.values()));
        } catch (e) {
            console.error("Failed to fetch tools", e);
        } finally {
            setLoading(false);
        }
    }, [service]);

    useEffect(() => {
        fetchTools();
    }, [fetchTools]);

    const parseSchema = (schema: string) => {
        try {
            return JSON.stringify(JSON.parse(schema), null, 2);
        } catch {
            return schema;
        }
    }

    if (loading) {
        return <div className="text-muted-foreground text-sm">Loading tools...</div>;
    }

    if (tools.length === 0) {
        return <div className="text-muted-foreground text-sm">No tools found for this service.</div>;
    }

    return (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {tools.map((tool) => (
                <Card key={tool.name} className="flex flex-col">
                    <CardHeader className="pb-3">
                        <div className="flex items-start justify-between">
                             <CardTitle className="text-base font-bold truncate pr-2" title={tool.name}>
                                {tool.name}
                            </CardTitle>
                             <Link href={`/playground?tool=${encodeURIComponent(tool.name)}`}>
                                <Button size="sm" variant="ghost" className="h-6 w-6 p-0" title="Try in Playground">
                                    <Play className="h-4 w-4 text-green-500" />
                                </Button>
                            </Link>
                        </div>
                        <CardDescription className="line-clamp-2 text-xs">
                            {tool.description || "No description provided."}
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="flex-1 flex flex-col gap-2">
                        {tool.inputSchema && (
                             <div className="rounded-md overflow-hidden border bg-muted/50 text-[10px]">
                                <SyntaxHighlighter
                                    language="json"
                                    style={vscDarkPlus}
                                    customStyle={{ margin: 0, padding: '0.5rem' }}
                                    wrapLongLines={true}
                                >
                                    {parseSchema(tool.inputSchema)}
                                </SyntaxHighlighter>
                            </div>
                        )}
                         {!tool.inputSchema && (
                            <div className="text-xs text-muted-foreground italic mt-auto">
                                No schema definition.
                            </div>
                        )}
                    </CardContent>
                </Card>
            ))}
        </div>
    );
}
