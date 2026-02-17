/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

/**
 * Handles GET requests to retrieve the list of configured middlewares.
 * Proxies to backend /api/v1/middleware.
 */
export async function GET(request: Request) {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080';
  const authHeader = request.headers.get('Authorization');
  const apiKey = request.headers.get('X-API-Key') || process.env.MCPANY_API_KEY;

  try {
    const headers: HeadersInit = {};
    if (authHeader) headers['Authorization'] = authHeader;
    if (apiKey) headers['X-API-Key'] = apiKey;

    const res = await fetch(`${backendUrl}/api/v1/middleware`, {
        headers: headers,
        cache: 'no-store'
    });

    if (!res.ok) {
        console.warn(`Failed to fetch middleware from backend: ${res.status}`);
        return NextResponse.json([], { status: res.status });
    }

    const data = await res.json();
    return NextResponse.json(data);
  } catch (error) {
      console.error("Error fetching middleware:", error);
      return NextResponse.json([], { status: 500 });
  }
}

/**
 * Handles POST requests to update the list of configured middlewares.
 * Proxies to backend /api/v1/middleware.
 */
export async function POST(request: Request) {
    const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080';
    const authHeader = request.headers.get('Authorization');
    const apiKey = request.headers.get('X-API-Key') || process.env.MCPANY_API_KEY;

    try {
      const body = await request.json();
      const headers: HeadersInit = { 'Content-Type': 'application/json' };
      if (authHeader) headers['Authorization'] = authHeader;
      if (apiKey) headers['X-API-Key'] = apiKey;

      const res = await fetch(`${backendUrl}/api/v1/middleware`, {
          method: 'POST',
          headers: headers,
          body: JSON.stringify(body)
      });

      if (!res.ok) {
          console.warn(`Failed to update middleware: ${res.status}`);
          return NextResponse.json({ error: "Failed to update" }, { status: res.status });
      }

      const data = await res.json();
      return NextResponse.json(data);
    } catch (error) {
        console.error("Error updating middleware:", error);
        return NextResponse.json({ error: "Internal Error" }, { status: 500 });
    }
}
