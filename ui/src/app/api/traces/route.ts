/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

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
}

interface DebugEntry {
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
        console.error(`Failed to fetch traces from ${backendUrl}/debug/entries: ${res.status} ${res.statusText}`);
        return NextResponse.json([]);
    }

    const entries: DebugEntry[] = await res.json();

    // Validate if entries is array
    if (!Array.isArray(entries)) {
        console.error("Backend returned non-array for traces");
        return NextResponse.json([]);
    }

    const traces: Trace[] = entries.map(entry => {
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
            trigger: 'user'
        };
    });

    // Sort by timestamp descending
    traces.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());

    return NextResponse.json(traces);
  } catch (error) {
    console.error("Error fetching traces:", error);
    return NextResponse.json([]);
  }
}
