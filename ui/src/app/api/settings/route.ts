/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { MockDB } from '@/lib/server/mock-db';

export async function GET() {
  return NextResponse.json(MockDB.settings);
}

export async function POST(request: Request) {
    try {
        const body = await request.json();
        MockDB.settings = { ...MockDB.settings, ...body };
        return NextResponse.json(MockDB.settings);
    } catch (e) {
        return NextResponse.json({ error: "Invalid request" }, { status: 400 });
    }
}
