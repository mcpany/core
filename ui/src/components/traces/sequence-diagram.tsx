/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useMemo } from "react";
import { Trace, Span } from "@/types/trace";
import { cn } from "@/lib/utils";
import { ScrollArea } from "@/components/ui/scroll-area";
import { User, Cpu, Globe, Terminal, Database, Activity } from "lucide-react";
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from "@/components/ui/tooltip";

interface SequenceDiagramProps {
    trace: Trace;
}

interface Actor {
    id: string;
    name: string;
    type: Span['type'] | 'user';
    icon: React.ReactNode;
}

interface Message {
    id: string;
    from: string;
    to: string;
    label: string;
    type: 'request' | 'response';
    status: 'success' | 'error';
    timestamp: number;
    duration?: number;
    details?: string;
}

export function SequenceDiagram({ trace }: SequenceDiagramProps) {
    const { actors, messages } = useMemo(() => {
        const actorsMap = new Map<string, Actor>();
        const messagesList: Message[] = [];

        // Always add User/Client and Core
        actorsMap.set('user', { id: 'user', name: 'User / Client', type: 'user', icon: <User className="h-4 w-4" /> });

        // Helper to get or create actor
        const getActorId = (span: Span): string => {
             // For core spans, map to a singleton Core actor
             if (span.type === 'core') {
                 if (!actorsMap.has('core')) {
                     actorsMap.set('core', { id: 'core', name: 'MCP Core', type: 'core', icon: <Cpu className="h-4 w-4" /> });
                 }
                 return 'core';
             }

             // For services/tools, use their unique ID/Name
             const id = `${span.type}-${span.name}`; // Simple ID generation
             if (!actorsMap.has(id)) {
                 let icon = <Activity className="h-4 w-4" />;
                 if (span.type === 'service') icon = <Globe className="h-4 w-4" />;
                 if (span.type === 'tool') icon = <Terminal className="h-4 w-4" />;
                 if (span.type === 'resource') icon = <Database className="h-4 w-4" />;

                 actorsMap.set(id, {
                     id,
                     name: span.name,
                     type: span.type,
                     icon
                 });
             }
             return id;
        };

        // Recursive function to process spans
        const processSpan = (span: Span, parentId: string) => {
            const currentActorId = getActorId(span);

            // Request Message (Parent -> Current)
            // If parent is same as current (e.g. internal function call in same service), we can skip or show self-call
            // For sequence diagrams, usually we want to see calls between components.
            if (parentId !== currentActorId) {
                messagesList.push({
                    id: `${span.id}-req`,
                    from: parentId,
                    to: currentActorId,
                    label: span.name,
                    type: 'request',
                    status: 'success', // Request is always "sent"
                    timestamp: span.startTime,
                    details: JSON.stringify(span.input || {}, null, 2)
                });
            }

            // Process children
            if (span.children) {
                span.children.forEach(child => processSpan(child, currentActorId));
            }

            // Response Message (Current -> Parent)
            if (parentId !== currentActorId) {
                messagesList.push({
                    id: `${span.id}-res`,
                    from: currentActorId,
                    to: parentId,
                    label: `return`, // or output summary?
                    type: 'response',
                    status: span.status,
                    timestamp: span.endTime,
                    duration: span.endTime - span.startTime,
                    details: span.errorMessage || JSON.stringify(span.output || {}, null, 2)
                });
            }
        };

        // Start processing from root
        // Root span parent is 'user'
        processSpan(trace.rootSpan, 'user');

        return {
            actors: Array.from(actorsMap.values()),
            messages: messagesList.sort((a, b) => a.timestamp - b.timestamp)
        };
    }, [trace]);

    // Layout Configuration
    const actorWidth = 160;
    const actorGap = 40;
    const msgHeight = 40;
    const headerHeight = 60;
    const topPadding = 20;

    const width = actors.length * actorWidth + (actors.length - 1) * actorGap + 40; // + padding
    const height = messages.length * msgHeight + headerHeight + 40;

    const getActorX = (index: number) => 20 + index * (actorWidth + actorGap) + actorWidth / 2;

    return (
        <ScrollArea className="h-full w-full bg-white dark:bg-zinc-950 border rounded-md">
            <div className="min-w-full p-4 flex justify-center">
                <svg width={width} height={height} className="font-sans text-xs">
                    <defs>
                        <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                            <polygon points="0 0, 10 3.5, 0 7" fill="currentColor" className="text-muted-foreground" />
                        </marker>
                         <marker id="arrowhead-error" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                            <polygon points="0 0, 10 3.5, 0 7" fill="currentColor" className="text-red-500" />
                        </marker>
                    </defs>

                    {/* Lifelines */}
                    {actors.map((actor, i) => {
                        const x = getActorX(i);
                        return (
                            <g key={actor.id}>
                                {/* Line */}
                                <line
                                    x1={x} y1={headerHeight}
                                    x2={x} y2={height - 20}
                                    stroke="currentColor"
                                    strokeDasharray="4 4"
                                    className="text-border opacity-50"
                                />

                                {/* Actor Box */}
                                <foreignObject x={x - actorWidth/2} y={0} width={actorWidth} height={headerHeight}>
                                    <div className="flex flex-col items-center justify-center h-full p-2 border rounded-lg bg-card shadow-sm">
                                        <div className={cn(
                                            "p-1.5 rounded-full mb-1",
                                            actor.type === 'user' ? "bg-zinc-100 dark:bg-zinc-800" :
                                            actor.type === 'core' ? "bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400" :
                                            actor.type === 'tool' ? "bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400" :
                                            "bg-indigo-100 text-indigo-600 dark:bg-indigo-900/30 dark:text-indigo-400"
                                        )}>
                                            {actor.icon}
                                        </div>
                                        <div className="font-medium truncate w-full text-center text-xs" title={actor.name}>
                                            {actor.name}
                                        </div>
                                    </div>
                                </foreignObject>
                            </g>
                        );
                    })}

                    {/* Messages */}
                    {messages.map((msg, i) => {
                        const fromIndex = actors.findIndex(a => a.id === msg.from);
                        const toIndex = actors.findIndex(a => a.id === msg.to);

                        if (fromIndex === -1 || toIndex === -1) return null;

                        const x1 = getActorX(fromIndex);
                        const x2 = getActorX(toIndex);
                        const y = headerHeight + topPadding + i * msgHeight;

                        const isRight = x2 > x1;
                        const color = msg.status === 'error' ? "text-red-500" : "text-foreground";
                        const strokeColor = msg.status === 'error' ? "red" : "currentColor";

                        return (
                            <g key={msg.id} className="group cursor-pointer">
                                <TooltipProvider>
                                    <Tooltip>
                                        <TooltipTrigger asChild>
                                             {/* Hit area for easier hovering */}
                                            <rect
                                                x={Math.min(x1, x2)}
                                                y={y - 15}
                                                width={Math.abs(x2 - x1)}
                                                height={20}
                                                fill="transparent"
                                            />
                                        </TooltipTrigger>
                                        <TooltipContent className="max-w-[300px] font-mono text-xs">
                                            <div className="font-bold mb-1 border-b pb-1">{msg.label}</div>
                                            {msg.duration && <div>Duration: {msg.duration}ms</div>}
                                            <div className="max-h-[200px] overflow-y-auto whitespace-pre-wrap break-all text-[10px] text-muted-foreground">
                                                {msg.details}
                                            </div>
                                        </TooltipContent>
                                    </Tooltip>
                                </TooltipProvider>

                                {/* Arrow Line */}
                                <line
                                    x1={x1} y1={y}
                                    x2={isRight ? x2 - 5 : x2 + 5} y2={y}
                                    stroke={strokeColor}
                                    strokeWidth="1.5"
                                    strokeDasharray={msg.type === 'response' ? "4 2" : "0"}
                                    markerEnd={msg.status === 'error' ? "url(#arrowhead-error)" : "url(#arrowhead)"}
                                    className={cn("transition-all opacity-70 group-hover:opacity-100 group-hover:stroke-[2px]", color)}
                                />

                                {/* Label */}
                                <text
                                    x={(x1 + x2) / 2}
                                    y={y - 5}
                                    textAnchor="middle"
                                    fill="currentColor"
                                    className={cn("text-[10px] font-medium opacity-80 group-hover:opacity-100 select-none", color)}
                                >
                                    {msg.label} {msg.duration ? `(${msg.duration}ms)` : ''}
                                </text>
                            </g>
                        );
                    })}
                </svg>
            </div>
        </ScrollArea>
    );
}
