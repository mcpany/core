/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

let prompts = [
    { name: "summarize_text", description: "Summarize a given text", disable: false, serviceId: "ai-helper", inputSchema: { type: "object", properties: { text: { type: "string" } }, required: ["text"] }, title: "Summarize Text", messages: [] },
    { name: "code_review", description: "Review code snippet", disable: false, serviceId: "dev-assistant", inputSchema: { type: "object", properties: { code: { type: "string" } }, required: ["code"] }, title: "Code Review", messages: [] },
    { name: "generate_sql", description: "Convert natural language to SQL", disable: true, serviceId: "db-assistant", inputSchema: { type: "object", properties: { query: { type: "string" } }, required: ["query"] }, title: "Generate SQL", messages: [] },
];

export async function GET() {
  return NextResponse.json({ prompts });
}

export async function POST(request: Request) {
    const body = await request.json();
    if (body.name) {
        prompts = prompts.map(p => p.name === body.name ? { ...p, disable: body.disable } : p);
        return NextResponse.json({ message: "Updated" });
    }
    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
}
