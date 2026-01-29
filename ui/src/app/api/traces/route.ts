/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { Trace } from '@/types/trace';

export type { SpanStatus, Span, Trace } from '@/types/trace';

/**
 * GET.
 *
 * @param request - The request.
 */
export async function GET(request: Request) {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:50059';

  try {
    // Switch to persistent audit trace endpoint
    const res = await fetch(`${backendUrl}/api/v1/traces`, {
        headers: {
            'Authorization': request.headers.get('Authorization') || '',
            'X-API-Key': request.headers.get('X-API-Key') || process.env.MCPANY_API_KEY || ''
        },
        cache: 'no-store'
    });

    if (!res.ok) {
        console.warn(`Failed to fetch traces from ${backendUrl}/api/v1/traces: ${res.status} ${res.statusText}`);
        return NextResponse.json([]);
    }

    const traces: Trace[] = await res.json();

    // Validate if traces is array
    if (!Array.isArray(traces)) {
        console.error("Backend returned non-array for traces");
        return NextResponse.json([]);
    }

    // Sort by timestamp descending (newest first)
    // Optimization: Compare strings directly instead of creating Date objects.
    traces.sort((a, b) => (a.timestamp > b.timestamp ? -1 : (a.timestamp < b.timestamp ? 1 : 0)));

    return NextResponse.json(traces);
  } catch (error) {
    console.error("Error fetching traces:", error);
    return NextResponse.json([]);
  }
}
