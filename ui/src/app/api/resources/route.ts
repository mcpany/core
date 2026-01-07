/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

let resources = [
    { name: "System Logs", uri: "file:///var/log/syslog", mimeType: "text/plain", disable: false, serviceId: "local-files", profiles: [], tags: [], mergeStrategy: 0, size: 0 },
    { name: "Project Config", uri: "file:///app/config.json", mimeType: "application/json", disable: false, serviceId: "local-files", profiles: [], tags: [], mergeStrategy: 0, size: 0 },
    { name: "Knowledge Base", uri: "postgres://db/knowledge", mimeType: "application/x-postgres", disable: false, serviceId: "memory-store", profiles: [], tags: [], mergeStrategy: 0, size: 0 },
];

export async function GET() {
  return NextResponse.json({ resources });
}

export async function POST(request: Request) {
    const body = await request.json();
    if (body.uri) {
        resources = resources.map(r => r.uri === body.uri ? { ...r, disable: body.disable } : r);
        return NextResponse.json({ message: "Updated" });
    }
    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
}
