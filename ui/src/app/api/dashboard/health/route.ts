/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

interface ServiceHealth {
  id: string;
  name: string;
  status: string;
  latency: string;
  uptime: string;
  message?: string;
}

interface HistoryPoint {
  timestamp: number;
  status: string;
  latency_ms: number;
}

interface BackendHealthResponse {
  services: ServiceHealth[];
  history: Record<string, HistoryPoint[]>;
}

/**
 * GET retrieves the health status of upstream services.
 *
 * It queries the backend API for service status and history.
 *
 * @param request - The incoming Request.
 * @returns A NextResponse containing the service health object with services list and history map.
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

    // Call the dedicated dashboard health endpoint
    const res = await fetch(`${backendUrl}/api/v1/dashboard/health`, {
      cache: 'no-store', // Always fetch fresh data
      headers: headers
    });

    if (!res.ok) {
        console.warn(`Failed to fetch health from backend: ${res.status} ${res.statusText}`);
        return NextResponse.json({ error: "Failed to fetch service health" }, { status: res.status });
    }

    const data: BackendHealthResponse = await res.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("Error connecting to backend for health check:", error);
    return NextResponse.json({ services: [], history: {} });
  }
}
