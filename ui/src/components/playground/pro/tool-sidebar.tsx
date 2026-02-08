/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ToolDefinition } from "@/lib/client";
import { Search, Command, ArrowRight, Zap, Filter } from "lucide-react";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useState, useMemo } from "react";
import { cn } from "@/lib/utils";

interface ToolSidebarProps {
    tools: ToolDefinition[];
    onSelectTool: (tool: ToolDefinition) => void;
    className?: string;
}

/**
 * ToolSidebar.
 *
 * @param className - The className.
 */
interface ExtendedToolDefinition extends ToolDefinition {
    tags?: string[];
}

export function ToolSidebar({ tools, onSelectTool, className }: ToolSidebarProps) {
    const [searchQuery, setSearchQuery] = useState("");
    const [selectedFilter, setSelectedFilter] = useState<{ type: 'service' | 'tag', value: string } | null>(null);

    const filteredTools = useMemo(() => {
        return tools.filter(tool => {
            const matchesSearch = tool.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                                  tool.description?.toLowerCase().includes(searchQuery.toLowerCase());

            if (!matchesSearch) return false;
            if (!selectedFilter) return true;

            if (selectedFilter.type === 'service') {
                return (tool.serviceId || 'builtin') === selectedFilter.value;
            } else {
                const t = tool as ExtendedToolDefinition;
                return t.tags && Array.isArray(t.tags) && t.tags.includes(selectedFilter.value);
            }
        });
    }, [tools, searchQuery, selectedFilter]);

    const filters = useMemo(() => {
        const services = new Set<string>();
        const tags = new Set<string>();

        tools.forEach(t => {
            services.add(t.serviceId || 'builtin');
            const tool = t as ExtendedToolDefinition;
            if (tool.tags && Array.isArray(tool.tags)) {
                tool.tags.forEach((tag: string) => tags.add(tag));
            }
        });

        return {
            services: Array.from(services).sort(),
            tags: Array.from(tags).sort()
        };
    }, [tools]);

    return (
        <div className={cn("flex flex-col h-full bg-muted/10 border-r", className)}>
            <div className="p-4 border-b space-y-3">
                <div className="flex items-center gap-2 text-sm font-semibold text-muted-foreground">
                    <Command className="h-4 w-4" />
                    Library
                </div>
                <div className="relative">
                    <Search className="absolute left-2 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
                    <Input
                        placeholder="Search tools..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="pl-8 h-9 text-xs"
                    />
                </div>
                {(filters.services.length > 1 || filters.tags.length > 0) && (
                    <div className="flex gap-1 overflow-x-auto pb-1 no-scrollbar">
                        <Badge
                            variant={selectedFilter === null ? "default" : "outline"}
                            className="cursor-pointer text-[10px] whitespace-nowrap"
                            onClick={() => setSelectedFilter(null)}
                        >
                            All
                        </Badge>
                        {filters.services.map(s => (
                            <Badge
                                key={`svc-${s}`}
                                variant={selectedFilter?.type === 'service' && selectedFilter.value === s ? "default" : "outline"}
                                className="cursor-pointer text-[10px] whitespace-nowrap"
                                onClick={() => setSelectedFilter({ type: 'service', value: s })}
                            >
                                {s}
                            </Badge>
                        ))}
                        {filters.tags.map(t => (
                            <Badge
                                key={`tag-${t}`}
                                variant={selectedFilter?.type === 'tag' && selectedFilter.value === t ? "secondary" : "outline"}
                                className="cursor-pointer text-[10px] whitespace-nowrap border-primary/20"
                                onClick={() => setSelectedFilter({ type: 'tag', value: t })}
                            >
                                #{t}
                            </Badge>
                        ))}
                    </div>
                )}
            </div>

            <ScrollArea className="flex-1">
                <div className="p-3 space-y-2">
                    {filteredTools.length === 0 && (
                        <div className="text-center py-8 text-muted-foreground text-xs">
                            <Filter className="h-8 w-8 mx-auto mb-2 opacity-20" />
                            No tools found matching your search.
                        </div>
                    )}

                    {filteredTools.map((tool) => (
                        <div
                            key={tool.name}
                            className="group flex flex-col gap-2 p-3 rounded-lg border bg-card hover:bg-accent/50 hover:border-primary/30 transition-all cursor-pointer shadow-sm"
                            onClick={() => onSelectTool(tool)}
                        >
                            <div className="flex items-start justify-between">
                                <div className="flex items-center gap-2">
                                    <div className="bg-primary/10 p-1 rounded-md text-primary group-hover:bg-primary group-hover:text-primary-foreground transition-colors">
                                        <Zap className="h-3 w-3" />
                                    </div>
                                    <span className="font-semibold text-sm">{tool.name}</span>
                                </div>
                                <Badge variant="secondary" className="text-[9px] h-4 px-1">{tool.serviceId || 'core'}</Badge>
                            </div>

                            <p className="text-[11px] text-muted-foreground line-clamp-2 leading-relaxed">
                                {tool.description || "No description provided."}
                            </p>

                            <div className="flex items-center justify-between pt-1 opacity-60 group-hover:opacity-100 transition-opacity">
                                <span className="text-[10px] text-muted-foreground font-mono">
                                    {Object.keys(tool.inputSchema?.properties || {}).length} args
                                </span>
                                <Button variant="ghost" size="sm" className="h-5 text-[10px] px-2 gap-1 hover:bg-background">
                                    Use <ArrowRight className="h-2 w-2" />
                                </Button>
                            </div>
                        </div>
                    ))}
                </div>
            </ScrollArea>

            <div className="p-3 border-t bg-muted/5 text-[10px] text-muted-foreground text-center">
                {tools.length} available tools
            </div>
        </div>
    );
}
