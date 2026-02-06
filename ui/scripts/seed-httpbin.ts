/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

// ui/scripts/seed-httpbin.ts

// Polyfill fetch if needed (Node 18+ has it natively)
if (!globalThis.fetch) {
    console.error("This script requires Node.js 18+ with native fetch.");
    process.exit(1);
}

const API_BASE = process.env.API_BASE || "http://localhost:50050";

async function main() {
    console.log(`Seeding httpbin service to ${API_BASE}...`);

    const serviceConfig = {
        name: "httpbin-seed",
        id: "httpbin-seed",
        http_service: {
            address: "https://httpbin.org"
        },
        disable: false
    };

    try {
        const response = await fetch(`${API_BASE}/api/v1/services`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(serviceConfig)
        });

        if (!response.ok) {
            // Check if it already exists (409 Conflict typically, but backend might return something else)
            // If it exists, we might want to update or ignore.
            const text = await response.text();
            if (response.status === 409 || text.includes("already exists")) {
                 console.log("Service already exists. Skipping.");
                 return;
            }
            throw new Error(`Failed to seed service: ${response.status} ${text}`);
        }

        const data = await response.json();
        console.log("Successfully seeded service:", data);
    } catch (error) {
        console.error("Error seeding service:", error);
        process.exit(1);
    }
}

main();
