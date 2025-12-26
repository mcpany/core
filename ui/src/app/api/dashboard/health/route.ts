/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

export async function GET() {
  const services = [
    {
      id: "srv-001",
      name: "Core API Gateway",
      status: "healthy",
      latency: "24ms",
      uptime: "99.99%",
    },
    {
      id: "srv-002",
      name: "Authentication Service",
      status: "healthy",
      latency: "45ms",
      uptime: "99.95%",
    },
    {
      id: "srv-003",
      name: "Vector Database",
      status: "degraded",
      latency: "150ms",
      uptime: "99.00%",
    },
    {
      id: "srv-004",
      name: "LLM Orchestrator",
      status: "healthy",
      latency: "320ms",
      uptime: "99.99%",
    },
     {
      id: "srv-005",
      name: "Webhook Processor",
      status: "healthy",
      latency: "12ms",
      uptime: "100%",
    },
  ];

  return NextResponse.json(services);
}
