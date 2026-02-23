/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { Trace, Span } from '@/types/trace';

export type { SpanStatus, Span, Trace } from '@/types/trace';

interface AuditEntry {
  timestamp: string;
  tool_name: string;
  user_id?: string;
  profile_id?: string;
  trace_id?: string;
  span_id?: string;
  parent_id?: string;
  arguments: string; // JSON string
  result: string; // JSON string
  error?: string;
  duration: string;
  duration_ms: number;
}

/**
 * GET.
 *
 * @param request - The request.
 */
export async function GET(request: Request) {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:50059';

  try {
    const res = await fetch(`${backendUrl}/api/v1/audit/logs?limit=100`, {
        headers: {
            'Authorization': request.headers.get('Authorization') || '',
            'X-API-Key': request.headers.get('X-API-Key') || process.env.MCPANY_API_KEY || ''
        },
        cache: 'no-store'
    });

    if (!res.ok) {
        console.warn(`Failed to fetch traces from ${backendUrl}/api/v1/audit/logs: ${res.status} ${res.statusText}`);
        return NextResponse.json([]);
    }

    const data = await res.json();
    const entries: AuditEntry[] = data.entries || [];

    // Group entries by trace_id
    const traceGroups = new Map<string, AuditEntry[]>();
    entries.forEach(entry => {
        const traceId = entry.trace_id || entry.span_id || "unknown"; // Fallback
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
            const durationMs = entry.duration_ms;

            let input: Record<string, any> | undefined;
            try {
                input = JSON.parse(entry.arguments || "{}");
            } catch {
                input = { raw: entry.arguments };
            }

            let output: Record<string, any> | undefined;
            try {
                output = JSON.parse(entry.result || "{}");
            } catch {
                output = { raw: entry.result };
            }

            const span: Span = {
                id: entry.span_id || `span-${startTime}`,
                name: entry.tool_name,
                type: 'tool',
                startTime: startTime,
                endTime: startTime + durationMs,
                status: entry.error ? 'error' : 'success',
                input: input,
                output: output,
                errorMessage: entry.error,
                children: [],
                serviceName: 'backend'
            };
            spanMap.set(span.id, span);
        });

        // Link spans
        let rootSpan: Span | null = null;
        const roots: Span[] = [];

        spanMap.forEach(span => {
            const entry = group.find(e => (e.span_id || `span-${new Date(e.timestamp).getTime()}`) === span.id);
            if (entry && entry.parent_id && spanMap.has(entry.parent_id)) {
                const parent = spanMap.get(entry.parent_id)!;
                parent.children.push(span);
            } else {
                roots.push(span);
            }
        });

        // Handle multiple roots (orphans) or single root
        if (roots.length > 0) {
            // Sort roots by start time
            roots.sort((a, b) => a.startTime - b.startTime);

            // If multiple roots, create a virtual root if they belong to same trace
            if (roots.length > 1) {
                const first = roots[0];
                const last = roots[roots.length - 1]; // This logic assumes sequential execution for duration calculation, roughly

                // Better duration calc: max(endTime) - min(startTime)
                const traceStart = Math.min(...roots.map(r => r.startTime));
                const traceEnd = Math.max(...roots.map(r => r.endTime));

                 rootSpan = {
                    id: `virtual-${traceId}`,
                    name: `Trace ${traceId.substring(0,8)}`,
                    type: 'root',
                    startTime: traceStart,
                    endTime: traceEnd,
                    status: roots.some(r => r.status === 'error') ? 'error' : 'success',
                    children: roots,
                    serviceName: 'gateway'
                };
            } else {
                rootSpan = roots[0];
            }
        }

        if (rootSpan) {
            traces.push({
                id: traceId,
                rootSpan: rootSpan,
                timestamp: group[0].timestamp,
                totalDuration: rootSpan.endTime - rootSpan.startTime,
                status: rootSpan.status,
                trigger: group[0].user_id ? 'user' : 'system'
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
