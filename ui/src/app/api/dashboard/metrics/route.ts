/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { NextResponse } from 'next/server';

export async function GET() {
  // Mock data for dashboard metrics
  // In a real scenario, this would aggregate data from the backend

  // Randomize slightly to simulate real-time updates
  const randomize = (base: number, variance: number) => {
    return Math.floor(base + (Math.random() - 0.5) * variance);
  };

  const metrics = [
    {
      label: "Total Requests",
      value: randomize(2345, 100).toLocaleString(),
      change: "+20.1%",
      trend: "up",
      icon: "Activity",
      subLabel: "req/sec",
    },
    {
      label: "Active Services",
      value: "12",
      change: "+2",
      trend: "up",
      icon: "Server",
      subLabel: "Healthy",
    },
    {
      label: "Connected Tools",
      value: "573",
      change: "+201",
      trend: "up",
      icon: "Zap",
      subLabel: "Available",
    },
    {
      label: "Resources",
      value: "1,204",
      change: "+5%",
      trend: "up",
      icon: "Database",
      subLabel: "Managed",
    },
    {
      label: "Prompts",
      value: "89",
      change: "+12",
      trend: "up",
      icon: "MessageSquare",
      subLabel: "Templates",
    },
    {
      label: "Avg Latency",
      value: `${randomize(45, 10)}ms`,
      change: "-5ms",
      trend: "down", // Good thing
      icon: "Clock",
      subLabel: "Global Avg",
    },
    {
      label: "Error Rate",
      value: "0.02%",
      change: "-0.01%",
      trend: "down", // Good thing
      icon: "AlertCircle",
      subLabel: "Last 24h",
    },
  ];

  return NextResponse.json(metrics);
}
