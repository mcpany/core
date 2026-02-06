/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ToolDefinition } from "@/lib/client";
import { Search, Command, ArrowRight, Zap, Filter, Bookmark } from "lucide-react";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useState, useMemo } from "react";
import { cn } from "@/lib/utils";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PresetList } from "./preset-list";

interface ToolSidebarProps {
    tools: ToolDefinition[];
    onSelectTool: (tool: ToolDefinition) => void;
    onSelectPreset?: (toolName: string, args: Record<string, unknown>) => void;
    className?: string;
}

/**
 * ToolSidebar.
 *
 * @param className - The className.
 */
export function ToolSidebar({ tools, onSelectTool, onSelectPreset, className }: ToolSidebarProps) {
    const [searchQuery, setSearchQuery] = useState("");
    const [selectedService, setSelectedService] = useState<string | null>(null);

    const filteredTools = useMemo(() => {
        return tools.filter(tool => {
            const matchesSearch = tool.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                                  tool.description?.toLowerCase().includes(searchQuery.toLowerCase());
            const matchesService = selectedService ? tool.serviceId === selectedService : true;
            return matchesSearch && matchesService;
        });
    }, [tools, searchQuery, selectedService]);

    const services = useMemo(() => {
        const s = new Set(tools.map(t => t.serviceId || 'builtin'));
        return Array.from(s);
    }, [tools]);

    return (
        <div className={cn("flex flex-col h-full bg-muted/10 border-r", className)}>
            <Tabs defaultValue="tools" className="flex flex-col h-full">
                <div className="px-4 pt-4 pb-2 border-b bg-muted/20">
                    <TabsList className="grid w-full grid-cols-2 h-8">
                        <TabsTrigger value="tools" className="text-xs">
                            <Command className="h-3 w-3 mr-2" />
                            Tools
                        </TabsTrigger>
                        <TabsTrigger value="presets" className="text-xs">
                            <Bookmark className="h-3 w-3 mr-2" />
                            Presets
                        </TabsTrigger>
                    </TabsList>
                </div>

                <TabsContent value="tools" className="flex-1 flex flex-col min-h-0 mt-0">
                    <div className="p-4 border-b space-y-3">
                        <div className="relative">
                            <Search className="absolute left-2 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
                            <Input
                                placeholder="Search tools..."
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                                className="pl-8 h-9 text-xs"
                            />
                        </div>
                        {services.length > 1 && (
                            <div className="flex gap-1 overflow-x-auto pb-1 no-scrollbar">
                                <Badge
                                    variant={selectedService === null ? "default" : "outline"}
                                    className="cursor-pointer text-[10px] whitespace-nowrap"
                                    onClick={() => setSelectedService(null)}
                                >
                                    All
                                </Badge>
                                {services.map(s => (
                                    <Badge
                                        key={s}
                                        variant={selectedService === s ? "default" : "outline"}
                                        className="cursor-pointer text-[10px] whitespace-nowrap"
                                        onClick={() => setSelectedService(s)}
                                    >
                                        {s}
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
                </TabsContent>

                <TabsContent value="presets" className="flex-1 flex flex-col min-h-0 mt-0">
                    {onSelectPreset ? (
                        <PresetList onSelect={onSelectPreset} />
                    ) : (
                        <div className="p-4 text-center text-xs text-muted-foreground">
                            Presets unavailable
                        </div>
                    )}
                </TabsContent>
            </Tabs>
        </div>
    );
}
