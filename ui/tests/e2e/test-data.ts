/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { request, APIRequestContext } from '@playwright/test';

const BASE_URL = process.env.BACKEND_URL || 'http://localhost:50050';
const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
const HEADERS = { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };

export const seedServices = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    const services = [
        {
            id: "svc_01_new",
            name: "Payment Gateway New",
            version: "v1.2.0",
            command_line_service: {
                command: "echo",
                tools: [
                    { name: "process_payment", description: "Process a payment" }
                ]
            }
        },
        {
            id: "svc_02_new",
            name: "User Service New",
            version: "v1.0",
            command_line_service: {
                command: "echo",
                tools: [
                     { name: "get_user", description: "Get user details" }
                ]
            }
        },
        {
            id: "svc_03_new",
            name: "Math New",
            version: "v1.0",
            command_line_service: {
                command: "echo",
                tools: [
                    { name: "calculator", description: "Perform basic math" }
                ]
            }
        }
    ];

    for (const svc of services) {
        try {
            await context.delete(`/api/v1/services/${svc.name}`, { headers: HEADERS }).catch(() => {});
            const res = await context.post('/api/v1/services', { data: svc, headers: HEADERS });
            if (!res.ok()) {
                console.log(`Failed to seed service ${svc.name}: ${res.status()} ${await res.text()}`);
            }
        } catch (e) {
            console.log(`Failed to seed service ${svc.name}: ${e}`);
        }
    }
};

export const seedCollection = async (name: string, requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    const collection = {
        name: name,
        services: [
            {
                name: "weather-service",
                mcp_service: {
                    stdio_connection: {
                        command: "weather",
                        container_image: "mcpany/weather-service:latest",
                        env: {
                            API_KEY: { plain_text: "secret" }
                        }
                    }
                }
            }
        ]
    };
    try {
        await context.delete(`/api/v1/collections/${name}`, { headers: HEADERS }).catch(() => {});
        const res = await context.put(`/api/v1/collections/${name}`, { data: collection, headers: HEADERS });
        if (!res.ok()) {
            console.log(`Failed to seed collection ${name}: ${res.status()} ${await res.text()}`);
        }
    } catch (e) {
        console.log(`Failed to seed collection ${name}: ${e}`);
    }
};

export const seedTraffic = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });

    const now = new Date();
    const trafficPoints = [];

    // Seed last 15 minutes
    for (let i = 14; i >= 0; i--) {
        const t = new Date(now.getTime() - i * 60000);
        const localTimeStr = t.getHours().toString().padStart(2, '0') + ':' + t.getMinutes().toString().padStart(2, '0');

        trafficPoints.push({
            time: localTimeStr,
            requests: 100 + Math.floor(Math.random() * 50),
            errors: i % 5 === 0 ? 5 : 0,
            latency: 50 + Math.floor(Math.random() * 20)
        });
    }

    try {
        const res = await context.post('/api/v1/debug/seed_traffic', { data: trafficPoints, headers: HEADERS });
        if (!res.ok()) {
             console.log(`Failed to seed traffic: ${res.status()} ${await res.text()}`);
        }
    } catch (e) {
        console.log(`Failed to seed traffic: ${e}`);
    }
};

export const cleanupServices = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    try {
        await context.delete('/api/v1/services/Payment Gateway New', { headers: HEADERS });
        await context.delete('/api/v1/services/User Service New', { headers: HEADERS });
        await context.delete('/api/v1/services/Math New', { headers: HEADERS });
    } catch (e) {
        console.log(`Failed to cleanup services: ${e}`);
    }
};

export const cleanupCollection = async (name: string, requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    try {
        await context.delete(`/api/v1/collections/${name}`, { headers: HEADERS });
    } catch (e) {
        console.log(`Failed to cleanup collection ${name}: ${e}`);
    }
};

export const seedUser = async (requestContext?: APIRequestContext, username: string = "e2e-admin") => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });

    // First try to delete the user to ensure a clean state
    try {
        await context.delete(`/api/v1/users/${username}`, { headers: HEADERS }).catch(() => {});
    } catch (e) {
        // Ignore deletion errors
    }

    const user = {
        id: username,
        authentication: {
            basic_auth: {
                username: username,
                // hash for "password" (bcrypt cost 12)
                password_hash: "$2a$12$KPRtQETm7XKJP/L6FjYYxuCFpTK/oRs7v9U6hWx9XFnWy6UuDqK/a"
            }
        },
        roles: ["admin"]
    };
    try {
        const res = await context.post('/api/v1/users', { data: { user }, headers: HEADERS });
        if (!res.ok()) {
            // Check if it failed because it already exists (race condition or persistence)
            const text = await res.text();
            if (!text.includes("already exists")) {
                 console.log(`Failed to seed user: ${res.status()} ${text}`);
            } else {
                 // If it exists, we assume it's good (maybe from previous run)
                 // But strictly we should have deleted it.
                 console.log(`User already exists: ${text}`);
            }
        }
    } catch (e) {
        console.log(`Failed to seed user: ${e}`);
    }
};

export const cleanupUser = async (requestContext?: APIRequestContext, username: string = "e2e-admin") => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    try {
        await context.delete(`/api/v1/users/${username}`, { headers: HEADERS });
    } catch (e) {
        console.log(`Failed to cleanup user: ${e}`);
    }
};
