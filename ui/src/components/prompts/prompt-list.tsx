/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { MessageSquare, Search, Plus, Terminal, ChevronRight } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { PromptDefinition } from "@/lib/client";
import { cn } from "@/lib/utils";

interface PromptListProps {
    prompts: PromptDefinition[];
    selectedPrompt: PromptDefinition | null;
    onSelect: (prompt: PromptDefinition) => void;
    onNew: () => void;
    searchQuery: string;
    onSearchChange: (query: string) => void;
}

export function PromptList({
    prompts,
    selectedPrompt,
    onSelect,
    onNew,
    searchQuery,
    onSearchChange
}: PromptListProps) {
    const filteredPrompts = prompts.filter(
        (p) =>
          p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
          (p.description && p.description.toLowerCase().includes(searchQuery.toLowerCase()))
    );

    const getArguments = (prompt: PromptDefinition) => {
        if (!prompt.inputSchema || !prompt.inputSchema.properties) return [];
        // Handle google.protobuf.Struct or plain JSON object
        const props = (prompt.inputSchema as any).fields?.properties?.structValue?.fields ||
                      (prompt.inputSchema as any).properties ||
                      {};
        return Object.keys(props);
    };

    return (
        <div className="w-[300px] md:w-[350px] border-r flex flex-col bg-muted/10 h-full">
            <div className="p-4 border-b space-y-3">
                <div className="flex items-center justify-between">
                    <h3 className="font-semibold text-sm flex items-center gap-2">
                        <MessageSquare className="h-4 w-4" /> Prompt Library
                    </h3>
                    <Button variant="ghost" size="icon" onClick={onNew} className="h-8 w-8" title="Create New Prompt">
                        <Plus className="h-4 w-4" />
                    </Button>
                </div>
                <div className="relative">
                    <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                        placeholder="Search prompts..."
                        value={searchQuery}
                        onChange={(e) => onSearchChange(e.target.value)}
                        className="pl-8 h-9 text-sm"
                    />
                </div>
            </div>
            <ScrollArea className="flex-1">
                <div className="flex flex-col p-2 gap-1">
                    {filteredPrompts.map((prompt) => (
                        <button
                            key={prompt.name}
                            onClick={() => onSelect(prompt)}
                            className={cn(
                                "flex flex-col items-start gap-1 p-3 rounded-md text-left transition-colors hover:bg-accent hover:text-accent-foreground",
                                selectedPrompt?.name === prompt.name ? "bg-accent text-accent-foreground shadow-sm" : ""
                            )}
                        >
                            <div className="flex items-center justify-between w-full">
                                <span className="font-medium text-sm truncate">{prompt.name}</span>
                                {selectedPrompt?.name === prompt.name && <ChevronRight className="h-3 w-3 opacity-50" />}
                            </div>
                            {prompt.description && (
                                <p className="text-xs text-muted-foreground line-clamp-2">
                                    {prompt.description}
                                </p>
                            )}
                            <div className="flex items-center gap-2 mt-1">
                                <Badge variant="outline" className="text-[10px] px-1 py-0 h-4">
                                    {prompt.name.startsWith("user-library.") ? "User" : "System"}
                                </Badge>
                                {(getArguments(prompt).length || 0) > 0 && (
                                    <span className="text-[10px] text-muted-foreground flex items-center gap-0.5">
                                        <Terminal className="h-3 w-3" /> {getArguments(prompt).length} args
                                    </span>
                                )}
                            </div>
                        </button>
                    ))}
                    {filteredPrompts.length === 0 && (
                        <div className="p-8 text-center text-sm text-muted-foreground flex flex-col items-center gap-2">
                            <p>No prompts found.</p>
                            <Button variant="outline" size="sm" onClick={onNew} className="h-6 text-xs gap-1">
                                <Plus className="h-3 w-3" /> Create Prompt
                            </Button>
                        </div>
                    )}
                </div>
            </ScrollArea>
        </div>
    );
}
