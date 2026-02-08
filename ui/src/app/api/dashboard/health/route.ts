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
    } else if (process.env.MCPANY_API_KEY) {
      headers['X-API-Key'] = process.env.MCPANY_API_KEY;
    }

    // Proxy directly to backend dashboard health API
    const res = await fetch(`${backendUrl}/api/v1/dashboard/health`, {
      cache: 'no-store',
      headers: headers
    });

    if (!res.ok) {
        console.warn(`Failed to fetch health from backend: ${res.status} ${res.statusText}`);
        return NextResponse.json({ services: [], history: {} }, { status: res.status });
    }

    const data = await res.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("Error connecting to backend for health check:", error);
    return NextResponse.json({ services: [], history: {} });
  }
}
