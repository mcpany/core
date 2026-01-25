/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ScrollArea } from "@/components/ui/scroll-area";
import { MCPMessage } from "@/hooks/use-inspector-stream";
import { Badge } from "@/components/ui/badge";
import { ArrowDown, ArrowUp, AlertTriangle } from "lucide-react";
import { cn } from "@/lib/utils";

interface MessageListProps {
    messages: MCPMessage[];
    selectedId: string | null;
    onSelect: (message: MCPMessage) => void;
}

export function MessageList({ messages, selectedId, onSelect }: MessageListProps) {
    if (messages.length === 0) {
        return (
            <div className="h-full flex items-center justify-center text-muted-foreground p-4 text-center text-sm">
                No messages captured yet. <br /> Use the toolbar to simulate traffic or connect a client.
            </div>
        );
    }

    return (
        <ScrollArea className="h-full">
            <div className="flex flex-col divide-y divide-border/50">
                {messages.map((msg) => (
                    <button
                        key={msg.id}
                        onClick={() => onSelect(msg)}
                        className={cn(
                            "flex items-start gap-3 p-3 text-left hover:bg-muted/50 transition-colors w-full",
                            selectedId === msg.id && "bg-muted border-l-2 border-l-primary pl-[10px]"
                        )}
                    >
                        <div className={cn(
                            "mt-1 w-6 h-6 rounded-full flex items-center justify-center shrink-0",
                            msg.direction === 'inbound' ? "bg-blue-500/10 text-blue-500" : "bg-green-500/10 text-green-500",
                            msg.isError && "bg-destructive/10 text-destructive"
                        )}>
                            {msg.isError ? (
                                <AlertTriangle className="h-3 w-3" />
                            ) : msg.direction === 'inbound' ? (
                                <ArrowDown className="h-3 w-3" />
                            ) : (
                                <ArrowUp className="h-3 w-3" />
                            )}
                        </div>

                        <div className="flex-1 min-w-0 overflow-hidden">
                            <div className="flex items-center gap-2 mb-1">
                                {msg.method && (
                                    <Badge variant="outline" className="font-mono text-[10px] h-5 px-1.5 truncate max-w-[120px]">
                                        {msg.method}
                                    </Badge>
                                )}
                                <span className="text-xs text-muted-foreground ml-auto font-mono">
                                    {new Date(msg.timestamp).toLocaleTimeString()}
                                </span>
                            </div>
                            <div className="text-xs font-mono text-muted-foreground truncate">
                                ID: {msg.id}
                            </div>
                            {msg.isError && (
                                <div className="text-xs text-destructive mt-1 truncate">
                                    Error: {msg.payload?.error?.message || "Unknown error"}
                                </div>
                            )}
                        </div>
                    </button>
                ))}
            </div>
        </ScrollArea>
    );
}
