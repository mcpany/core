/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { request, APIRequestContext } from '@playwright/test';

const BASE_URL = process.env.BACKEND_URL || 'http://localhost:50050';
const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
const HEADERS = { 'X-API-Key': API_KEY };

export const seedServices = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    const services = [
        {
            id: "svc_01",
            name: "Payment Gateway",
            version: "v1.2.0",
            http_service: {
                // Use localhost instead of stripe.com to avoid external network dependency/blocking in CI.
                // Connection refused is expected and acceptable, tools should still register.
                address: "http://127.0.0.1:12345",
                tools: [
                    { name: "process_payment", description: "Process a payment" }
                ]
            }
        },
        {
            id: "svc_02",
            name: "User Service",
            version: "v1.0",
            http_service: {
                address: "http://127.0.0.1:50051", // Dummy address, visibility checks don't need health
                tools: [
                     { name: "get_user", description: "Get user details" }
                ]
            }
        },
        // Add a service with calculator for existing test compatibility if desired
        {
            id: "svc_03",
            name: "Math",
            version: "v1.0",
            http_service: {
                address: "http://127.0.0.1:8080", // Dummy
                tools: [
                    { name: "calculator", description: "calc" }
                ]
            }
        },
        {
            id: "svc_echo",
            name: "Echo Service",
            version: "v1.0",
            command_line_service: {
                command: "/bin/echo",
                tools: [
                    {
                        name: "echo_tool",
                        description: "Echoes back input",
                        inputSchema: { type: "object" },
                        call_id: "echo_call"
                    }
                ],
                calls: {
                    echo_call: {
                        args: ["echoed_output"]
                    }
                }
            }
        }
    ];

    for (const svc of services) {
        try {
            await context.post('/api/v1/services', { data: svc, headers: HEADERS });
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
        const res = await context.post('/api/v1/collections', { data: collection, headers: HEADERS });
        if (!res.ok()) {
            console.log(`Failed to seed collection ${name}: ${res.status()} ${await res.text()}`);
        }
    } catch (e) {
        console.log(`Failed to seed collection ${name}: ${e}`);
    }
};

export const seedTraffic = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    const points = [
        { timestamp: new Date().toISOString(), requests: 100, errors: 2 }
    ];
    try {
        await context.post('/api/v1/debug/seed_traffic', { data: points, headers: HEADERS });
    } catch (e) {
        console.log(`Failed to seed traffic: ${e}`);
    }
};

export const cleanupServices = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    try {
        // Fetch all services first to ensure complete cleanup
        const response = await context.get('/api/v1/services', { headers: HEADERS });
        if (response.ok()) {
            const data = await response.json();
            const services = data.services || [];
            for (const svc of services) {
                // Delete by name (as ID might not be in the list view response depending on API)
                // Assuming 'name' is the key or we can use ID if available.
                // The API usually takes Name or ID in the path. Using Name is safer if ID is hidden.
                const key = svc.name || svc.id;
                if (key) {
                    await context.delete(`/api/v1/services/${key}`, { headers: HEADERS });
                }
            }
        }
    } catch (e) {
        console.log(`Failed to cleanup services: ${e}`);
        // Fallback to known names if list fails
        const names = ['Payment Gateway', 'Payment Gateway Updated', 'User Service', 'Math', 'Echo Service'];
        for (const name of names) {
             await context.delete(`/api/v1/services/${name}`, { headers: HEADERS }).catch(() => {});
        }
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

export const seedUser = async (requestContext?: APIRequestContext, username: string = "admin") => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
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
        // We use the internal API to seed the user. This request uses HEADERS (API Key) which bypasses auth on backend.
        await context.post('/api/v1/users', { data: { user }, headers: HEADERS });
    } catch (e) {
        console.log(`Failed to seed user: ${e}`);
    }
};

export const cleanupUser = async (requestContext?: APIRequestContext, username: string = "admin") => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    try {
        await context.delete(`/api/v1/users/${username}`, { headers: HEADERS });
    } catch (e) {
        console.log(`Failed to cleanup user: ${e}`);
    }
};
