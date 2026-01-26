/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useMemo } from "react";
import { Trace, Span } from "@/types/trace";
import { cn } from "@/lib/utils";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Cpu, Database, Globe, Terminal, User, Activity } from "lucide-react";

// --- Types ---

interface Participant {
    id: string;
    name: string;
    type: Span['type'] | 'client';
    x: number;
}

interface Event {
    id: string;
    type: 'call' | 'return';
    sourceId: string;
    targetId: string;
    time: number;
    y: number;
    label: string;
    payload?: any;
    status?: 'success' | 'error';
    spanId: string;
}

interface Activation {
    participantId: string;
    startY: number;
    endY: number;
    spanId: string;
    status: 'success' | 'error' | 'pending';
}

// --- Layout Constants ---

const PARTICIPANT_WIDTH = 180;
const PARTICIPANT_GAP = 40;
const HEADER_HEIGHT = 60;
const EVENT_HEIGHT = 40; // Minimum vertical space per event
const ACTIVATION_WIDTH = 10;

// --- Helper Functions ---

function getParticipantIcon(type: Participant['type']) {
    switch (type) {
        case 'client': return <User className="h-4 w-4" />;
        case 'core': return <Cpu className="h-4 w-4 text-blue-500" />;
        case 'service': return <Globe className="h-4 w-4 text-indigo-500" />;
        case 'tool': return <Terminal className="h-4 w-4 text-amber-500" />;
        case 'resource': return <Database className="h-4 w-4 text-cyan-500" />;
        default: return <Activity className="h-4 w-4" />;
    }
}

/**
 * Calculates the layout for the sequence diagram.
 */
export function calculateLayout(trace: Trace) {
    const participants: Participant[] = [];
    const events: Event[] = [];
    const activations: Activation[] = [];

    // 1. Identify Participants
    // "Client" is always the initiator of the root span
    const clientParticipant: Participant = { id: 'client', name: 'Client', type: 'client', x: 0 };
    participants.push(clientParticipant);

    const participantMap = new Map<string, Participant>();
    participantMap.set('client', clientParticipant);

    // Recursive traversal to find participants and generate events
    function traverse(span: Span, parentId: string, depth: number) {
        // Determine participant for this span
        // If it's a 'core' span, it maps to 'Core'.
        // If it's a 'service', it maps to the service name.
        // If it's a 'tool', it maps to the tool name (or tool@service).

        let participantId = span.id; // Default unique ID
        let participantName = span.name;

        if (span.type === 'core') {
            participantId = 'core';
            participantName = 'MCP Any Core';
        } else if (span.type === 'service') {
            participantId = `service:${span.serviceName || span.name}`;
            participantName = span.name;
        } else if (span.type === 'tool') {
            participantId = `tool:${span.name}`;
            participantName = span.name;
        }

        // Add participant if new
        if (!participantMap.has(participantId)) {
            const p: Participant = {
                id: participantId,
                name: participantName,
                type: span.type,
                x: 0 // Will set later
            };
            participants.push(p);
            participantMap.set(participantId, p);
        }

        // 2. Generate Call Event
        events.push({
            id: `call-${span.id}`,
            type: 'call',
            sourceId: parentId,
            targetId: participantId,
            time: span.startTime,
            y: 0, // Will set later
            label: span.name, // Or specific method name if available
            payload: span.input,
            spanId: span.id
        });

        // Recursively process children
        if (span.children) {
            // Sort children by start time to ensure correct sequence order
            const sortedChildren = [...span.children].sort((a, b) => a.startTime - b.startTime);
            for (const child of sortedChildren) {
                traverse(child, participantId, depth + 1);
            }
        }

        // 3. Generate Return Event
        events.push({
            id: `return-${span.id}`,
            type: 'return',
            sourceId: participantId,
            targetId: parentId,
            time: span.endTime,
            y: 0, // Will set later
            label: `return`, // Could be return value summary
            payload: span.output,
            status: span.status === 'error' ? 'error' : 'success',
            spanId: span.id
        });
    }

    // Start traversal from root
    // Root span is called by 'client'
    traverse(trace.rootSpan, 'client', 0);

    // 4. Assign X Coordinates (Participants)
    // We keep the order of discovery roughly, but we might want to group by type
    // For now, discovery order (DFS) usually puts Core first, then Service, then Tool.
    participants.forEach((p, i) => {
        p.x = i * (PARTICIPANT_WIDTH + PARTICIPANT_GAP) + PARTICIPANT_WIDTH / 2;
    });

    // 5. Assign Y Coordinates (Events) based on logical order (sequence index)
    // We ignore exact timestamps for vertical spacing to ensure readability,
    // effectively converting time to logical steps.
    events.forEach((e, i) => {
        e.y = HEADER_HEIGHT + (i + 1) * EVENT_HEIGHT;
    });

    // 6. Generate Activations
    // An activation for a span starts at its Call event Y and ends at its Return event Y.
    // We need to map spanId to start/end Y.
    const spanY = new Map<string, { start: number, end: number, participantId: string }>();

    // We also need to know which participant a span belongs to.
    // Re-traverse or infer from events.
    // Events have spanId. Call event is start, Return event is end.

    events.forEach(e => {
        if (e.type === 'call') {
            if (!spanY.has(e.spanId)) {
                spanY.set(e.spanId, { start: e.y, end: 0, participantId: e.targetId });
            }
        } else if (e.type === 'return') {
            const spanData = spanY.get(e.spanId);
            if (spanData) {
                spanData.end = e.y;
            }
        }
    });

    spanY.forEach((val, key) => {
        // Find status from trace (need to look up span again? or store in event)
        // We stored status in Return event.
        const returnEvent = events.find(e => e.spanId === key && e.type === 'return');

        activations.push({
            participantId: val.participantId,
            startY: val.start,
            endY: val.end,
            spanId: key,
            status: returnEvent?.status === 'error' ? 'error' : 'success'
        });
    });

    const totalHeight = HEADER_HEIGHT + (events.length + 1) * EVENT_HEIGHT;
    const totalWidth = participants.length * (PARTICIPANT_WIDTH + PARTICIPANT_GAP);

    return { participants, events, activations, totalWidth, totalHeight };
}

