/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

/**
 * GET proxies the request to the backend /api/v1/dashboard/tool-usage endpoint.
 *
 * @param request - The incoming Request.
 * @returns A NextResponse containing the tool usage stats.
 */
export async function GET(request: Request) {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080';
  const { searchParams } = new URL(request.url);
  const serviceId = searchParams.get('serviceId');

  try {
    const headers: HeadersInit = {};
    const authHeader = request.headers.get('Authorization');
    if (authHeader) headers['Authorization'] = authHeader;
    if (process.env.MCPANY_API_KEY) headers['X-API-Key'] = process.env.MCPANY_API_KEY;

    let url = `${backendUrl}/api/v1/dashboard/tool-usage`;
    if (serviceId) {
        url += `?serviceId=${encodeURIComponent(serviceId)}`;
    }

    const res = await fetch(url, {
      headers: headers,
      cache: 'no-store'
    });

    if (!res.ok) {
        return NextResponse.json({ error: "Failed to fetch tool usage" }, { status: res.status });
    }

    const data = await res.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("Error fetching tool usage:", error);
    return NextResponse.json([], { status: 500 });
  }
}
