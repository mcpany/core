/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

export async function GET(request: Request) {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080';
  const authHeader = request.headers.get('Authorization');

  try {
    const headers: HeadersInit = {};
    if (authHeader) {
      headers['Authorization'] = authHeader;
    }

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
