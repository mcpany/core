/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { Trace, Span, SpanStatus } from '@/types/trace';

export type { SpanStatus, Span, Trace } from '@/types/trace';

interface DebugEntry {
  id: string;
  trace_id: string;
  span_id: string;
  parent_span_id?: string;
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

    // Group by TraceID
    const spansByTrace = new Map<string, Span[]>();
    // Map to keep track of parent IDs for reconstruction
    const parentMap = new Map<string, string | undefined>();

    for (const entry of entries) {
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
                // Truncate raw body if it's too long
                errorMessage = output.raw.length > 200 ? output.raw.substring(0, 200) + '...' : output.raw;
            }
        }

        const spanId = entry.span_id || entry.id;

        const span: Span = {
            id: spanId,
            name: `${entry.method} ${entry.path}`,
            type: 'tool', // Assume tool call for now
            startTime: startTime,
            endTime: startTime + durationMs,
            status: entry.status >= 400 ? 'error' : 'success',
            input: input,
            output: output,
            errorMessage: errorMessage,
            children: [],
            serviceName: 'backend'
        };

        const traceId = entry.trace_id || entry.id;

        if (!spansByTrace.has(traceId)) {
            spansByTrace.set(traceId, []);
        }
        spansByTrace.get(traceId)!.push(span);
        parentMap.set(spanId, entry.parent_span_id);
    }

    const traces: Trace[] = [];

    // Reconstruct trees
    for (const [traceId, spans] of spansByTrace) {
         const spanMap = new Map<string, Span>();
         spans.forEach(s => spanMap.set(s.id, s));

         const roots: Span[] = [];

         spans.forEach(s => {
             const parentId = parentMap.get(s.id);

             if (parentId && spanMap.has(parentId)) {
                 const parent = spanMap.get(parentId)!;
                 if (!parent.children) parent.children = [];
                 parent.children.push(s);
             } else {
                 roots.push(s);
             }
         });

         // Handle multiple roots (e.g. parallel requests without a common captured parent)
         roots.sort((a, b) => a.startTime - b.startTime);
         let root = roots[0];

         if (!root) continue;

         // If multiple roots exist, create a virtual root to hold them all
         if (roots.length > 1) {
             let minStart = roots[0].startTime;
             let maxEnd = roots[0].endTime;
             let hasError = false;

             roots.forEach(r => {
                 if (r.startTime < minStart) minStart = r.startTime;
                 if (r.endTime > maxEnd) maxEnd = r.endTime;
                 if (r.status === 'error') hasError = true;
             });

             root = {
                 id: `virtual-root-${traceId}`,
                 name: "Trace Group",
                 type: 'core',
                 startTime: minStart,
                 endTime: maxEnd,
                 status: hasError ? 'error' : 'success',
                 children: roots,
                 serviceName: 'virtual'
             };
         }

         // Calculate total duration
         let minStart = root.startTime;
         let maxEnd = root.endTime;
         let hasError = root.status === 'error';

         const traverse = (s: Span) => {
             if (s.startTime < minStart) minStart = s.startTime;
             if (s.endTime > maxEnd) maxEnd = s.endTime;
             if (s.status === 'error') hasError = true;
             s.children?.forEach(traverse);
         }
         traverse(root);

         traces.push({
             id: traceId,
             rootSpan: root,
             timestamp: new Date(minStart).toISOString(),
             totalDuration: maxEnd - minStart,
             status: hasError ? 'error' : 'success',
             trigger: 'user'
         });
    }

    // Sort traces by timestamp descending
    traces.sort((a, b) => (a.timestamp > b.timestamp ? -1 : (a.timestamp < b.timestamp ? 1 : 0)));

    return NextResponse.json(traces);
  } catch (error) {
    console.error("Error fetching traces:", error);
    return NextResponse.json([]);
  }
}
