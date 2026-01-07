/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

let tools = [
    { name: "get_weather", description: "Get current weather for a location", disable: false, serviceId: "weather-service", inputSchema: { type: "object", properties: { location: { type: "string" } } } },
    { name: "read_file", description: "Read file from filesystem", disable: false, serviceId: "local-files", inputSchema: { type: "object", properties: { path: { type: "string" } } } },
    { name: "list_directory", description: "List directory contents", disable: true, serviceId: "local-files", inputSchema: { type: "object", properties: { path: { type: "string" } } } },
    { name: "search_memory", description: "Search vector memory", disable: false, serviceId: "memory-store", inputSchema: { type: "object", properties: { query: { type: "string" } } } },
];

export async function GET() {
  return NextResponse.json({ tools });
}

export async function POST(request: Request) {
    const body = await request.json();
    if (body.name) {
        // Toggle tool status (mock)
        const toolIndex = tools.findIndex(t => t.name === body.name);
        if (toolIndex !== -1) {
             tools[toolIndex] = { ...tools[toolIndex], disable: body.disable };
        }
        tools = tools.map(t => t.name === body.name ? { ...t, disable: body.disable } : t);
    }
    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
}
