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
            openapi_service: {
                address: "http://ui-http-echo-server:5678",
                spec_content: JSON.stringify({
                    openapi: "3.0.0",
                    info: { title: "Payment Gateway", version: "1.0.0" },
                    paths: {
                        "/pay": {
                            post: {
                                operationId: "process_payment",
                                summary: "Process a payment",
                                responses: { "200": { description: "OK" } }
                            }
                        }
                    }
                })
            }
        },
        {
            id: "svc_02",
            name: "User Service",
            version: "v1.0",
            openapi_service: {
                address: "http://ui-http-echo-server:5678",
                spec_content: JSON.stringify({
                    openapi: "3.0.0",
                    info: { title: "User Service", version: "1.0.0" },
                    paths: {
                        "/user": {
                            get: {
                                operationId: "get_user",
                                summary: "Get user details",
                                responses: { "200": { description: "OK" } }
                            }
                        }
                    }
                })
            }
        },
        {
            id: "svc_03",
            name: "Math",
            version: "v1.0",
            openapi_service: {
                address: "http://ui-http-echo-server:5678",
                spec_content: JSON.stringify({
                    openapi: "3.0.0",
                    info: { title: "Math", version: "1.0.0" },
                    paths: {
                        "/calc": {
                            post: {
                                operationId: "calculator",
                                summary: "calc",
                                responses: { "200": { description: "OK" } }
                            }
                        }
                    }
                })
            }
        }
    ];

    try {
        const res = await context.post('/api/v1/debug/seed', {
            data: { services },
            headers: HEADERS
        });
        if (!res.ok()) {
            console.log(`Failed to seed services: ${res.status()} ${await res.text()}`);
        }
    } catch (e) {
        console.log(`Failed to seed services: ${e}`);
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

    // Generate 60 points
    const now = new Date();
    const traffic = [];
    for (let i = 59; i >= 0; i--) {
        const t = new Date(now.getTime() - i * 60000);
        const timeStr = t.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });
        traffic.push({
            time: timeStr,
            requests: 100,
            errors: 2,
            latency: 50
        });
    }

    try {
        const res = await context.post('/api/v1/debug/seed', {
            data: { traffic },
            headers: HEADERS
        });
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
        await context.delete('/api/v1/services/Payment Gateway', { headers: HEADERS });
        await context.delete('/api/v1/services/User Service', { headers: HEADERS });
        await context.delete('/api/v1/services/Math', { headers: HEADERS });
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
        // Use debug seed endpoint for atomic seeding
        const res = await context.post('/api/v1/debug/seed', {
            data: { users: [user] },
            headers: HEADERS
        });
        if (!res.ok()) {
             // Fallback to internal API if seed fails? No, debug seed should work.
             console.log(`Failed to seed user: ${res.status()} ${await res.text()}`);
        }
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
