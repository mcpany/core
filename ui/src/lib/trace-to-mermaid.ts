/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Trace, Span } from "@/types/trace";

/**
 * Converts a Trace object into a Mermaid Sequence Diagram string.
 * @param trace - The trace object to convert.
 * @returns A string containing the Mermaid diagram syntax.
 */
export function traceToMermaid(trace: Trace): string {
  const participants = new Map<string, string>(); // ID -> Label
  const participantOrder: string[] = [];

  // Helper to get safe ID for Mermaid
  const getSafeId = (name: string) => {
    // Replace non-alphanumeric characters with underscore
    const safe = name.replace(/[^a-zA-Z0-9]/g, '_');
    return safe || 'Unknown';
  };

  // 1. Identify Participants
  // Always include User and Core
  const initiatorLabel = trace.trigger ? trace.trigger.charAt(0).toUpperCase() + trace.trigger.slice(1) : 'User';
  participants.set('User', initiatorLabel);
  participantOrder.push('User');

  participants.set('Core', 'MCP Any');
  participantOrder.push('Core');

  const collectParticipants = (span: Span) => {
    if (span.serviceName) {
      const id = getSafeId(span.serviceName);
      if (!participants.has(id)) {
        participants.set(id, span.serviceName);
        participantOrder.push(id);
      }
    }
    span.children?.forEach(collectParticipants);
  };
  collectParticipants(trace.rootSpan);

  // 2. Build Diagram Header
  let graph = "sequenceDiagram\n";
  graph += "  autonumber\n";

  // Add participants definition in order
  participantOrder.forEach(id => {
    graph += `  participant ${id} as ${participants.get(id)}\n`;
  });

  graph += "\n";

  // 3. Process Spans Recursive
  const processSpan = (span: Span, parentId: string) => {
    let currentId = 'Core'; // Default to Core

    // Determine where this span executes
    if (span.serviceName) {
      currentId = getSafeId(span.serviceName);
    } else if (span.type === 'core') {
      currentId = 'Core';
    }
    // Note: If span.type is 'tool' but no serviceName, it implies Core or unknown service.
    // We default to Core.

    // Determine message label (escape special chars if needed)
    // Mermaid handles some special chars, but let's be safe.
    const label = span.name.replace(/:/g, ' ');
    const isError = span.status === 'error';

    // Request Arrow
    if (currentId !== parentId) {
        graph += `  ${parentId}->>${currentId}: ${label}\n`;
    } else {
        // Self-call
        graph += `  ${parentId}->>${currentId}: (self) ${label}\n`;
    }

    // Process Children
    span.children?.forEach(child => processSpan(child, currentId));

    // Response Arrow
    // Calculate duration if valid
    const duration = (span.endTime > 0 && span.startTime > 0) ? (span.endTime - span.startTime) : 0;
    let suffix = duration > 0 ? `${duration}ms` : '';
    if (isError) suffix = suffix ? `${suffix} (Error)` : 'Error';

    const responseLabel = suffix || 'return';

    if (currentId !== parentId) {
        if (isError) {
             graph += `  ${currentId}-->>-${parentId}: ${responseLabel}\n`; // Dotted line with X? Mermaid doesn't support X on arrow easily.
        } else {
             graph += `  ${currentId}-->>${parentId}: ${responseLabel}\n`;
        }
    } else {
         // Self return (often omitted in simple diagrams, but let's keep for completeness)
         // graph += `  ${currentId}-->>${parentId}: ${responseLabel}\n`;
    }
  };

  // Start processing from Root
  // The Root Span is triggered by the User (or System/Webhook)
  // We treat the "Trace" trigger as the initiator.
  // For now, mapping everything to 'User' participant ID defined above.

  processSpan(trace.rootSpan, 'User');

  return graph;
}
