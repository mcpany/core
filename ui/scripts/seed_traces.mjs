/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

// Seed Traces for Tool Activity Feed

const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:50050';
const API_KEY = process.env.MCPANY_API_KEY || 'test-token';

async function seed() {
    console.log(`Seeding traces to ${BACKEND_URL}...`);
    try {
        const payload = {
            name: 'calculate_sum',
            arguments: { a: 1, b: 2 }
        };
        const res = await fetch(`${BACKEND_URL}/api/v1/execute`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': API_KEY
            },
            body: JSON.stringify(payload)
        });

        if (!res.ok) {
            const text = await res.text();
            throw new Error(`Execution failed: ${res.status} ${res.statusText} - ${text}`);
        }

        const data = await res.json();
        console.log('Seeded trace successfully. Result:', data);
    } catch (e) {
        console.error('Failed to seed trace:', e);
        process.exit(1);
    }
}

seed();
