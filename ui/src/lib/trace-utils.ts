/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Represents the status of a span.
 */
export type SpanStatus = 'success' | 'error' | 'pending';

/**
 * Represents a span in a trace.
 */
export interface Span {
  id: string;
  name: string;
  type: 'tool' | 'service' | 'resource' | 'prompt' | 'core';
  startTime: number;
  endTime: number;
  status: SpanStatus;
  input?: Record<string, any>;
  output?: Record<string, any>;
  children?: Span[];
  serviceName?: string;
  errorMessage?: string;
}

/**
 * Represents a full trace.
 */
export interface Trace {
  id: string;
  rootSpan: Span;
  timestamp: string;
  totalDuration: number;
  status: SpanStatus;
  trigger: 'user' | 'webhook' | 'scheduler' | 'system';
  isSummary?: boolean;
}

export interface DebugEntry {
  id: string;
  timestamp: string;
  method: string;
  path: string;
  status: number;
  duration: number; // nanoseconds
  request_headers: Record<string, string[]>;
  response_headers: Record<string, string[]>;
  request_body: string;
  response_body: string;
}

/**
 * Maps a DebugEntry to a Trace object.
 * @param entry The debug entry to map.
 * @param isSummary Whether this is a summary trace (missing bodies).
 * @returns The mapped Trace object.
 */
export function mapEntryToTrace(entry: DebugEntry, isSummary = false): Trace {
    const startTime = new Date(entry.timestamp).getTime();
    const durationMs = entry.duration / 1000000; // ns to ms

    let input: Record<string, any> | undefined;
    // If request_body is empty string (summary mode), keep input undefined or empty
    if (!entry.request_body) {
         // Optimization: don't parse empty body
    } else {
        try {
            input = JSON.parse(entry.request_body);
        } catch {
            input = { raw: entry.request_body };
        }
    }

    let output: Record<string, any> | undefined;
    if (!entry.response_body) {
        // Optimization: don't parse empty body
    } else {
        try {
            output = JSON.parse(entry.response_body);
        } catch {
            output = { raw: entry.response_body };
        }
    }

    const span: Span = {
        id: entry.id,
        name: `${entry.method} ${entry.path}`,
        type: 'tool', // Assume tool call for now
        startTime: startTime,
        endTime: startTime + durationMs,
        status: entry.status >= 400 ? 'error' : 'success',
        input: input,
        output: output,
        children: [],
        serviceName: 'backend'
    };

    return {
        id: entry.id,
        rootSpan: span,
        timestamp: entry.timestamp,
        totalDuration: durationMs,
        status: span.status,
        trigger: 'user',
        isSummary
    };
}
