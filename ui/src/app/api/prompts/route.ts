/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

let prompts = [
    { name: "summarize_text", description: "Summarize a given text", enabled: true, serviceName: "ai-helper", arguments: [{ name: "text", required: true }] },
    { name: "code_review", description: "Review code snippet", enabled: true, serviceName: "dev-assistant", arguments: [{ name: "code", required: true }] },
    { name: "generate_sql", description: "Convert natural language to SQL", enabled: false, serviceName: "db-assistant", arguments: [{ name: "query", required: true }] },
];

export async function GET() {
  return NextResponse.json({ prompts });
}

export async function POST(request: Request) {
    const body = await request.json();
    if (body.name) {
        prompts = prompts.map(p => p.name === body.name ? { ...p, enabled: body.enabled } : p);
        return NextResponse.json({ message: "Updated" });
    }
    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
}
