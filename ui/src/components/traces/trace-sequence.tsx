/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useEffect, useState } from "react";
import mermaid from "mermaid";
import type { Trace, Span } from "@/app/api/traces/route";
import { Loader2 } from "lucide-react";

interface TraceSequenceDiagramProps {
  trace: Trace;
}

export function TraceSequenceDiagram({ trace }: TraceSequenceDiagramProps) {
  const [svg, setSvg] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    mermaid.initialize({
      startOnLoad: false,
      theme: "default",
      securityLevel: "loose",
      fontFamily: "monospace",
    });
  }, []);

  useEffect(() => {
    if (!trace) return;

    const renderDiagram = async () => {
      setLoading(true);
      setError(null);
      try {
        const graphDefinition = generateMermaidSequence(trace);
        // mermaid.render will generate the SVG string.
        // We use a unique ID to avoid conflicts if multiple diagrams are rendered.
        const { svg } = await mermaid.render(`mermaid-${trace.id}`, graphDefinition);
        setSvg(svg);
      } catch (err) {
        console.error("Failed to render mermaid diagram", err);
        setError("Failed to render sequence diagram.");
      } finally {
        setLoading(false);
      }
    };

    renderDiagram();
  }, [trace]);

  if (loading) {
     return <div className="flex items-center justify-center p-8 text-muted-foreground"><Loader2 className="mr-2 h-4 w-4 animate-spin"/> Generating diagram...</div>;
  }

  if (error) {
      return <div className="p-4 text-red-500">{error}</div>;
  }

  return <div data-testid="mermaid-container" className="w-full overflow-auto p-4 flex justify-center bg-white rounded-md" dangerouslySetInnerHTML={{ __html: svg || '' }} />;
}

/**
 * Generates Mermaid sequence diagram syntax from a Trace.
 */
export function generateMermaidSequence(trace: Trace): string {
  let diagram = "sequenceDiagram\n";
  diagram += "    autonumber\n"; // Add numbering for clearer flow

  // Collect participants to ensure order (optional, but good for consistent layout)
  // We can just emit interactions and Mermaid figures it out, but declaring order is nicer.
  // Root -> Child -> Child...

  // Helper to sanitize names
  const sanitize = (name: string) => name.replace(/[^a-zA-Z0-9_-]/g, "_");

  const participants = new Map<string, string>();
  participants.set("User", "User");

  let interactions = "";

  // Recursive traversal to build interactions
  const buildInteractions = (span: Span, caller: string) => {
    // Determine the "callee" (this span's actor)
    let callee = span.serviceName || "System";
    if (span.type === 'tool') {
        callee = span.name;
    }

    const sanitizedCallee = sanitize(callee);
    const sanitizedCaller = sanitize(caller);

    if (!participants.has(sanitizedCallee)) {
        participants.set(sanitizedCallee, callee);
    }

    // Request
    let label = span.name;
    // Escape special characters for label? Mermaid usually handles them if we don't quote?
    // Let's strip newlines
    label = label.replace(/\n/g, " ");
    if (label.length > 40) label = label.substring(0, 37) + "...";

    interactions += `    ${sanitizedCaller}->>${sanitizedCallee}: ${label}\n`;

    // Children
    if (span.children && span.children.length > 0) {
        span.children.forEach(child => {
            buildInteractions(child, callee);
        });
    }

    // Response
    let statusLabel = span.status === 'success' ? 'OK' : 'ERR';
    if (span.errorMessage) {
        statusLabel = `ERR: ${span.errorMessage}`;
        statusLabel = statusLabel.replace(/\n/g, " ");
        if (statusLabel.length > 40) statusLabel = statusLabel.substring(0, 37) + "...";
    }

    const arrow = span.status === 'success' ? '-->>' : '--x';
    interactions += `    ${sanitizedCallee}${arrow}${sanitizedCaller}: ${statusLabel}\n`;
  };

  buildInteractions(trace.rootSpan, "User");

  // Output participants with aliases
  participants.forEach((label, id) => {
      // Escape label quotes
      const safeLabel = label.replace(/"/g, "'");
      diagram += `    participant ${id} as ${safeLabel}\n`;
  });

  diagram += interactions;

  return diagram;
}
