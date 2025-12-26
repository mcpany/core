/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { NextResponse } from 'next/server';

export async function GET() {
  // Mock data for service health
  const health = [
    {
      id: "svc-1",
      name: "Payment Service",
      status: "healthy",
      latency: "45ms",
      uptime: "99.9%",
    },
    {
      id: "svc-2",
      name: "Auth Service",
      status: "degraded",
      latency: "150ms",
      uptime: "99.5%",
    },
    {
      id: "svc-3",
      name: "Notification Service",
      status: "healthy",
      latency: "32ms",
      uptime: "99.99%",
    },
    {
        id: "svc-4",
        name: "Data Processor",
        status: "unhealthy",
        latency: "Timeout",
        uptime: "95.0%",
    }
  ];

  return NextResponse.json(health);
}
