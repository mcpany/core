/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

let services = [
    { name: "weather-service", version: "1.2.0", disable: false, http_service: { address: "http://weather:8080" }, id: "srv-1" },
    { name: "memory-store", version: "0.9.5", disable: true, grpc_service: { address: "memory:9090" }, id: "srv-2" },
    { name: "local-files", version: "1.0.0", disable: false, command_line_service: { command: "npx", args: ["-y", "@modelcontextprotocol/server-filesystem", "/users/me/docs"] }, id: "srv-3" },
];

export async function GET() {
  return NextResponse.json({ services });
}

export async function POST(request: Request) {
    const body = await request.json();

    // Toggle status
    if (body.action === 'toggle' && body.name) {
        services = services.map(s => s.name === body.name ? { ...s, disable: body.disable } : s);
        return NextResponse.json({ message: "Updated" });
    }

    // Register/Update
    if (body.name) {
        const existing = services.find(s => s.name === body.name);
        if (existing) {
             services = services.map(s => s.name === body.name ? { ...body, id: s.id } : s);
        } else {
            services.push({ ...body, id: `srv-${Date.now()}` });
        }
        return NextResponse.json({ message: "Saved" });
    }

    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
}
