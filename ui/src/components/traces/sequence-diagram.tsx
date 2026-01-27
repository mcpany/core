/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo } from 'react';
import { Trace, Span } from '@/types/trace';
import { cn } from "@/lib/utils";

interface SequenceDiagramProps {
  trace: Trace;
}

interface Actor {
  id: string;
  label: string;
  type: Span['type'] | 'user';
  x: number;
}

interface Message {
  id: string;
  from: string;
  to: string;
  label: string;
  y: number;
  type: 'request' | 'response' | 'error';
  timestamp: number;
  duration?: number;
  status: 'success' | 'error';
}

/**
 * SequenceDiagram component.
 * Visualizes the trace as a sequence diagram.
 */
export function SequenceDiagram({ trace }: SequenceDiagramProps) {
  const { actors, messages, height, width } = useMemo(() => {
    const actorsMap = new Map<string, Actor>();
    const msgs: Message[] = [];
    const actorGap = 200;
    const msgGap = 40;
    const startY = 60;

    // Helper to get or create actor
    const getActorId = (span: Span): string => {
        if (span.type === 'core') return 'core';
        if (span.serviceName) return `service-${span.serviceName}`;
        if (span.type === 'tool') return `tool-${span.name}`;
        return `other-${span.id}`;
    };

    const getActorLabel = (span: Span): string => {
        if (span.type === 'core') return 'MCP Core';
        if (span.serviceName) return span.serviceName;
        if (span.type === 'tool') return span.name;
        return span.name;
    };

    // 1. Identify Actors
    // Always start with User
    actorsMap.set('user', { id: 'user', label: 'User / Client', type: 'user', x: 50 });
    // Then Core
    actorsMap.set('core', { id: 'core', label: 'MCP Core', type: 'core', x: 50 + actorGap });

    // Traverse to find other actors and messages
    const traverse = (span: Span, parentActorId: string) => {
        let currentActorId = parentActorId;

        // If this span represents a hop to another component, register it
        // Logic: Core spans stay on Core. Tool spans go to Tool actor. Service spans go to Service actor.
        // However, usually Core calls Tool.
        // If span.type is core, it stays on parent (if parent is core) or is handled by Core.

        let targetActorId = parentActorId;

        if (span.type === 'core') {
            targetActorId = 'core';
        } else if (span.type === 'service' || span.serviceName) {
            targetActorId = `service-${span.serviceName || 'unknown'}`;
            if (!actorsMap.has(targetActorId)) {
                actorsMap.set(targetActorId, {
                    id: targetActorId,
                    label: span.serviceName || 'Service',
                    type: 'service',
                    x: 0 // set later
                });
            }
        } else if (span.type === 'tool') {
             targetActorId = `tool-${span.name}`;
             if (!actorsMap.has(targetActorId)) {
                actorsMap.set(targetActorId, {
                    id: targetActorId,
                    label: span.name,
                    type: 'tool',
                    x: 0 // set later
                });
            }
        } else if (span.type === 'resource') {
             targetActorId = `res-${span.name}`;
             if (!actorsMap.has(targetActorId)) {
                actorsMap.set(targetActorId, {
                    id: targetActorId,
                    label: span.name,
                    type: 'resource',
                    x: 0
                });
            }
        }

        // Add Request Message (Parent -> Target)
        // Only if they are different actors, OR if it's a self-call we want to visualize (maybe)
        // For simplicity, only distinct actors or if explicitly desired.
        // If root span, it comes from User to Core (assuming root is handled by core)

        const isRoot = span.id === trace.rootSpan.id;
        const from = isRoot ? 'user' : parentActorId;
        // If root is 'tool' type (e.g. direct tool execution test), target is tool.
        // But usually goes through Core?
        // Let's assume User -> Target directly if root, unless target is Core.

        // If target == from, it's an internal function call, maybe skip visual message or loop?
        // Let's skip self-messages for clarity unless it's root
        if (from !== targetActorId || isRoot) {
            msgs.push({
                id: `${span.id}-req`,
                from,
                to: targetActorId,
                label: span.name,
                y: 0, // set later
                type: 'request',
                timestamp: span.startTime,
                status: span.status
            });
        }

        // Recurse children
        if (span.children) {
            span.children.forEach(child => traverse(child, targetActorId));
        }

        // Add Response Message (Target -> Parent)
        if (from !== targetActorId || isRoot) {
             msgs.push({
                id: `${span.id}-res`,
                from: targetActorId,
                to: from,
                label: span.status === 'error' ? (span.errorMessage || 'Error') : 'return',
                y: 0, // set later
                type: span.status === 'error' ? 'error' : 'response',
                timestamp: span.endTime,
                duration: span.endTime - span.startTime,
                status: span.status
            });
        }
    };

    traverse(trace.rootSpan, 'core'); // Start assuming context is Core, but root comes from User

    // Assign X coordinates for dynamically added actors
    let currentX = 50 + actorGap; // Core is here
    Array.from(actorsMap.values()).forEach(actor => {
        if (actor.id !== 'user' && actor.id !== 'core') {
            currentX += actorGap;
            actor.x = currentX;
        }
    });

    const totalWidth = currentX + 150;

    // Assign Y coordinates based on order in msgs list (which is roughly DFS traversal order,
    // but timestamps are better for strict sequence if async).
    // However, DFS structure naturally pairs req/res.
    // If we sort by timestamp, it might be better for async?
    // Let's stick to DFS order for nesting visualization, or sort by timestamp?
    // Real sequence diagrams usually follow time.
    msgs.sort((a, b) => a.timestamp - b.timestamp);

    let currentY = startY;
    msgs.forEach(msg => {
        msg.y = currentY;
        currentY += msgGap;
    });

    return {
        actors: Array.from(actorsMap.values()),
        messages: msgs,
        height: currentY + 50,
        width: totalWidth
    };
  }, [trace]);

  return (
    <div className="overflow-auto border rounded-md bg-white dark:bg-slate-950 p-4">
      <svg width={width} height={height} className="min-w-full font-sans text-xs">
        {/* Lifelines */}
        {actors.map(actor => (
            <g key={actor.id}>
                {/* Vertical Line */}
                <line
                    x1={actor.x} y1={50}
                    x2={actor.x} y2={height - 20}
                    stroke="currentColor"
                    strokeOpacity={0.2}
                    strokeDasharray="4 4"
                />

                {/* Actor Box */}
                <rect
                    x={actor.x - 60} y={10}
                    width={120} height={40}
                    rx={4}
                    className={cn(
                        "fill-muted stroke-border",
                        actor.type === 'user' && "fill-blue-100 dark:fill-blue-900/20 stroke-blue-200",
                        actor.type === 'core' && "fill-purple-100 dark:fill-purple-900/20 stroke-purple-200",
                        actor.type === 'service' && "fill-indigo-100 dark:fill-indigo-900/20 stroke-indigo-200",
                        actor.type === 'tool' && "fill-amber-100 dark:fill-amber-900/20 stroke-amber-200"
                    )}
                    strokeWidth={1}
                />
                <text
                    x={actor.x} y={35}
                    textAnchor="middle"
                    className="font-medium fill-foreground"
                >
                    {actor.label}
                </text>
            </g>
        ))}

        {/* Messages */}
        {messages.map(msg => {
            const source = actors.find(a => a.id === msg.from);
            const target = actors.find(a => a.id === msg.to);
            if (!source || !target) return null;

            const isRight = target.x > source.x;
            const color = msg.status === 'error' || msg.type === 'error' ? 'red' : 'currentColor';
            const dash = msg.type !== 'request' ? "4 4" : "0";

            return (
                <g key={msg.id} className="group hover:opacity-100">
                    {/* Arrow Line */}
                    <line
                        x1={source.x} y1={msg.y}
                        x2={target.x} y2={msg.y}
                        stroke={color}
                        strokeWidth={1.5}
                        strokeDasharray={dash}
                        className="opacity-70 group-hover:opacity-100 transition-opacity"
                    />

                    {/* Arrow Head */}
                    {isRight ? (
                        <path d={`M ${target.x - 6} ${msg.y - 4} L ${target.x} ${msg.y} L ${target.x - 6} ${msg.y + 4}`} fill="none" stroke={color} strokeWidth={1.5} />
                    ) : (
                        <path d={`M ${target.x + 6} ${msg.y - 4} L ${target.x} ${msg.y} L ${target.x + 6} ${msg.y + 4}`} fill="none" stroke={color} strokeWidth={1.5} />
                    )}

                    {/* Label */}
                    <text
                        x={(source.x + target.x) / 2}
                        y={msg.y - 6}
                        textAnchor="middle"
                        fill={color}
                        className="bg-background text-[10px]"
                    >
                        {msg.label} {msg.duration ? `(${msg.duration.toFixed(0)}ms)` : ''}
                    </text>

                    {/* Hover area for easier interaction */}
                    <rect
                        x={Math.min(source.x, target.x)}
                        y={msg.y - 10}
                        width={Math.abs(target.x - source.x)}
                        height={20}
                        fill="transparent"
                    >
                         <title>{msg.label} ({msg.type})</title>
                    </rect>
                </g>
            );
        })}
      </svg>
    </div>
  );
}
