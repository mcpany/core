/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState } from "react";
import { Trace, Span } from "@/types/trace";
import { User, Cpu, Terminal, ArrowRight, ArrowLeft, MessageSquare } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { JsonView } from "@/components/ui/json-view";
import { Badge } from "@/components/ui/badge";

interface SequenceDiagramProps {
  trace: Trace;
}

interface Interaction {
  id: number;
  from: "user" | "core" | "tool";
  to: "user" | "core" | "tool";
  label: string;
  type: "request" | "response";
  payload: any;
  status?: "success" | "error";
  description?: string;
}

/**
 * SequenceDiagram renders an interactive sequence diagram of a trace.
 * It visualizes the flow of messages between the user, MCP core, and tools.
 *
 * @param props - The component props.
 * @param props.trace - The trace data to visualize.
 * @returns A rendered SVG sequence diagram.
 */
export function SequenceDiagram({ trace }: SequenceDiagramProps) {
  const [selectedInteraction, setSelectedInteraction] = useState<Interaction | null>(null);

  // Derive interactions from the trace
  const interactions: Interaction[] = [];

  function traverse(span: Span, parentSource: 'user' | 'core' | 'tool' = 'user') {
      let from = parentSource;
      let to: 'user' | 'core' | 'tool' = 'core';

      if (span.type === 'core') {
          to = 'core';
          if (parentSource === 'user') from = 'user';
      } else if (span.type === 'tool') {
          to = 'tool';
          from = 'core'; // Tools are called by Core
      }

      // Request
      interactions.push({
          id: interactions.length + 1,
          from: from,
          to: to,
          label: span.name,
          type: 'request',
          payload: span.input,
          description: `Request to ${span.name}`
      });

      // Children
      if (span.children) {
          span.children.forEach(child => traverse(child, to));
      }

      // Response
      interactions.push({
          id: interactions.length + 1,
          from: to,
          to: from,
          label: `${span.name} Result`,
          type: 'response',
          payload: span.output,
          status: span.status,
          description: `Response from ${span.name}`
      });
  }

  // Start traversal
  traverse(trace.rootSpan, 'user');

  const participants = [
    { id: "user", label: "Client", icon: User, color: "text-blue-500", bg: "bg-blue-500/10" },
    { id: "core", label: "MCP Core", icon: Cpu, color: "text-purple-500", bg: "bg-purple-500/10" },
    { id: "tool", label: "Tool", icon: Terminal, color: "text-amber-500", bg: "bg-amber-500/10" },
  ];

  // Config
  const svgHeight = 400;
  const svgWidth = 800; // Responsive via viewBox
  const startY = 80;
  const stepHeight = 70;
  const colWidth = 250;
  const paddingX = 100;

  const getX = (id: string) => {
    const idx = participants.findIndex((p) => p.id === id);
    return paddingX + idx * colWidth;
  };

  return (
    <div className="w-full flex flex-col items-center py-8 select-none">
      <div className="relative w-full max-w-4xl overflow-x-auto">
        <svg
          viewBox={`0 0 ${svgWidth} ${svgHeight}`}
          className="w-full h-auto min-w-[600px] font-sans"
        >
          {/* Lifelines */}
          {participants.map((p) => {
            const x = getX(p.id);
            return (
              <g key={p.id}>
                {/* Line */}
                <line
                  x1={x}
                  y1={startY}
                  x2={x}
                  y2={svgHeight - 20}
                  stroke="currentColor"
                  strokeOpacity={0.1}
                  strokeWidth={2}
                  strokeDasharray="6 6"
                />
                {/* Header */}
                <foreignObject x={x - 60} y={0} width={120} height={80}>
                  <div className="flex flex-col items-center justify-center h-full">
                    <div
                      className={cn(
                        "p-3 rounded-xl border shadow-sm transition-transform hover:scale-105 mb-2",
                        p.bg,
                        "bg-background" // Ensure readable on dark mode
                      )}
                    >
                      <p.icon className={cn("w-6 h-6", p.color)} />
                    </div>
                    <span className="text-xs font-semibold text-muted-foreground">
                      {p.label}
                    </span>
                  </div>
                </foreignObject>
              </g>
            );
          })}

          {/* Interactions */}
          {interactions.map((interaction, i) => {
            const y = startY + (i + 1) * stepHeight * 0.8;
            const x1 = getX(interaction.from);
            const x2 = getX(interaction.to);
            const isRight = x2 > x1;
            const color =
              interaction.status === "error"
                ? "text-red-500"
                : "text-primary";
            const strokeColor =
              interaction.status === "error" ? "#ef4444" : "currentColor";

            return (
              <g
                key={interaction.id}
                className="group cursor-pointer"
                onClick={() => setSelectedInteraction(interaction)}
              >
                {/* Hit area for easier clicking */}
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
                  stroke={strokeColor}
                  strokeWidth={2}
                  className={cn(
                    "transition-all duration-300 group-hover:stroke-[3px]",
                    interaction.type === "response" && "stroke-dasharray 4 4"
                  )}
                  markerEnd={`url(#arrowhead-${interaction.status === "error" ? "error" : "default"})`}
                />

                {/* Label Box */}
                <foreignObject
                    x={Math.min(x1, x2) + Math.abs(x2 - x1) / 2 - 100}
                    y={y - 28}
                    width={200}
                    height={30}
                >
                    <div className="flex items-center justify-center">
                        <span className={cn(
                            "text-[10px] font-medium px-2 py-0.5 rounded-full bg-background border shadow-sm transition-all group-hover:scale-110",
                            color,
                            interaction.status === "error" ? "border-red-200 bg-red-50 dark:bg-red-950/30 dark:border-red-900" : "border-border"
                        )}>
                            {interaction.label}
                        </span>
                    </div>
                </foreignObject>

                {/* Payload Icon (on hover) */}
                {interaction.payload && (
                   <foreignObject
                        x={x1 + (x2 - x1) / 2 - 10}
                        y={y + 5}
                        width={20}
                        height={20}
                        className="opacity-0 group-hover:opacity-100 transition-opacity"
                   >
                        <MessageSquare className="w-4 h-4 text-muted-foreground" />
                   </foreignObject>
                )}

              </g>
            );
          })}

          {/* Defs for arrowheads */}
          <defs>
            <marker
              id="arrowhead-default"
              markerWidth="10"
              markerHeight="7"
              refX="9"
              refY="3.5"
              orient="auto"
            >
              <polygon points="0 0, 10 3.5, 0 7" fill="currentColor" className="text-primary" />
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
          </defs>
        </svg>
      </div>

      <Dialog open={!!selectedInteraction} onOpenChange={(open) => !open && setSelectedInteraction(null)}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
                {selectedInteraction?.type === 'request' ? <ArrowRight className="h-4 w-4" /> : <ArrowLeft className="h-4 w-4" />}
                {selectedInteraction?.label}
            </DialogTitle>
            <DialogDescription>
                {selectedInteraction?.description}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="flex items-center justify-between text-sm">
                 <div className="flex items-center gap-2">
                    <span className="text-muted-foreground">From:</span>
                    <Badge variant="outline" className="uppercase">{selectedInteraction?.from}</Badge>
                 </div>
                 <ArrowRight className="h-3 w-3 text-muted-foreground" />
                 <div className="flex items-center gap-2">
                    <span className="text-muted-foreground">To:</span>
                    <Badge variant="outline" className="uppercase">{selectedInteraction?.to}</Badge>
                 </div>
            </div>

            {selectedInteraction?.status === 'error' && (
                <div className="p-3 bg-red-50 dark:bg-red-950/20 text-red-600 dark:text-red-400 rounded-md text-xs font-mono border border-red-200 dark:border-red-900">
                    Error
                </div>
            )}

            <div className="space-y-2">
                <span className="text-xs font-medium text-muted-foreground">Payload</span>
                <div className="max-h-[300px] overflow-auto rounded-md border">
                    <JsonView data={selectedInteraction?.payload} />
                </div>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
