/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { NextResponse } from 'next/server';

export async function GET() {
  // Simulate backend data
  const metrics = [
    {
      label: "Total Requests",
      value: "1.2M",
      change: "+12.5%",
      trend: "up",
      icon: "Activity",
    },
    {
      label: "Active Services",
      value: "14",
      change: "+2",
      trend: "up",
      icon: "Server",
    },
    {
      label: "Avg Latency",
      value: "45ms",
      change: "-5ms",
      trend: "down",
      icon: "Zap",
    },
    {
      label: "Active Users",
      value: "573",
      change: "+201",
      trend: "up",
      icon: "Users",
    },
  ];

  return NextResponse.json(metrics);
}
