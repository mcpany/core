/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { MockDB } from '@/lib/server/mock-db';

export async function GET() {
  return NextResponse.json({ services: MockDB.services });
}

export async function POST(request: Request) {
    let body;
    try {
        body = await request.json();
    } catch (e) {
        return NextResponse.json({ error: "Invalid request body" }, { status: 400 });
    }

    // Toggle status
    if (body.action === 'toggle' && body.name) {
        MockDB.services = MockDB.services.map(s => s.name === body.name ? { ...s, disable: body.disable } : s);
        return NextResponse.json({ message: "Updated" });
    }

    // Register/Update
    if (body.name) {
        const existing = MockDB.services.find(s => s.name === body.name);
        if (existing) {
             MockDB.services = MockDB.services.map(s => s.name === body.name ? { ...body, id: s.id } : s);
        } else {
            MockDB.services.push({ ...body, id: `srv-${Date.now()}` });
        }
        return NextResponse.json({ message: "Saved" });
    }

    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
}