/**
 * TraceSequence component.
 * @param props - The component props.
 * @param props.trace - The trace to visualize.
 * @returns The rendered component.
 */
export function TraceSequence({ trace }: { trace: Trace | null }) {
    const layout = useMemo(() => {
        if (!trace) return null;
        return calculateLayout(trace);
    }, [trace]);

    if (!trace || !layout) {
        return <div className="p-8 text-center text-muted-foreground">No trace data available.</div>;
    }

    const { participants, events, activations, totalWidth, totalHeight } = layout;

    return (
        <ScrollArea className="h-full w-full bg-white dark:bg-zinc-950/50 rounded-md border">
            <div style={{ width: Math.max(totalWidth, 800), height: totalHeight }} className="relative font-sans text-xs">

                {/* 1. Participant Headers (Sticky?) - Not easy in SVG, we just render at top */}
                {participants.map(p => (
                    <div
                        key={p.id}
                        className="absolute top-0 flex flex-col items-center justify-center border-b bg-background/90 backdrop-blur z-20"
                        style={{
                            left: p.x - PARTICIPANT_WIDTH/2,
                            width: PARTICIPANT_WIDTH,
                            height: HEADER_HEIGHT
                        }}
                    >
                        <div className="flex items-center gap-2 px-3 py-1.5 rounded-full border bg-card shadow-sm mb-1">
                            {getParticipantIcon(p.type)}
                            <span className="font-medium truncate max-w-[120px]" title={p.name}>{p.name}</span>
                        </div>
                    </div>
                ))}

                <svg width="100%" height="100%" className="absolute top-0 left-0 pointer-events-none">
                    <defs>
                        <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                            <polygon points="0 0, 10 3.5, 0 7" fill="currentColor" className="text-muted-foreground" />
                        </marker>
                        <marker id="arrowhead-error" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                            <polygon points="0 0, 10 3.5, 0 7" fill="currentColor" className="text-red-500" />
                        </marker>
                    </defs>

                    {/* 2. Lifelines */}
                    {participants.map(p => (
                        <line
                            key={`line-${p.id}`}
                            x1={p.x} y1={HEADER_HEIGHT}
                            x2={p.x} y2={totalHeight}
                            className="stroke-border stroke-dashed stroke-[1px]"
                            strokeDasharray="4 4"
                        />
                    ))}

                    {/* 3. Activations */}
                    {activations.map(a => {
                        const p = participants.find(part => part.id === a.participantId);
                        if (!p) return null;
                        return (
                            <rect
                                key={`act-${a.spanId}`}
                                x={p.x - ACTIVATION_WIDTH/2}
                                y={a.startY}
                                width={ACTIVATION_WIDTH}
                                height={a.endY - a.startY}
                                className={cn(
                                    "fill-white dark:fill-zinc-800 stroke-[1px]",
                                    a.status === 'error' ? "stroke-red-400" : "stroke-blue-400"
                                )}
                                rx={2}
                            />
                        );
                    })}

                    {/* 4. Messages (Arrows) */}
                    {events.map(e => {
                        const source = participants.find(p => p.id === e.sourceId);
                        const target = participants.find(p => p.id === e.targetId);
                        if (!source || !target) return null;

                        // Call: Solid line, arrow at target
                        // Return: Dashed line, arrow at target

                        const isCall = e.type === 'call';
                        const isSelfCall = source.id === target.id;

                        // For self-calls (recursive), we loop back
                        if (isSelfCall) {
                             return (
                                <g key={e.id}>
                                    <path
                                        d={`M ${source.x + ACTIVATION_WIDTH/2} ${e.y} L ${source.x + 40} ${e.y} L ${source.x + 40} ${e.y + 15} L ${source.x + ACTIVATION_WIDTH/2 + 2} ${e.y + 15}`}
                                        fill="none"
                                        className="stroke-muted-foreground"
                                        markerEnd="url(#arrowhead)"
                                    />
                                     <text x={source.x + 45} y={e.y + 10} className="fill-foreground text-[10px] opacity-70">
                                        {e.label}
                                    </text>
                                </g>
                            );
                        }

                        const direction = target.x > source.x ? 1 : -1;
                        const startX = source.x + (direction * ACTIVATION_WIDTH/2);
                        const endX = target.x - (direction * ACTIVATION_WIDTH/2);

                        const colorClass = e.status === 'error' ? "stroke-red-500 text-red-500" : "stroke-muted-foreground text-foreground";
                        const marker = e.status === 'error' ? "url(#arrowhead-error)" : "url(#arrowhead)";

                        return (
                            <g key={e.id}>
                                <line
                                    x1={startX} y1={e.y}
                                    x2={endX} y2={e.y}
                                    className={cn(
                                        "stroke-[1.5px]",
                                        colorClass
                                    )}
                                    strokeDasharray={isCall ? "" : "4 2"}
                                    markerEnd={marker}
                                />
                                <text
                                    x={(startX + endX) / 2}
                                    y={e.y - 6}
                                    textAnchor="middle"
                                    className={cn("text-[10px] font-medium fill-current", isCall ? "opacity-90" : "opacity-60")}
                                >
                                    {e.label}
                                </text>
                            </g>
                        );
                    })}
                </svg>
            </div>
        </ScrollArea>
    );
}
