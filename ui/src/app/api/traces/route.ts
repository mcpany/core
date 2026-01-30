/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { Trace } from '@/types/trace';

export async function GET(request: Request) {
  // Default to 50050 which is the standard port
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:50050';

  try {
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
    return NextResponse.json(traces);
  } catch (error) {
    console.error("Error fetching traces:", error);
    return NextResponse.json([]);
  }
}
