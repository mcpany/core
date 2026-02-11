/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ToolDefinition } from "@/lib/client";
import { Search, Command, ArrowRight, Zap, Filter, MessageSquare, MoreHorizontal, Plus, Trash2, Edit2, Clock } from "lucide-react";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useState, useMemo } from "react";
import { cn } from "@/lib/utils";

interface Session {
    id: string;
    name: string;
    updatedAt: number;
}

interface ToolSidebarProps {
    tools: ToolDefinition[];
    onSelectTool: (tool: ToolDefinition) => void;
    className?: string;

    // Session Props
    sessions?: Session[];
    currentSessionId?: string | null;
    onCreateSession?: () => void;
    onSwitchSession?: (id: string) => void;
    onRenameSession?: (id: string, name: string) => void;
    onDeleteSession?: (id: string) => void;
}

/**
 * ToolSidebar.
 *
 * @param className - The className.
 */
export function ToolSidebar({
    tools,
    onSelectTool,
    className,
    sessions = [],
    currentSessionId,
    onCreateSession,
    onSwitchSession,
    onRenameSession,
    onDeleteSession
}: ToolSidebarProps) {
    const [searchQuery, setSearchQuery] = useState("");
    const [selectedService, setSelectedService] = useState<string | null>(null);
    const [editingSessionId, setEditingSessionId] = useState<string | null>(null);
    const [editName, setEditName] = useState("");

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

    const sortedSessions = useMemo(() => {
        return [...sessions].sort((a, b) => b.updatedAt - a.updatedAt);
    }, [sessions]);

    const startEditing = (session: Session) => {
        setEditingSessionId(session.id);
        setEditName(session.name);
    };

    const saveEditing = () => {
        if (editingSessionId && onRenameSession) {
            onRenameSession(editingSessionId, editName);
        }
        setEditingSessionId(null);
    };

    return (
        <div className={cn("flex flex-col h-full bg-muted/10 border-r", className)}>
            <Tabs defaultValue="library" className="flex flex-col h-full">
                <div className="border-b bg-background/50 backdrop-blur-sm">
                    <TabsList className="w-full justify-start h-10 bg-transparent rounded-none p-0">
                        <TabsTrigger
                            value="library"
                            className="flex-1 rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent"
                        >
                            Library
                        </TabsTrigger>
                        <TabsTrigger
                            value="history"
                            className="flex-1 rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent"
                        >
                            History
                        </TabsTrigger>
                    </TabsList>
                </div>

                <TabsContent value="library" className="flex-1 flex flex-col min-h-0 mt-0">
                    <div className="p-4 border-b space-y-3">
                        <div className="flex items-center gap-2 text-sm font-semibold text-muted-foreground">
                            <Command className="h-4 w-4" />
                            Tools
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

                <TabsContent value="history" className="flex-1 flex flex-col min-h-0 mt-0">
                    <div className="p-4 border-b">
                        <Button
                            className="w-full justify-start gap-2"
                            variant="outline"
                            onClick={onCreateSession}
                        >
                            <Plus className="h-4 w-4" /> New Session
                        </Button>
                    </div>
                    <ScrollArea className="flex-1">
                        <div className="p-3 space-y-1">
                            {sortedSessions.map((session) => (
                                <div
                                    key={session.id}
                                    className={cn(
                                        "group flex items-center justify-between p-2 rounded-md cursor-pointer text-sm transition-colors",
                                        currentSessionId === session.id
                                            ? "bg-primary/10 text-primary font-medium"
                                            : "hover:bg-muted text-muted-foreground hover:text-foreground"
                                    )}
                                    onClick={() => onSwitchSession?.(session.id)}
                                >
                                    <div className="flex items-center gap-2 flex-1 min-w-0">
                                        <MessageSquare className="h-4 w-4 shrink-0" />
                                        {editingSessionId === session.id ? (
                                            <Input
                                                value={editName}
                                                onChange={(e) => setEditName(e.target.value)}
                                                onKeyDown={(e) => {
                                                    if (e.key === "Enter") saveEditing();
                                                    if (e.key === "Escape") setEditingSessionId(null);
                                                }}
                                                onBlur={saveEditing}
                                                autoFocus
                                                className="h-6 text-xs"
                                                onClick={(e) => e.stopPropagation()}
                                            />
                                        ) : (
                                            <div className="flex flex-col min-w-0">
                                                <span className="truncate">{session.name}</span>
                                                <span className="text-[10px] opacity-60 flex items-center gap-1">
                                                    <Clock className="h-3 w-3" />
                                                    {new Date(session.updatedAt).toLocaleTimeString()}
                                                </span>
                                            </div>
                                        )}
                                    </div>

                                    <DropdownMenu>
                                        <DropdownMenuTrigger asChild>
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity"
                                                onClick={(e) => e.stopPropagation()}
                                            >
                                                <MoreHorizontal className="h-3 w-3" />
                                            </Button>
                                        </DropdownMenuTrigger>
                                        <DropdownMenuContent align="end">
                                            <DropdownMenuItem onClick={(e) => { e.stopPropagation(); startEditing(session); }}>
                                                <Edit2 className="mr-2 h-3 w-3" /> Rename
                                            </DropdownMenuItem>
                                            <DropdownMenuItem
                                                className="text-destructive focus:text-destructive"
                                                onClick={(e) => { e.stopPropagation(); onDeleteSession?.(session.id); }}
                                            >
                                                <Trash2 className="mr-2 h-3 w-3" /> Delete
                                            </DropdownMenuItem>
                                        </DropdownMenuContent>
                                    </DropdownMenu>
                                </div>
                            ))}
                        </div>
                    </ScrollArea>
                </TabsContent>
            </Tabs>
        </div>
    );
}
