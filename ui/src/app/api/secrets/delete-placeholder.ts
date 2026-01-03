/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { SecretsStore } from '@/lib/server/secrets-store';

export async function DELETE(request: Request, { params }: { params: { id: string } }) {
    // In Next.js 15, params is a Promise or needs to be awaited if dynamic
    // But since this is a route handler without dynamic path segment in filename (it's /api/secrets/[id]/route.ts if used)
    // Actually wait, I need to check where this file is.
    // The previous plan said `ui/src/app/api/secrets/route.ts` handles GET and POST.
    // I need a separate file for DELETE if I want `/api/secrets/[id]`.
    return NextResponse.json({ error: "Use /api/secrets/[id]" }, { status: 404 });
}
