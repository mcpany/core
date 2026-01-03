/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { BuiltInTools } from '@/lib/server/tools';

export async function GET() {
  const tools = Object.values(BuiltInTools).map(tool => ({
    name: tool.name,
    description: tool.description,
    schema: tool.schema,
    enabled: true,
    serviceName: "builtin-nextjs",
    source: "builtin"
  }));

  return NextResponse.json({ tools });
}
