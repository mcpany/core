/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState, useMemo } from "react";
import { Trace, Span } from "@/types/trace";
import { User, Cpu, Box, CheckCircle, AlertCircle, ArrowRight, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { JsonView } from "@/components/ui/json-view";
import { Badge } from "@/components/ui/badge";

interface SequenceDiagramProps {
  trace: Trace;
}

interface Participant {
  id: string;
  name: string;
  icon: React.ElementType;
  color: string;
}

interface Interaction {
  id: string;
  from: string;
  to: string;
  label: string;
  payload?: unknown;
  isError?: boolean;
  type: "request" | "response";
}

export function SequenceDiagram({ trace }: SequenceDiagramProps) {
  const [selectedInteraction, setSelectedInteraction] = useState<Interaction | null>(null);

  const { participants, interactions } = useMemo(() => {
    const parts = new Map<string, Participant>();
    const acts: Interaction[] = [];

    // Default Participants
    parts.set("client", { id: "client", name: "Client", icon: User, color: "text-blue-500" });
    parts.set("core", { id: "core", name: "MCP Core", icon: Cpu, color: "text-purple-500" });

    const ensureParticipant = (id: string, name: string, icon: React.ElementType, color: string) => {
        if (!parts.has(id)) {
            parts.set(id, { id, name, icon, color });
        }
    };

    // 1. Client -> Core Request
    acts.push({
        id: "req-init",
        from: "client",
        to: "core",
        label: "Request",
        type: "request",
        payload: { tool: trace.rootSpan.name, ...trace.rootSpan.input }
    });

    // Recursive Span Processor
    const processSpan = (span: Span, parentId: string) => {
        // Determine participant for this span
        // If it's a tool, we use tool name/id. If service, use service name.
        const targetId = span.type === 'tool' ? `tool-${span.name}` : (span.serviceName || `svc-${span.name}`);
        const targetName = span.name;

        ensureParticipant(targetId, targetName, Box, "text-green-500");

        // Request Arrow
        acts.push({
            id: `req-${span.id}`,
            from: parentId,
            to: targetId,
            label: span.name, // "Execute X" or just "X"
            type: "request",
            payload: span.input
        });

        // Process children
        if (span.children) {
            span.children.forEach(child => processSpan(child, targetId));
        }

        // Response Arrow
        acts.push({
            id: `res-${span.id}`,
            from: targetId,
            to: parentId,
            label: span.status === 'error' ? "Error" : "Result",
            type: "response",
            isError: span.status === 'error',
            payload: span.output
        });
    };

    // Start processing from rootSpan (which is called by Core)
    if (trace.rootSpan) {
        processSpan(trace.rootSpan, "core");
    }

    // 4. Core -> Client Response
    acts.push({
        id: "res-init",
        from: "core",
        to: "client",
        label: trace.status === "error" ? "Error" : "Response",
        type: "response",
        isError: trace.status === "error",
        payload: trace.rootSpan?.output
    });

    return {
        participants: Array.from(parts.values()),
        interactions: acts
    };
  }, [trace]);


  // Layout constants
  const PADDING_X = 100;
  const PADDING_Y = 60;
  const COLUMN_WIDTH = 250;
  const ROW_HEIGHT = 80;
  const HEADER_HEIGHT = 60;

  const getX = (id: string) => {
    const index = participants.findIndex((p) => p.id === id);
    return PADDING_X + index * COLUMN_WIDTH;
  };

  const width = Math.max(800, PADDING_X * 2 + (participants.length - 1) * COLUMN_WIDTH);
  const height = PADDING_Y * 2 + interactions.length * ROW_HEIGHT;

  return (
    <div className="flex h-full border rounded-lg overflow-hidden bg-background">
      {/* Diagram Area */}
      <div className="flex-1 overflow-auto p-8 relative">
        <svg width={width} height={height} className="mx-auto block">
          <defs>
            <marker
              id="arrowhead"
              markerWidth="10"
              markerHeight="7"
              refX="9"
              refY="3.5"
              orient="auto"
            >
              <polygon points="0 0, 10 3.5, 0 7" fill="currentColor" className="text-muted-foreground" />
            </marker>
            <marker
                id="arrowhead-error"
                markerWidth="10"
                markerHeight="7"
                refX="9"
                refY="3.5"
                orient="auto"
            >
                <polygon points="0 0, 10 3.5, 0 7" fill="#ef4444" />
            </marker>
             <marker
                id="arrowhead-success"
                markerWidth="10"
                markerHeight="7"
                refX="9"
                refY="3.5"
                orient="auto"
            >
                <polygon points="0 0, 10 3.5, 0 7" fill="#22c55e" />
            </marker>
          </defs>

          {/* Lifelines */}
          {participants.map((p) => {
            const x = getX(p.id);
            return (
              <g key={p.id}>
                {/* Dashed Line */}
                <line
                  x1={x}
                  y1={HEADER_HEIGHT + 20}
                  x2={x}
                  y2={height - 20}
                  stroke="currentColor"
                  strokeWidth="1"
                  strokeDasharray="4 4"
                  className="text-border"
                />
              </g>
            );
          })}

          {/* Interactions */}
          {interactions.map((interaction, i) => {
            const y = HEADER_HEIGHT + PADDING_Y + i * ROW_HEIGHT;
            const x1 = getX(interaction.from);
            const x2 = getX(interaction.to);
            const isRight = x2 > x1;
            const isSelected = selectedInteraction?.id === interaction.id;

            return (
              <g
                key={interaction.id}
                className={cn(
                    "cursor-pointer transition-opacity hover:opacity-80",
                    isSelected ? "opacity-100" : "opacity-90"
                )}
                onClick={() => setSelectedInteraction(interaction)}
              >
                {/* Click target area (invisible but wider) */}
                <rect
                    x={Math.min(x1, x2)}
                    y={y - 20}
                    width={Math.abs(x2 - x1)}
                    height={40}
                    fill="transparent"
                />

                {/* Arrow Line */}
                <line
                  x1={x1 + (isRight ? 10 : -10)}
                  y1={y}
                  x2={x2 + (isRight ? -15 : 15)}
                  y2={y}
                  stroke={interaction.isError ? "#ef4444" : "currentColor"}
                  strokeWidth={isSelected ? 3 : 2}
                  className={cn(
                      !interaction.isError && "text-primary/70",
                      interaction.type === "response" && "stroke-dasharray-2"
                  )}
                  markerEnd={interaction.isError ? "url(#arrowhead-error)" : "url(#arrowhead)"}
                  strokeDasharray={interaction.type === "response" ? "5 5" : undefined}
                />

                {/* Label Background */}
                 <rect
                    x={(x1 + x2) / 2 - (interaction.label.length * 4) - 10}
                    y={y - 24}
                    width={(interaction.label.length * 8) + 20}
                    height={20}
                    rx={4}
                    fill="var(--background)"
                    className="opacity-80"
                />

                {/* Label Text */}
                <text
                  x={(x1 + x2) / 2}
                  y={y - 10}
                  textAnchor="middle"
                  className={cn(
                      "text-xs font-medium fill-foreground select-none pointer-events-none",
                      interaction.isError && "fill-red-500"
                  )}
                >
                  {interaction.label}
                </text>
              </g>
            );
          })}
        </svg>

        {/* Header Elements (DOM overlay for better styling) */}
        <div className="absolute top-0 left-0 w-full" style={{ height: HEADER_HEIGHT }}>
            {participants.map((p) => {
                const x = getX(p.id);
                return (
                    <div
                        key={p.id}
                        className="absolute transform -translate-x-1/2 flex flex-col items-center gap-2"
                        style={{ left: x, top: 10 }}
                    >
                        <div className={cn("p-2 rounded-full bg-muted border shadow-sm", p.color)}>
                            <p.icon className="w-5 h-5" />
                        </div>
                        <span className="text-xs font-bold text-muted-foreground max-w-[120px] truncate text-center" title={p.name}>
                            {p.name}
                        </span>
                    </div>
                );
            })}
        </div>
      </div>

      {/* Details Panel */}
      {selectedInteraction && (
        <div className="w-[350px] border-l bg-muted/10 flex flex-col transition-all duration-300 ease-in-out">
          <div className="p-4 border-b flex items-center justify-between bg-background">
            <h3 className="font-semibold text-sm flex items-center gap-2">
                {selectedInteraction.isError ? <AlertCircle className="w-4 h-4 text-red-500" /> : <CheckCircle className="w-4 h-4 text-green-500" />}
                Message Details
            </h3>
            <Button variant="ghost" size="icon" className="h-6 w-6" onClick={() => setSelectedInteraction(null)}>
                <X className="w-4 h-4" />
            </Button>
          </div>
          <ScrollArea className="flex-1 p-4">
            <div className="space-y-4">
                <div>
                    <span className="text-xs text-muted-foreground uppercase tracking-wider font-semibold">Interaction</span>
                    <div className="mt-1 p-2 rounded-md bg-background border text-sm flex items-center gap-2">
                         <Badge variant="outline">{selectedInteraction.from}</Badge>
                         <ArrowRight className="w-3 h-3 text-muted-foreground" />
                         <Badge variant="outline">{selectedInteraction.to}</Badge>
                    </div>
                </div>

                <div>
                     <span className="text-xs text-muted-foreground uppercase tracking-wider font-semibold">Type</span>
                     <div className="mt-1">
                        <Badge variant={selectedInteraction.type === "request" ? "default" : "secondary"}>
                            {selectedInteraction.type.toUpperCase()}
                        </Badge>
                     </div>
                </div>

                <div>
                    <span className="text-xs text-muted-foreground uppercase tracking-wider font-semibold">Payload</span>
                    <div className="mt-1">
                        <JsonView data={selectedInteraction.payload} />
                    </div>
                </div>
            </div>
          </ScrollArea>
        </div>
      )}
    </div>
  );
}
