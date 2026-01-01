/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { SecretsStore } from '@/lib/server/secrets-store';

export async function GET() {
    try {
        const secrets = SecretsStore.getAllDecrypted();
        return NextResponse.json(secrets);
    } catch (error) {
        return NextResponse.json({ error: 'Failed to fetch secrets' }, { status: 500 });
    }
}

export async function POST(request: Request) {
    try {
        const body = await request.json();

        // Basic validation
        if (!body.name || !body.key || !body.value) {
             return NextResponse.json({ error: 'Missing required fields' }, { status: 400 });
        }

        const newSecret = SecretsStore.add({
            id: body.id || Math.random().toString(36).substring(7),
            name: body.name,
            key: body.key,
            value: body.value,
            provider: body.provider || 'custom',
            createdAt: new Date().toISOString(),
            lastUsed: 'Never'
        });

        // Return the version with 'value' so the UI can update immediately if needed,
        // though usually we mask it.
        return NextResponse.json({
            ...newSecret,
            value: body.value, // Return decrypted for confirmation
            encryptedValue: undefined
        });

    } catch (error) {
        return NextResponse.json({ error: 'Failed to create secret' }, { status: 500 });
    }
}
