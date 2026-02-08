/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

export async function GET(request: Request) {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080';
  const authHeader = request.headers.get('Authorization');
  const apiKey = request.headers.get('X-API-Key');

  try {
    const headers: HeadersInit = {};
    if (authHeader) {
      headers['Authorization'] = authHeader;
    }
    if (apiKey) {
      headers['X-API-Key'] = apiKey;
    }

    const res = await fetch(`${backendUrl}/api/v1/dashboard/health`, {
      cache: 'no-store', // Always fetch fresh data
      headers: headers
    });

    if (!res.ok) {
        console.warn(`Failed to fetch service health from backend: ${res.status} ${res.statusText}`);
        return NextResponse.json({ error: "Failed to fetch service health" }, { status: res.status });
    }

    // Backend now returns the correct structure { services: [...], history: {...} }
    const data = await res.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("Error connecting to backend for health check:", error);
    // Return empty list or error so the widget doesn't crash,
    // but maybe the widget shows an error state.
    return NextResponse.json([]);
  }
}
