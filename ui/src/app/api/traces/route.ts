/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { Trace, Span } from '@/types/trace';

export type { SpanStatus, Span, Trace } from '@/types/trace';

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
        console.warn(`Failed to fetch traces from ${backendUrl}/debug/entries: ${res.status} ${res.statusText}`);
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
            // Try parsing SSE (event: message\ndata: {...})
            const sseMatch = entry.response_body.match(/data: ({.*})/);
            if (sseMatch && sseMatch[1]) {
                try {
                    output = JSON.parse(sseMatch[1]);
                } catch {
                    output = { raw: entry.response_body };
                }
            } else {
                output = { raw: entry.response_body };
            }
        }

        let errorMessage: string | undefined;
        if (entry.status >= 400) {
            // Check for HTTP error
             errorMessage = `HTTP ${entry.status}`;
        }

        // Check for JSON-RPC error in output
        if (output && output.error) {
             if (typeof output.error === 'string') {
                errorMessage = output.error;
            } else if (output.error && typeof output.error.message === 'string') {
                errorMessage = output.error.message;
            }
        } else if (output && typeof output.message === 'string' && entry.status >= 400) {
             // Fallback for non-standard error responses
             errorMessage = output.message;
        }

        // Detect JSON-RPC tool call
        let isToolCall = false;
        let toolName = "";
        // Check for standard JSON-RPC request format
        if (input && typeof input === 'object' && input.method === "tools/call" && input.params && typeof input.params === 'object' && input.params.name) {
             isToolCall = true;
             toolName = input.params.name;
        }

        // Base root span
        const rootSpan: Span = {
            id: entry.id,
            name: `${entry.method} ${entry.path}`,
            type: 'core',
            startTime: startTime,
            endTime: startTime + durationMs,
            status: (entry.status >= 400 || errorMessage) ? 'error' : 'success',
            input: input,
            output: output,
            errorMessage: errorMessage,
            children: [],
            serviceName: 'mcp-any'
        };

        if (isToolCall) {
            rootSpan.name = "Execute Request";

            // Calculate timing for the child span (tool execution)
            // We assume a small overhead for the core processing
            const overhead = Math.min(durationMs * 0.1, 20); // 10% or 20ms
            const toolStartTime = startTime + (overhead / 2);
            const toolEndTime = startTime + durationMs - (overhead / 2);

            const toolSpan: Span = {
                id: `${entry.id}-tool`,
                name: toolName,
                type: 'tool',
                startTime: toolStartTime,
                endTime: toolEndTime,
                status: rootSpan.status,
                input: input?.params?.arguments || {},
                output: output?.result || output, // Use result if available (standard MCP), else raw output
                errorMessage: errorMessage,
                children: [],
                serviceName: 'upstream'
            };

            rootSpan.children = [toolSpan];
        }

        return {
            id: entry.id,
            rootSpan: rootSpan,
            timestamp: entry.timestamp,
            totalDuration: durationMs,
            status: rootSpan.status,
            trigger: 'user'
        };
    });

    // Sort by timestamp descending
    traces.sort((a, b) => (a.timestamp > b.timestamp ? -1 : (a.timestamp < b.timestamp ? 1 : 0)));

    return NextResponse.json(traces);
  } catch (error) {
    console.error("Error fetching traces:", error);
    return NextResponse.json([]);
  }
}
