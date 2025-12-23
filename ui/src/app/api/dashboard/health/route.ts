/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { NextResponse } from 'next/server';

export async function GET() {
  // Simulate backend data
  const services = [
    {
      name: "Payment Gateway",
      version: "v1.2.0",
      uptime: "99.99%",
      status: "healthy",
    },
    {
      name: "User Service",
      version: "v2.1.0",
      uptime: "99.95%",
      status: "healthy",
    },
    {
      name: "Notification Service",
      version: "v1.0.1",
      uptime: "98.50%",
      status: "degraded",
    },
    {
      name: "Search Indexer",
      version: "v0.9.0",
      uptime: "85.00%",
      status: "unhealthy",
    },
    {
      name: "Analytics Engine",
      version: "v3.0.0",
      uptime: "99.90%",
      status: "healthy",
    },
  ];

  return NextResponse.json(services);
}
