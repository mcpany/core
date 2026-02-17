/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

/**
 * Handles GET requests to retrieve the list of configured middlewares.
 */
export async function GET(request: Request) {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080';
  const authHeader = request.headers.get('Authorization');

  try {
    const headers: HeadersInit = {};
    if (authHeader) {
      headers['Authorization'] = authHeader;
    }
    if (process.env.MCPANY_API_KEY) {
      headers['X-API-Key'] = process.env.MCPANY_API_KEY;
    }

    const res = await fetch(`${backendUrl}/api/v1/middleware`, {
      cache: 'no-store',
      headers: headers
    });

    if (!res.ok) {
        console.warn(`Failed to fetch middleware from backend: ${res.status} ${res.statusText}`);
        return NextResponse.json({ error: "Failed to fetch middleware" }, { status: res.status });
    }

    const data = await res.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("Error connecting to backend for middleware:", error);
    return NextResponse.json([]);
  }
}

/**
 * Handles POST requests to update the list of configured middlewares.
 */
export async function POST(request: Request) {
    const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080';
    const authHeader = request.headers.get('Authorization');

    try {
        const body = await request.json();
        const headers: HeadersInit = {
            'Content-Type': 'application/json'
        };
        if (authHeader) {
            headers['Authorization'] = authHeader;
        }
        if (process.env.MCPANY_API_KEY) {
            headers['X-API-Key'] = process.env.MCPANY_API_KEY;
        }

        const res = await fetch(`${backendUrl}/api/v1/middleware`, {
            method: 'POST',
            headers: headers,
            body: JSON.stringify(body)
        });

        if (!res.ok) {
            console.warn(`Failed to update middleware in backend: ${res.status} ${res.statusText}`);
            return NextResponse.json({ error: "Failed to update middleware" }, { status: res.status });
        }

        return NextResponse.json({ success: true });
    } catch (error) {
        console.error("Error connecting to backend for middleware update:", error);
        return NextResponse.json({ error: "Connection failed" }, { status: 500 });
    }
}
