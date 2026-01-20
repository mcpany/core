/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { DebugEntry, mapEntryToTrace, Trace } from '@/lib/trace-utils';

export type { Trace, Span, SpanStatus } from '@/lib/trace-utils';

/**
 * GET.
 *
 * @param request - The request.
 */
export async function GET(request: Request) {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:50059';
  const { searchParams } = new URL(request.url);
  const summary = searchParams.get('summary');

  try {
    const url = `${backendUrl}/debug/entries` + (summary === 'true' ? '?summary=true' : '');

    const res = await fetch(url, {
        headers: {
            'Authorization': request.headers.get('Authorization') || '',
            'X-API-Key': request.headers.get('X-API-Key') || process.env.MCPANY_API_KEY || ''
        },
        cache: 'no-store'
    });

    if (!res.ok) {
        console.error(`Failed to fetch traces from ${url}: ${res.status} ${res.statusText}`);
        return NextResponse.json([]);
    }

    const entries: DebugEntry[] = await res.json();

    // Validate if entries is array
    if (!Array.isArray(entries)) {
        console.error("Backend returned non-array for traces");
        return NextResponse.json([]);
    }

    const traces: Trace[] = entries.map(e => mapEntryToTrace(e, summary === 'true'));

    // Sort by timestamp descending
    traces.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());

    return NextResponse.json(traces);
  } catch (error) {
    console.error("Error fetching traces:", error);
    return NextResponse.json([]);
  }
}
