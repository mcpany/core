/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { Trace, Span } from '@/types/trace';

export type { SpanStatus, Span, Trace } from '@/types/trace';

/**
 * GET.
 *
 * @param request - The request.
 */
export async function GET(request: Request) {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:50059';

  try {
    // Forward query params (limit)
    const { searchParams } = new URL(request.url);
    const limit = searchParams.get('limit') || '';

    const res = await fetch(`${backendUrl}/api/v1/traces?limit=${limit}`, {
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

    return NextResponse.json(traces);
  } catch (error) {
    console.error("Error fetching traces:", error);
    return NextResponse.json([]);
  }
}
