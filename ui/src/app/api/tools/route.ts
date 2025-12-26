/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

let tools = [
    { name: "get_weather", description: "Get weather for a location", serviceName: "weather-service", enabled: true },
    { name: "list_files", description: "List files in directory", serviceName: "local-files", enabled: true },
    { name: "read_file", description: "Read file content", serviceName: "local-files", enabled: false },
    { name: "save_memory", description: "Save a key-value pair", serviceName: "memory-store", enabled: true },
];

export async function GET() {
  return NextResponse.json({ tools });
}

export async function POST(request: Request) {
    const body = await request.json();
    if (body.name && typeof body.enabled === 'boolean') {
        tools = tools.map(t => t.name === body.name ? { ...t, enabled: body.enabled } : t);
        return NextResponse.json({ message: "Updated" });
    }
    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
}
