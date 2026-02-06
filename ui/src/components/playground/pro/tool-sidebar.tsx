/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ToolDefinition, PromptDefinition } from "@/lib/client";
import { Search, Command, ArrowRight, Zap, Filter, MessageSquare } from "lucide-react";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useState, useMemo } from "react";
import { cn } from "@/lib/utils";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

interface ToolSidebarProps {
    tools: ToolDefinition[];
    prompts?: PromptDefinition[];
    onSelectTool: (tool: ToolDefinition) => void;
    onSelectPrompt?: (prompt: PromptDefinition) => void;
    className?: string;
}

/**
 * ToolSidebar.
 *
 * @param className - The className.
 */
export function ToolSidebar({ tools, prompts = [], onSelectTool, onSelectPrompt, className }: ToolSidebarProps) {
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

    const filteredPrompts = useMemo(() => {
        return prompts.filter(prompt => {
            const matchesSearch = prompt.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                                  prompt.description?.toLowerCase().includes(searchQuery.toLowerCase());
            return matchesSearch;
        });
    }, [prompts, searchQuery]);

    const services = useMemo(() => {
        const s = new Set(tools.map(t => t.serviceId || 'builtin'));
        return Array.from(s);
    }, [tools]);

    return (
        <div className={cn("flex flex-col h-full bg-muted/10 border-r", className)}>
            <Tabs defaultValue="tools" className="flex-1 flex flex-col h-full">
                <div className="p-4 border-b space-y-3 shrink-0">
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2 text-sm font-semibold text-muted-foreground">
                            <Command className="h-4 w-4" />
                            Library
                        </div>
                        <TabsList className="h-7 p-0 bg-transparent gap-2">
                             <TabsTrigger value="tools" className="h-6 text-[10px] px-2 data-[state=active]:bg-muted">Tools</TabsTrigger>
                             <TabsTrigger value="prompts" className="h-6 text-[10px] px-2 data-[state=active]:bg-muted">Prompts</TabsTrigger>
                        </TabsList>
                    </div>

                    <div className="relative">
                        <Search className="absolute left-2 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
                        <Input
                            placeholder="Search..."
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="pl-8 h-9 text-xs"
                        />
                    </div>

                    <TabsContent value="tools" className="mt-0">
                        {services.length > 1 && (
                            <div className="flex gap-1 overflow-x-auto pb-1 no-scrollbar mt-2">
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
                    </TabsContent>
                </div>

                <TabsContent value="tools" className="flex-1 overflow-hidden mt-0 data-[state=inactive]:hidden">
                    <ScrollArea className="h-full">
                        <div className="p-3 space-y-2">
                            {filteredTools.length === 0 && (
                                <div className="text-center py-8 text-muted-foreground text-xs">
                                    <Filter className="h-8 w-8 mx-auto mb-2 opacity-20" />
                                    No tools found.
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
                                            <span className="font-semibold text-sm truncate max-w-[120px]" title={tool.name}>{tool.name}</span>
                                        </div>
                                        <Badge variant="secondary" className="text-[9px] h-4 px-1 shrink-0">{tool.serviceId || 'core'}</Badge>
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
                        <div className="p-3 border-t bg-muted/5 text-[10px] text-muted-foreground text-center">
                            {tools.length} available tools
                        </div>
                    </ScrollArea>
                </TabsContent>

                <TabsContent value="prompts" className="flex-1 overflow-hidden mt-0 data-[state=inactive]:hidden">
                    <ScrollArea className="h-full">
                        <div className="p-3 space-y-2">
                            {filteredPrompts.length === 0 && (
                                <div className="text-center py-8 text-muted-foreground text-xs">
                                    <MessageSquare className="h-8 w-8 mx-auto mb-2 opacity-20" />
                                    No prompts found.
                                </div>
                            )}

                            {filteredPrompts.map((prompt) => (
                                <div
                                    key={prompt.name}
                                    className="group flex flex-col gap-2 p-3 rounded-lg border bg-card hover:bg-accent/50 hover:border-primary/30 transition-all cursor-pointer shadow-sm"
                                    onClick={() => onSelectPrompt && onSelectPrompt(prompt)}
                                >
                                    <div className="flex items-start justify-between">
                                        <div className="flex items-center gap-2">
                                            <div className="bg-amber-500/10 p-1 rounded-md text-amber-500 group-hover:bg-amber-500 group-hover:text-amber-50 transition-colors">
                                                <MessageSquare className="h-3 w-3" />
                                            </div>
                                            <span className="font-semibold text-sm truncate max-w-[150px]" title={prompt.name}>{prompt.name}</span>
                                        </div>
                                    </div>

                                    <p className="text-[11px] text-muted-foreground line-clamp-2 leading-relaxed">
                                        {prompt.description || "No description provided."}
                                    </p>

                                    <div className="flex items-center justify-between pt-1 opacity-60 group-hover:opacity-100 transition-opacity">
                                        <span className="text-[10px] text-muted-foreground font-mono">
                                            Prompt
                                        </span>
                                        <Button variant="ghost" size="sm" className="h-5 text-[10px] px-2 gap-1 hover:bg-background">
                                            Use <ArrowRight className="h-2 w-2" />
                                        </Button>
                                    </div>
                                </div>
                            ))}
                        </div>
                         <div className="p-3 border-t bg-muted/5 text-[10px] text-muted-foreground text-center">
                            {prompts.length} available prompts
                        </div>
                    </ScrollArea>
                </TabsContent>
            </Tabs>
        </div>
    );
}
