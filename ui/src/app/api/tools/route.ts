/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

import { MockDB } from '@/lib/server/mock-db';

export async function GET() {
  return NextResponse.json({ tools: MockDB.tools });
}

export async function POST(request: Request) {
    const body = await request.json();
    if (body.name) {
        MockDB.tools = MockDB.tools.map(t => t.name === body.name ? { ...t, enabled: body.enabled } : t);
        return NextResponse.json({ message: "Updated" });
    }
    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
}
