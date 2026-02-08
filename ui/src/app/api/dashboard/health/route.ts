/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

export async function GET(request: Request) {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:50050';
  const authHeader = request.headers.get('Authorization');

  try {
    const headers: HeadersInit = {};
    if (authHeader) {
      headers['Authorization'] = authHeader;
    } else {
      const apiKey = process.env.MCPANY_API_KEY || process.env.NEXT_PUBLIC_MCPANY_API_KEY;
      if (apiKey) {
        headers['X-API-Key'] = apiKey;
      }
    }

    // Proxy directly to backend dashboard health API
    const res = await fetch(`${backendUrl}/api/v1/dashboard/health`, {
      cache: 'no-store',
      headers: headers
    });

    if (!res.ok) {
        console.warn(`Failed to fetch health from backend: ${res.status} ${res.statusText}`);
        // Debug auth issues
        if (res.status === 401) {
             const keyUsed = headers['X-API-Key'] ? 'SET' : 'MISSING';
             console.error(`Auth failed. MCPANY_API_KEY: ${!!process.env.MCPANY_API_KEY}, NEXT_PUBLIC: ${!!process.env.NEXT_PUBLIC_MCPANY_API_KEY}, Header: ${keyUsed}`);
        }
        return NextResponse.json({ services: [], history: {} }, { status: res.status });
    }

    const data = await res.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("Error connecting to backend for health check:", error);
    return NextResponse.json({ services: [], history: {} });
  }
}
