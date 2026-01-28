/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useMemo } from "react";
import { Trace, Span } from "@/types/trace";
import { cn } from "@/lib/utils";
import { Activity, User, Server, Database, Globe, Cpu } from "lucide-react";

interface SequenceDiagramProps {
  trace: Trace;
}

interface Actor {
  id: string;
  name: string;
  type: 'client' | 'core' | 'service';
  x: number;
}

interface SequenceEvent {
  id: string;
  type: 'request' | 'response';
  from: string; // Actor ID
  to: string; // Actor ID
  label: string;
  timestamp: number;
  status: 'success' | 'error' | 'pending';
  duration?: number;
  data?: any;
}

const ACTOR_WIDTH = 120;
const ACTOR_GAP = 160;
const STEP_HEIGHT = 50; // Vertical space per event
const HEADER_HEIGHT = 60;

function resolveActorId(span: Span): string {
    if (span.type === 'core') return 'mcp-core';
    if (span.serviceName) return `service-${span.serviceName}`;
    // Default to core if no service name, unless it's the root span which we handle separately
    return 'mcp-core';
}

function getActorName(id: string): string {
    if (id === 'client') return 'Client';
    if (id === 'mcp-core') return 'MCP Any';
    return id.replace('service-', '');
}

export function SequenceDiagram({ trace }: SequenceDiagramProps) {
    const { actors, events, height } = useMemo(() => {
        const uniqueActors = new Set<string>();
        uniqueActors.add('client');
        uniqueActors.add('mcp-core');

        const eventsList: SequenceEvent[] = [];

        function traverse(span: Span, parentActorId: string) {
            const currentActorId = resolveActorId(span);
            uniqueActors.add(currentActorId);

            // Request
            eventsList.push({
                id: `req-${span.id}`,
                type: 'request',
                from: parentActorId,
                to: currentActorId,
                label: span.name,
                timestamp: span.startTime,
                status: 'pending', // Initial status
                data: span.input
            });

            if (span.children) {
                // Sort children by time to keep sequence correct
                const sortedChildren = [...span.children].sort((a, b) => a.startTime - b.startTime);
                sortedChildren.forEach(child => traverse(child, currentActorId));
            }

            // Response
            eventsList.push({
                id: `res-${span.id}`,
                type: 'response',
                from: currentActorId,
                to: parentActorId,
                label: `${span.endTime - span.startTime}ms`,
                timestamp: span.endTime,
                status: span.status,
                duration: span.endTime - span.startTime,
                data: span.output
            });
        }

        traverse(trace.rootSpan, 'client');

        // Layout Actors
        const actorsArray: Actor[] = Array.from(uniqueActors).map((id, index) => ({
            id,
            name: getActorName(id),
            type: id === 'client' ? 'client' : id === 'mcp-core' ? 'core' : 'service',
            x: 50 + index * ACTOR_GAP
        }));

        // Sort actors: Client -> Core -> Services
        actorsArray.sort((a, b) => {
             const order = { client: 0, core: 1, service: 2 };
             const typeScore = order[a.type] - order[b.type];
             if (typeScore !== 0) return typeScore;
             return a.name.localeCompare(b.name);
        });

        // Re-assign X coordinates after sort
        actorsArray.forEach((actor, index) => {
            actor.x = 50 + index * ACTOR_GAP;
        });

        return {
            actors: actorsArray,
            events: eventsList,
            height: HEADER_HEIGHT + eventsList.length * STEP_HEIGHT + 50
        };
    }, [trace]);

    const width = actors.length * ACTOR_GAP + 100;

    return (
        <div className="w-full bg-white dark:bg-zinc-950 p-4 rounded-md border text-sm">
            <svg width={width} height={height} className="font-sans">
                {/* Lifelines */}
                {actors.map(actor => (
                    <g key={actor.id}>
                        {/* Line */}
                        <line
                            x1={actor.x + ACTOR_WIDTH / 2}
                            y1={HEADER_HEIGHT}
                            x2={actor.x + ACTOR_WIDTH / 2}
                            y2={height}
                            stroke="currentColor"
                            strokeOpacity={0.2}
                            strokeDasharray="4 4"
                            strokeWidth={1}
                        />
                        {/* Header */}
                        <g transform={`translate(${actor.x}, 10)`}>
                            <rect
                                width={ACTOR_WIDTH}
                                height={40}
                                rx={4}
                                fill="currentColor"
                                className={cn(
                                    "text-muted-foreground/10",
                                    actor.type === 'client' && "text-blue-500/10",
                                    actor.type === 'core' && "text-purple-500/10",
                                    actor.type === 'service' && "text-amber-500/10",
                                )}
                            />
                            <rect
                                width={ACTOR_WIDTH}
                                height={40}
                                rx={4}
                                stroke="currentColor"
                                strokeOpacity={0.2}
                                fill="none"
                            />
                            <g transform="translate(10, 10)">
                                {actor.type === 'client' && <User className="w-5 h-5 text-blue-500" />}
                                {actor.type === 'core' && <Cpu className="w-5 h-5 text-purple-500" />}
                                {actor.type === 'service' && <Globe className="w-5 h-5 text-amber-500" />}
                            </g>
                            <text
                                x={ACTOR_WIDTH / 2 + 10}
                                y={25}
                                textAnchor="middle"
                                className="fill-foreground text-xs font-semibold"
                            >
                                {actor.name}
                            </text>
                        </g>
                    </g>
                ))}

                {/* Events */}
                {events.map((event, index) => {
                    const fromActor = actors.find(a => a.id === event.from);
                    const toActor = actors.find(a => a.id === event.to);
                    if (!fromActor || !toActor) return null;

                    const y = HEADER_HEIGHT + (index + 1) * STEP_HEIGHT;
                    const x1 = fromActor.x + ACTOR_WIDTH / 2;
                    const x2 = toActor.x + ACTOR_WIDTH / 2;
                    const isSelf = event.from === event.to;

                    const color = event.status === 'error' ? 'red' : 'currentColor';
                    const strokeClass = event.status === 'error' ? 'text-red-500' : 'text-foreground';

                    if (isSelf) {
                        // Self call loop
                        return (
                            <g key={event.id}>
                                <path
                                    d={`M ${x1} ${y - 10} L ${x1 + 30} ${y - 10} L ${x1 + 30} ${y + 10} L ${x1} ${y + 10}`}
                                    fill="none"
                                    stroke={color}
                                    strokeWidth={1.5}
                                    className={strokeClass}
                                    markerEnd={`url(#arrowhead-${event.status})`}
                                />
                                <text x={x1 + 40} y={y + 4} className={cn("text-xs fill-muted-foreground", strokeClass)}>{event.label}</text>
                            </g>
                        )
                    }

                    return (
                        <g key={event.id}>
                            {/* Arrow Line */}
                            <line
                                x1={x1}
                                y1={y}
                                x2={x2}
                                y2={y}
                                stroke={color}
                                strokeWidth={1.5}
                                strokeDasharray={event.type === 'response' ? "4 4" : "0"}
                                className={strokeClass}
                                markerEnd={`url(#arrowhead-${event.status})`}
                            />
                            {/* Label */}
                            <text
                                x={(x1 + x2) / 2}
                                y={y - 8}
                                textAnchor="middle"
                                className={cn("text-xs fill-muted-foreground font-mono", strokeClass)}
                            >
                                {event.label}
                            </text>
                        </g>
                    );
                })}

                {/* Definitions for markers */}
                <defs>
                    <marker id="arrowhead-success" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                        <polygon points="0 0, 10 3.5, 0 7" className="fill-foreground" />
                    </marker>
                    <marker id="arrowhead-pending" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                        <polygon points="0 0, 10 3.5, 0 7" className="fill-foreground" />
                    </marker>
                    <marker id="arrowhead-error" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                        <polygon points="0 0, 10 3.5, 0 7" fill="red" />
                    </marker>
                </defs>
            </svg>
        </div>
    );
}
