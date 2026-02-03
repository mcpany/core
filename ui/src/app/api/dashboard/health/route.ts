/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

interface BackendService {
  id: string;
  name: string;
  disable: boolean;
  last_error?: string;
  config_error?: string;
  // Other fields we might use later
}

export async function GET(request: Request) {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080';
  const authHeader = request.headers.get('Authorization');
  const apiKeyHeader = request.headers.get('X-API-Key');

  try {
    const headers: HeadersInit = {};
    if (authHeader) {
      headers['Authorization'] = authHeader;
    }
    if (apiKeyHeader) {
      headers['X-API-Key'] = apiKeyHeader;
    }

    // Use the dedicated dashboard health endpoint which returns richer data
    const res = await fetch(`${backendUrl}/api/v1/dashboard/health`, {
      cache: 'no-store',
      headers: headers
    });

    if (!res.ok) {
        console.warn(`Failed to fetch dashboard health from backend: ${res.status} ${res.statusText}`);
        return NextResponse.json({ error: "Failed to fetch service health" }, { status: res.status });
    }

    const data = await res.json();
    // Pass through the data directly as it matches the expected structure
    return NextResponse.json(data);
  } catch (error) {
    console.error("Error connecting to backend for health check:", error);
    // Return empty list or error so the widget doesn't crash,
    // but maybe the widget shows an error state.
    return NextResponse.json([]);
  }
}
