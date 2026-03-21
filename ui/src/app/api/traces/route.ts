/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { Trace, Span } from '@/types/trace';

export type { SpanStatus, Span, Trace } from '@/types/trace';

interface HttpTraceEntry {
  id: string;
  timestamp: string;
  method: string;
  path: string;
  status: number;
  duration: number; // nanoseconds
  request_body?: string; // JSON string
  response_body?: string; // JSON string
}

/**
 * Transforms an HTTP trace entry into a Trace object.
 */
function transformEntry(entry: HttpTraceEntry): Trace | null {
    try {
        const startTime = new Date(entry.timestamp).getTime();
        // Convert nanoseconds to milliseconds
        const durationMs = Math.round(entry.duration / 1_000_000);
        const endTime = startTime + durationMs;

        let input: Record<string, any> | undefined;
        try {
            input = entry.request_body ? JSON.parse(entry.request_body) : undefined;
        } catch {
            input = entry.request_body ? { raw: entry.request_body } : undefined;
        }

        let output: Record<string, any> | undefined;
        let errorMessage: string | undefined;
        try {
            if (entry.response_body) {
                const parsed = JSON.parse(entry.response_body);
                output = parsed;
                // Extract error message from response body if status indicates error
                if (entry.status >= 400 && parsed?.error?.message) {
                    errorMessage = parsed.error.message;
                }
            }
        } catch {
            output = entry.response_body ? { raw: entry.response_body } : undefined;
        }

        const isError = entry.status >= 400;
        const spanStatus: 'success' | 'error' = isError ? 'error' : 'success';

        const rootSpan: Span = {
            id: `span-${entry.id}`,
            name: `${entry.method} ${entry.path}`,
            type: 'tool',
            startTime,
            endTime,
            status: spanStatus,
            input,
            output,
            errorMessage,
            children: [],
        };

        return {
            id: entry.id,
            rootSpan,
            timestamp: entry.timestamp,
            totalDuration: durationMs,
            status: spanStatus,
            trigger: 'user',
        };
    } catch {
        return null;
    }
}

/**
 * GET.
 *
 * @param request - The request.
 */
export async function GET(request: Request) {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:50050';

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
    // Support both direct array and { entries: [...] } response formats
    const entries: HttpTraceEntry[] = Array.isArray(data) ? data : (data.entries || []);

    const traces: Trace[] = entries
        .map(transformEntry)
        .filter((t): t is Trace => t !== null);

    // Sort by timestamp descending
    traces.sort((a, b) => (a.timestamp > b.timestamp ? -1 : (a.timestamp < b.timestamp ? 1 : 0)));

    return NextResponse.json(traces);
  } catch (error) {
    console.error("Error fetching traces:", error);
    return NextResponse.json([]);
  }
}
