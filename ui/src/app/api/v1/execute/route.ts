/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { executeTool } from '@/lib/server/tools';

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const { tool_name, arguments: args } = body;

    if (!tool_name) {
      return NextResponse.json({ error: 'Missing tool_name' }, { status: 400 });
    }

    const result = await executeTool(tool_name, args || {});

    // Simulate network delay for "realism"
    await new Promise(resolve => setTimeout(resolve, 500));

    return NextResponse.json(result);
  } catch (error: any) {
    return NextResponse.json({ error: error.message || 'Execution failed' }, { status: 500 });
  }
}
