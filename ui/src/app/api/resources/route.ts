/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

let resources = [
    { uri: "file:///users/me/docs/notes.txt", name: "notes.txt", mimeType: "text/plain", serviceName: "local-files", enabled: true },
    { uri: "weather://current/sf", name: "SF Weather", mimeType: "application/json", serviceName: "weather-service", enabled: true },
];

export async function GET() {
  return NextResponse.json({ resources });
}

export async function POST(request: Request) {
    const body = await request.json();
    if (body.uri && typeof body.enabled === 'boolean') {
        resources = resources.map(r => r.uri === body.uri ? { ...r, enabled: body.enabled } : r);
        return NextResponse.json({ message: "Updated" });
    }
    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
}
