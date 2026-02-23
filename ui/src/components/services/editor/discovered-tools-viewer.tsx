/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ToolDefinition } from "@/lib/client";
import { Badge } from "@/components/ui/badge";
import { Wrench } from "lucide-react";
import {
    Accordion,
    AccordionContent,
    AccordionItem,
    AccordionTrigger,
} from "@/components/ui/accordion";
import { JsonView } from "@/components/ui/json-view";

interface DiscoveredToolsViewerProps {
    tools: ToolDefinition[];
}

export function DiscoveredToolsViewer({ tools }: DiscoveredToolsViewerProps) {
    if (!tools || tools.length === 0) {
        return (
            <div className="text-center p-8 border border-dashed rounded-md text-muted-foreground bg-muted/20">
                <Wrench className="mx-auto h-8 w-8 mb-2 opacity-50" />
                <p>No tools discovered yet.</p>
                <p className="text-xs mt-1">Validate the service configuration to discover available tools.</p>
            </div>
        );
    }

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider flex items-center gap-2">
                    <Wrench className="h-4 w-4" /> Discovered Tools <Badge variant="secondary" className="text-xs">{tools.length}</Badge>
                </h3>
            </div>
            <div className="border rounded-md overflow-hidden bg-background">
                <Accordion type="multiple" className="w-full">
                    {tools.map((tool, index) => (
                        <AccordionItem value={`item-${index}`} key={index} className="border-b last:border-0 px-4">
                            <AccordionTrigger className="hover:no-underline py-3">
                                <div className="flex items-center gap-3 text-left w-full overflow-hidden">
                                    <span className="font-mono text-sm font-semibold text-primary truncate max-w-[200px]">
                                        {tool.name}
                                    </span>
                                    <span className="text-xs text-muted-foreground line-clamp-1 font-normal flex-1 truncate">
                                        {tool.description || "No description"}
                                    </span>
                                </div>
                            </AccordionTrigger>
                            <AccordionContent className="pb-4 pt-0">
                                <div className="space-y-3 mt-2 pl-1">
                                    {tool.description && (
                                        <div className="text-sm text-muted-foreground">
                                            {tool.description}
                                        </div>
                                    )}

                                    <div>
                                        <div className="text-xs font-semibold text-muted-foreground mb-1.5 uppercase tracking-wider">Input Schema</div>
                                        <div className="bg-muted/40 rounded-md border overflow-hidden">
                                            <JsonView
                                                data={tool.inputSchema}
                                                maxHeight={300}
                                                className="bg-transparent border-none text-xs"
                                            />
                                        </div>
                                    </div>
                                </div>
                            </AccordionContent>
                        </AccordionItem>
                    ))}
                </Accordion>
            </div>
        </div>
    );
}
