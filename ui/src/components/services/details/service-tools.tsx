/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ToolDefinition } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Play, Code, Info } from "lucide-react";
import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";

interface ServiceToolsProps {
    tools: ToolDefinition[];
}

/**
 * ServiceTools component.
 * Displays a grid of tools provided by the service.
 * @param props - The component props.
 * @param props.tools - The list of tools.
 * @returns The rendered component.
 */
export function ServiceTools({ tools }: ServiceToolsProps) {
    if (!tools || tools.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center py-12 text-muted-foreground bg-muted/20 rounded-lg border-2 border-dashed">
                <Code className="h-10 w-10 mb-4 opacity-50" />
                <h3 className="text-lg font-medium">No tools discovered</h3>
                <p className="text-sm">This service does not expose any tools or they haven't been discovered yet.</p>
            </div>
        );
    }

    return (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {tools.map((tool) => (
                <Card key={tool.name} className="flex flex-col h-full hover:shadow-md transition-shadow">
                    <CardHeader className="pb-3">
                        <div className="flex items-start justify-between gap-2">
                            <CardTitle className="font-mono text-base break-all flex-1">
                                {tool.name}
                            </CardTitle>
                            <Badge variant="outline" className="shrink-0">
                                Tool
                            </Badge>
                        </div>
                        <CardDescription className="line-clamp-2 min-h-[40px]">
                            {tool.description || "No description provided."}
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="flex-1 pb-2">
                        <div className="space-y-2">
                             <div className="flex items-center text-xs font-medium text-muted-foreground mb-1">
                                <Info className="h-3 w-3 mr-1" />
                                Input Schema
                            </div>
                            <ScrollArea className="h-[100px] w-full rounded-md border bg-muted/50 p-2">
                                <pre className="text-[10px] leading-relaxed font-mono">
                                    {JSON.stringify(tool.inputSchema || {}, null, 2)}
                                </pre>
                            </ScrollArea>
                        </div>
                    </CardContent>
                    <CardFooter className="pt-2">
                        <Link href={`/playground?tool=${encodeURIComponent(tool.name)}`} className="w-full">
                            <Button className="w-full gap-2 group">
                                <Play className="h-4 w-4 group-hover:fill-current transition-all" />
                                Test in Playground
                            </Button>
                        </Link>
                    </CardFooter>
                </Card>
            ))}
        </div>
    );
}
