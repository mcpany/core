/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { Trace, Span } from '@/types/trace';

export type { SpanStatus, Span, Trace } from '@/types/trace';

interface DebugEntry {
  id: string;
  trace_id?: string;
  span_id?: string;
  parent_id?: string;
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
 * GET.
 *
 * @param request - The request.
 */
export async function GET(request: Request) {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:50059';

  try {
    const res = await fetch(`${backendUrl}/debug/entries`, {
        headers: {
            'Authorization': request.headers.get('Authorization') || '',
            'X-API-Key': request.headers.get('X-API-Key') || process.env.MCPANY_API_KEY || ''
        },
        cache: 'no-store'
    });

    if (!res.ok) {
        console.warn(`Failed to fetch traces from ${backendUrl}/debug/entries: ${res.status} ${res.statusText}`);
        return NextResponse.json([]);
    }

    const entries: DebugEntry[] = await res.json();

    // Validate if entries is array
    if (!Array.isArray(entries)) {
        console.error("Backend returned non-array for traces");
        return NextResponse.json([]);
    }

    // Group entries by trace_id
    const traceGroups = new Map<string, DebugEntry[]>();
    entries.forEach(entry => {
        const traceId = entry.trace_id || entry.id;
        if (!traceGroups.has(traceId)) {
            traceGroups.set(traceId, []);
        }
        traceGroups.get(traceId)!.push(entry);
    });

    const traces: Trace[] = [];

    for (const [traceId, group] of traceGroups) {
        // Sort group by timestamp
        group.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());

        // Map spans
        const spanMap = new Map<string, Span>();

        group.forEach(entry => {
            const startTime = new Date(entry.timestamp).getTime();
            const durationMs = entry.duration / 1000000; // ns to ms

            let input: Record<string, any> | undefined;
            try {
                input = JSON.parse(entry.request_body);
            } catch {
                input = { raw: entry.request_body };
            }

            let output: Record<string, any> | undefined;
            try {
                output = JSON.parse(entry.response_body);
            } catch {
                output = { raw: entry.response_body };
            }

            let errorMessage: string | undefined;
            if (entry.status >= 400 && output) {
                if (typeof output.error === 'string') {
                    errorMessage = output.error;
                } else if (output.error && typeof output.error.message === 'string') {
                    errorMessage = output.error.message;
                } else if (typeof output.message === 'string') {
                    errorMessage = output.message;
                } else if (typeof output.detail === 'string') {
                    errorMessage = output.detail;
                } else if (output.raw && typeof output.raw === 'string') {
                    errorMessage = output.raw.length > 200 ? output.raw.substring(0, 200) + '...' : output.raw;
                }
            }

            const span: Span = {
                id: entry.span_id || entry.id,
                name: `${entry.method} ${entry.path}`,
                type: 'tool', // Default type
                startTime: startTime,
                endTime: startTime + durationMs,
                status: entry.status >= 400 ? 'error' : 'success',
                input: input,
                output: output,
                errorMessage: errorMessage,
                children: [],
                serviceName: 'backend'
            };
            spanMap.set(span.id, span);
        });

        // Link spans
        let rootSpan: Span | null = null;

        spanMap.forEach(span => {
            const entry = group.find(e => (e.span_id || e.id) === span.id);
            if (entry && entry.parent_id && spanMap.has(entry.parent_id)) {
                const parent = spanMap.get(entry.parent_id)!;
                parent.children.push(span);
            } else {
                // Potential root
                if (!rootSpan) {
                    rootSpan = span;
                } else {
                    // If multiple roots, attach to the first one found to avoid orphan spans
                    // Ideally we should have a virtual root, but for now this works visually
                    rootSpan.children.push(span);
                }
            }
        });

        if (rootSpan) {
            // Calculate total duration (end of last span - start of root)
            const rootStart = rootSpan.startTime;
            let maxEnd = rootSpan.endTime;

            spanMap.forEach(s => {
                if (s.endTime > maxEnd) maxEnd = s.endTime;
            });

            traces.push({
                id: traceId,
                rootSpan: rootSpan,
                timestamp: group[0].timestamp, // Start time of trace
                totalDuration: maxEnd - rootStart,
                status: group.some(e => e.status >= 400) ? 'error' : 'success',
                trigger: 'user'
            });
        }
    }

    // Sort by timestamp descending
    traces.sort((a, b) => (a.timestamp > b.timestamp ? -1 : (a.timestamp < b.timestamp ? 1 : 0)));

    return NextResponse.json(traces);
  } catch (error) {
    console.error("Error fetching traces:", error);
    return NextResponse.json([]);
  }
}
