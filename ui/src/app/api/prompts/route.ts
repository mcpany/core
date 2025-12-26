/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

let prompts = [
    { name: "summarize_file", description: "Summarize the content of a file", serviceName: "local-files", enabled: true },
    { name: "weather_report", description: "Generate a weather report", serviceName: "weather-service", enabled: false },
];

export async function GET() {
  return NextResponse.json({ prompts });
}

export async function POST(request: Request) {
    const body = await request.json();
    if (body.name && typeof body.enabled === 'boolean') {
        prompts = prompts.map(p => p.name === body.name ? { ...p, enabled: body.enabled } : p);
        return NextResponse.json({ message: "Updated" });
    }
    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
}
