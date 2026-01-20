/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { DebugEntry, mapEntryToTrace } from '@/lib/trace-utils';

/**
 * GET single trace by ID.
 *
 * @param request - The request.
 * @param params - The params.
 */
export async function GET(request: Request, { params }: { params: { id: string } }) {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:50059';
  const id = params.id;

  try {
    const url = `${backendUrl}/debug/entries?id=${id}`;

    const res = await fetch(url, {
        headers: {
            'Authorization': request.headers.get('Authorization') || '',
            'X-API-Key': request.headers.get('X-API-Key') || process.env.MCPANY_API_KEY || ''
        },
        cache: 'no-store'
    });

    if (!res.ok) {
        if (res.status === 404) {
            return new NextResponse(null, { status: 404 });
        }
        console.error(`Failed to fetch trace from ${url}: ${res.status} ${res.statusText}`);
        return new NextResponse(null, { status: res.status });
    }

    const entry: DebugEntry = await res.json();
    const trace = mapEntryToTrace(entry);

    return NextResponse.json(trace);
  } catch (error) {
    console.error("Error fetching trace:", error);
    return new NextResponse(null, { status: 500 });
  }
}
