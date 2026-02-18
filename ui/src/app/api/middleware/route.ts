/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

/**
 * Handles GET requests to retrieve the list of configured middlewares from the backend.
 */
export async function GET(request: Request) {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:50050';

  try {
    const res = await fetch(`${backendUrl}/api/v1/middleware`, {
        headers: {
            'Authorization': request.headers.get('Authorization') || '',
            'X-API-Key': request.headers.get('X-API-Key') || process.env.MCPANY_API_KEY || ''
        },
        cache: 'no-store'
    });

    if (!res.ok) {
        console.warn(`Failed to fetch middleware from backend: ${res.status}`);
        return NextResponse.json([], { status: res.status });
    }

    const data = await res.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("Error connecting to backend for middleware:", error);
    return NextResponse.json([], { status: 500 });
  }
}

/**
 * Handles POST requests to update the list of configured middlewares.
 */
export async function POST(request: Request) {
    const backendUrl = process.env.BACKEND_URL || 'http://localhost:50050';

    try {
        const body = await request.json();
        const res = await fetch(`${backendUrl}/api/v1/middleware`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': request.headers.get('Authorization') || '',
                'X-API-Key': request.headers.get('X-API-Key') || process.env.MCPANY_API_KEY || ''
            },
            body: JSON.stringify(body)
        });

        if (!res.ok) {
            return NextResponse.json({ error: "Failed to update middleware" }, { status: res.status });
        }

        return NextResponse.json({});
    } catch (error) {
        console.error("Error connecting to backend for middleware update:", error);
        return NextResponse.json({ error: "Internal Server Error" }, { status: 500 });
    }
}
