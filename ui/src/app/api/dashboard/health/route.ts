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

/**
 * GET executes GET logic.
 *
 * @param request - ...
 * @returns ...
 */
export async function GET(request: Request) {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080';
  const authHeader = request.headers.get('Authorization');

  try {
    const headers: HeadersInit = {};
    if (authHeader) {
      headers['Authorization'] = authHeader;
    }

    const res = await fetch(`${backendUrl}/api/v1/services`, {
      cache: 'no-store', // Always fetch fresh data
      headers: headers
    });

    if (!res.ok) {
        console.warn(`Failed to fetch services from backend: ${res.status} ${res.statusText}`);
        return NextResponse.json({ error: "Failed to fetch service health" }, { status: res.status });
    }

    const data = await res.json();
    const servicesList: BackendService[] = Array.isArray(data) ? data : (data.services || []);

    const services = servicesList.map(svc => {
      let status = "healthy";
      if (svc.disable) {
        status = "inactive";
      } else if (svc.last_error || svc.config_error) {
        status = "unhealthy";
      }

      // We don't have real latency/uptime yet, so we'll leave them as placeholders or specific "Unknown" indicators
      // that the UI can handle gracefully.
      return {
        id: svc.id || svc.name,
        name: svc.name,
        status: status,
        latency: "--", // Placeholder until metrics are available
        uptime: "--",  // Placeholder
        message: svc.last_error || svc.config_error // Pass error message to UI
      };
    });

    return NextResponse.json(services);
  } catch (error) {
    console.error("Error connecting to backend for health check:", error);
    // Return empty list or error so the widget doesn't crash,
    // but maybe the widget shows an error state.
    return NextResponse.json([]);
  }
}
