/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

let tools = [
    { name: "get_weather", description: "Get current weather for a location", enabled: true, serviceName: "weather-service", schema: { type: "object", properties: { location: { type: "string" } } } },
    { name: "read_file", description: "Read file from filesystem", enabled: true, serviceName: "local-files", schema: { type: "object", properties: { path: { type: "string" } } } },
    { name: "list_directory", description: "List directory contents", enabled: false, serviceName: "local-files", schema: { type: "object", properties: { path: { type: "string" } } } },
    { name: "search_memory", description: "Search vector memory", enabled: true, serviceName: "memory-store", schema: { type: "object", properties: { query: { type: "string" } } } },
];

export async function GET() {
  return NextResponse.json({ tools });
}

export async function POST(request: Request) {
    const body = await request.json();
    if (body.name) {
        tools = tools.map(t => t.name === body.name ? { ...t, enabled: body.enabled } : t);
        return NextResponse.json({ message: "Updated" });
    }
    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
}
