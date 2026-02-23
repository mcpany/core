/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useMemo, useState } from "react";
import { Trace, Span } from "@/types/trace";
import { cn } from "@/lib/utils";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { JsonView } from "@/components/ui/json-view";
import { Badge } from "@/components/ui/badge";
import { User, Cpu, Terminal, Globe, Database, HelpCircle, ArrowRight, ArrowLeft, Clock } from "lucide-react";

interface SequenceDiagramProps {
  trace: Trace;
}

interface Actor {
  id: string;
  label: string;
  type: 'user' | 'core' | 'tool' | 'service' | 'resource';
  icon: React.ElementType;
}

interface Message {
  id: string;
  from: string;
  to: string;
  label: string;
  type: 'request' | 'response';
  payload: any;
  status: string;
  error?: string;
  timestamp: number;
}

// Config
const ACTOR_WIDTH = 120;
const ACTOR_GAP = 160;
const MESSAGE_HEIGHT = 60;
const PADDING_TOP = 60;
const PADDING_BOTTOM = 40;
const PADDING_X = 40;

export function SequenceDiagram({ trace }: SequenceDiagramProps) {
  const [selectedMessage, setSelectedMessage] = useState<Message | null>(null);

  const { actors, messages } = useMemo(() => {
    const actorMap = new Map<string, Actor>();
    const msgs: Message[] = [];

    // Helper to ensure actor exists
    const getActorId = (span: Span | null, role: 'caller' | 'callee'): string => {
        if (!span) return 'user'; // Root caller is user

        // Normalize ID
        let id = span.serviceName ? `svc:${span.serviceName}` : `tool:${span.name}`;
        if (span.type === 'core') id = 'core';

        // Root caller override
        if (role === 'caller' && !actorMap.has('core')) {
             actorMap.set('core', { id: 'core', label: 'MCP Core', type: 'core', icon: Cpu });
        }

        if (id === 'core') return 'core';

        if (!actorMap.has(id)) {
            let label = span.name;
            let type: Actor['type'] = 'tool';
            let icon = Terminal;

            if (span.type === 'service') {
                type = 'service';
                icon = Globe;
                label = span.serviceName || span.name;
            } else if (span.type === 'resource') {
                type = 'resource';
                icon = Database;
            }

            actorMap.set(id, { id, label, type, icon });
        }
        return id;
    };

    // Add User and Core by default
    actorMap.set('user', { id: 'user', label: 'User', type: 'user', icon: User });
    actorMap.set('core', { id: 'core', label: 'MCP Core', type: 'core', icon: Cpu });

    let msgId = 0;

    const traverse = (span: Span, callerId: string) => {
        const calleeId = getActorId(span, 'callee');

        // Request
        msgs.push({
            id: `msg-${msgId++}`,
            from: callerId,
            to: calleeId,
            label: span.name,
            type: 'request',
            payload: span.input,
            status: 'pending', // Initial status
            timestamp: span.startTime
        });

        // Children
        if (span.children) {
            span.children.forEach(child => traverse(child, calleeId));
        }

        // Response
        msgs.push({
            id: `msg-${msgId++}`,
            from: calleeId,
            to: callerId,
            label: `Return`,
            type: 'response',
            payload: span.output,
            status: span.status,
            error: span.errorMessage,
            timestamp: span.endTime
        });
    };

    // Start with root span
    // Root span is usually executed by Core, triggered by User
    // So we add a "User -> Core" message first if the root span is the entry point

    // Initial Trigger
    msgs.push({
        id: `msg-${msgId++}`,
        from: 'user',
        to: 'core',
        label: trace.rootSpan.name,
        type: 'request',
        payload: trace.rootSpan.input,
        status: 'pending',
        timestamp: trace.rootSpan.startTime
    });

    // Traverse root children or treat root as the first execution inside Core?
    // If root is "tool", then Core calls Tool.
    traverse(trace.rootSpan, 'core');

    // Final Response
    msgs.push({
        id: `msg-${msgId++}`,
        from: 'core',
        to: 'user',
        label: 'Response',
        type: 'response',
        payload: trace.rootSpan.output,
        status: trace.status,
        error: trace.rootSpan.errorMessage,
        timestamp: trace.rootSpan.endTime
    });

    return {
        actors: Array.from(actorMap.values()),
        messages: msgs
    };
  }, [trace]);

  const width = (actors.length - 1) * ACTOR_GAP + ACTOR_WIDTH + PADDING_X * 2;
  const height = messages.length * MESSAGE_HEIGHT + PADDING_TOP + PADDING_BOTTOM;

  const getActorX = (index: number) => PADDING_X + index * ACTOR_GAP + ACTOR_WIDTH / 2;

  return (
    <div className="w-full overflow-x-auto border rounded-md bg-white dark:bg-zinc-950 p-4">
        <svg width={width} height={height} className="font-sans text-xs">
            {/* Lifelines */}
            {actors.map((actor, i) => {
                const x = getActorX(i);
                return (
                    <g key={actor.id}>
                        {/* Line */}
                        <line
                            x1={x}
                            y1={PADDING_TOP}
                            x2={x}
                            y2={height - PADDING_BOTTOM}
                            stroke="currentColor"
                            strokeWidth={1}
                            strokeDasharray="4 4"
                            className="text-muted-foreground/30"
                        />
                        {/* Header */}
                        <foreignObject x={x - ACTOR_WIDTH / 2} y={0} width={ACTOR_WIDTH} height={50}>
                            <div className="flex flex-col items-center justify-center h-full">
                                <div className={cn(
                                    "p-2 rounded-lg border shadow-sm flex items-center justify-center mb-1",
                                    actor.type === 'user' ? "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400 border-blue-200 dark:border-blue-800" :
                                    actor.type === 'core' ? "bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400 border-purple-200 dark:border-purple-800" :
                                    actor.type === 'tool' ? "bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400 border-amber-200 dark:border-amber-800" :
                                    "bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-400 border-slate-200 dark:border-slate-700"
                                )}>
                                    <actor.icon className="w-4 h-4" />
                                </div>
                                <span className="font-semibold truncate max-w-full text-[10px] text-muted-foreground">
                                    {actor.label}
                                </span>
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
                const y = PADDING_TOP + (i + 1) * MESSAGE_HEIGHT - MESSAGE_HEIGHT / 2;
                const isRight = x2 > x1;
                const isSelf = x1 === x2;

                const color = msg.error || msg.status === 'error' ? "text-red-500" : "text-primary";
                const strokeColor = msg.error || msg.status === 'error' ? "#ef4444" : "currentColor";

                return (
                    <g
                        key={msg.id}
                        className="group cursor-pointer hover:opacity-80 transition-opacity"
                        onClick={() => setSelectedMessage(msg)}
                    >
                        {/* Hit Area */}
                        <rect
                            x={Math.min(x1, x2)}
                            y={y - 15}
                            width={Math.abs(x2 - x1) || 60}
                            height={30}
                            fill="transparent"
                        />

                        {isSelf ? (
                            <>
                                <path
                                    d={`M ${x1} ${y} L ${x1 + 30} ${y} L ${x1 + 30} ${y + 15} L ${x1} ${y + 15}`}
                                    fill="none"
                                    stroke={strokeColor}
                                    strokeWidth={1.5}
                                    markerEnd="url(#arrowhead)"
                                    className="text-muted-foreground"
                                />
                                <text x={x1 + 35} y={y + 10} className={cn("text-[10px] fill-current", color)}>
                                    {msg.label}
                                </text>
                            </>
                        ) : (
                            <>
                                <line
                                    x1={x1 + (isRight ? 0 : 0)}
                                    y1={y}
                                    x2={x2 + (isRight ? -5 : 5)}
                                    y2={y}
                                    stroke={strokeColor}
                                    strokeWidth={1.5}
                                    strokeDasharray={msg.type === 'response' ? "4 2" : "none"}
                                    markerEnd="url(#arrowhead)"
                                    className={cn("transition-all", msg.type === 'response' && "opacity-70")}
                                />
                                <rect
                                    x={(x1 + x2) / 2 - (msg.label.length * 3 + 10)}
                                    y={y - 12}
                                    width={msg.label.length * 6 + 20}
                                    height={16}
                                    rx={4}
                                    fill="var(--background)"
                                    className="stroke-border"
                                    strokeWidth={0} // Hide border for cleaner look, relying on background masking
                                />
                                <text
                                    x={(x1 + x2) / 2}
                                    y={y - 1}
                                    textAnchor="middle"
                                    className={cn("text-[10px] font-medium fill-current select-none", color)}
                                >
                                    {msg.label}
                                </text>
                            </>
                        )}
                    </g>
                );
            })}

            <defs>
                <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                    <polygon points="0 0, 10 3.5, 0 7" fill="currentColor" className="text-muted-foreground/80" />
                </marker>
            </defs>
        </svg>

        {/* Message Details Dialog */}
        <Dialog open={!!selectedMessage} onOpenChange={(open) => !open && setSelectedMessage(null)}>
            <DialogContent className="sm:max-w-lg">
                <DialogHeader>
                    <DialogTitle className="flex items-center gap-2">
                        {selectedMessage?.type === 'request' ? <ArrowRight className="h-4 w-4" /> : <ArrowLeft className="h-4 w-4" />}
                        {selectedMessage?.label}
                    </DialogTitle>
                    <DialogDescription className="flex items-center gap-2 text-xs">
                        <Clock className="h-3 w-3" />
                        {selectedMessage && new Date(selectedMessage.timestamp).toLocaleTimeString()}
                        {selectedMessage?.status === 'error' && <Badge variant="destructive" className="ml-2 h-5">Error</Badge>}
                    </DialogDescription>
                </DialogHeader>

                <div className="space-y-4">
                    {selectedMessage?.error && (
                        <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md text-red-600 dark:text-red-400 text-xs font-mono break-all">
                            {selectedMessage.error}
                        </div>
                    )}

                    <div className="space-y-1">
                        <div className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Payload</div>
                        <div className="max-h-[300px] overflow-y-auto border rounded-md">
                            <JsonView data={selectedMessage?.payload || {}} />
                        </div>
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    </div>
  );
}
