/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { NextResponse } from 'next/server';

export async function GET() {
  const tools = [
    { name: "calculator", description: "Performs basic arithmetic", source: "discovered", service: "MathService" },
    { name: "weather_lookup", description: "Gets weather for a location", source: "configured", service: "WeatherService" },
    { name: "db_query", description: "Executes SQL queries", source: "configured", service: "DBService" },
  ];
  return NextResponse.json(tools);
}
