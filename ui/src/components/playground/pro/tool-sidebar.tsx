/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ExtendedToolDefinition } from "@/lib/client";
import { Search, Command, ArrowRight, Zap, Filter, Tag } from "lucide-react";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useState, useMemo } from "react";
import { cn } from "@/lib/utils";

interface ToolSidebarProps {
    tools: ExtendedToolDefinition[];
    onSelectTool: (tool: ExtendedToolDefinition) => void;
    className?: string;
}

type FilterType = { type: 'service' | 'tag', value: string } | null;

/**
 * ToolSidebar.
 *
 * @param className - The className.
 */
export function ToolSidebar({ tools, onSelectTool, className }: ToolSidebarProps) {
    const [searchQuery, setSearchQuery] = useState("");
    const [activeFilter, setActiveFilter] = useState<FilterType>(null);

    const filteredTools = useMemo(() => {
        return tools.filter(tool => {
            const matchesSearch = tool.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                                  tool.description?.toLowerCase().includes(searchQuery.toLowerCase());

            let matchesFilter = true;
            if (activeFilter) {
                if (activeFilter.type === 'service') {
                    matchesFilter = (tool.serviceId || 'builtin') === activeFilter.value;
                } else if (activeFilter.type === 'tag') {
                    matchesFilter = tool.tags?.includes(activeFilter.value) || false;
                }
            }
            return matchesSearch && matchesFilter;
        });
    }, [tools, searchQuery, activeFilter]);

    const services = useMemo(() => {
        const s = new Set(tools.map(t => t.serviceId || 'builtin'));
        return Array.from(s).sort();
    }, [tools]);

    const tags = useMemo(() => {
        const t = new Set<string>();
        tools.forEach(tool => {
            tool.tags?.forEach(tag => t.add(tag));
        });
        return Array.from(t).sort();
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
                {(services.length > 1 || tags.length > 0) && (
                    <div className="flex gap-1 overflow-x-auto pb-1 no-scrollbar items-center">
                        <Badge
                            variant={activeFilter === null ? "default" : "outline"}
                            className="cursor-pointer text-[10px] whitespace-nowrap"
                            onClick={() => setActiveFilter(null)}
                        >
                            All
                        </Badge>
                        {services.map(s => (
                            <Badge
                                key={`svc-${s}`}
                                variant={activeFilter?.type === 'service' && activeFilter.value === s ? "default" : "outline"}
                                className="cursor-pointer text-[10px] whitespace-nowrap"
                                onClick={() => setActiveFilter({ type: 'service', value: s })}
                            >
                                {s}
                            </Badge>
                        ))}
                        {tags.length > 0 && <div className="w-px h-4 bg-border mx-1 shrink-0" />}
                        {tags.map(t => (
                            <Badge
                                key={`tag-${t}`}
                                variant={activeFilter?.type === 'tag' && activeFilter.value === t ? "secondary" : "outline"}
                                className={cn("cursor-pointer text-[10px] whitespace-nowrap flex items-center gap-1",
                                    activeFilter?.type === 'tag' && activeFilter.value === t ? "bg-blue-100 text-blue-800 hover:bg-blue-200 dark:bg-blue-900/30 dark:text-blue-300" : "border-dashed"
                                )}
                                onClick={() => setActiveFilter({ type: 'tag', value: t })}
                            >
                                <Tag className="h-2 w-2" />
                                {t}
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
