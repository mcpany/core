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
  latency_ms?: number;
  last_check?: string;
}

export async function GET() {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080';

  try {
    const res = await fetch(`${backendUrl}/api/v1/services`, {
      cache: 'no-store', // Always fetch fresh data
      headers: {
        'X-API-Key': process.env.MCPANY_API_KEY || ''
      }
    });

    if (!res.ok) {
        console.error(`Failed to fetch services from backend: ${res.status} ${res.statusText}`);
        return NextResponse.json({ error: "Failed to fetch service health" }, { status: 500 });
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

      const latency = svc.latency_ms ? `${svc.latency_ms}ms` : "--";

      // Calculate a simple "last seen" or "status" string for uptime column for now
      let uptime = "--";
      if (svc.last_check) {
          const lastCheck = new Date(svc.last_check);
          const now = new Date();
          const diffSeconds = Math.floor((now.getTime() - lastCheck.getTime()) / 1000);
          if (diffSeconds < 60) {
              uptime = "Just now";
          } else if (diffSeconds < 3600) {
              uptime = `${Math.floor(diffSeconds / 60)}m ago`;
          } else {
              uptime = "Active";
          }
      }

      return {
        id: svc.id || svc.name,
        name: svc.name,
        status: status,
        latency: latency,
        uptime: uptime,
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
