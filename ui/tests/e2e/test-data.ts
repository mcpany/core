/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { request, APIRequestContext, expect } from '@playwright/test';

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
                address: "https://stripe.com",
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
                address: "http://localhost:50051", // Dummy address, visibility checks don't need health
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
                address: "http://localhost:8080", // Dummy
                tools: [
                    { name: "calculator", description: "calc" }
                ]
            }
        }
    ];

    for (const svc of services) {
        const response = await context.post('/api/v1/services', { data: svc, headers: HEADERS });
        if (!response.ok()) {
            throw new Error(`Failed to seed service ${svc.name}: ${response.status()} ${await response.text()}`);
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

    const res = await context.post('/api/v1/collections', { data: collection, headers: HEADERS });
    if (!res.ok()) {
        console.warn(`Failed to seed collection ${name}: ${res.status()} ${await res.text()}`);
        // Don't throw for collections as they might already exist or conflict less critically
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
        console.warn(`Failed to seed traffic: ${e}`);
    }
};

export const cleanupServices = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    try {
        await context.delete('/api/v1/services/Payment Gateway', { headers: HEADERS });
        await context.delete('/api/v1/services/User Service', { headers: HEADERS });
        await context.delete('/api/v1/services/Math', { headers: HEADERS });
    } catch (e) {
        console.warn(`Failed to cleanup services: ${e}`);
    }
};

export const cleanupCollection = async (name: string, requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    try {
        await context.delete(`/api/v1/collections/${name}`, { headers: HEADERS });
    } catch (e) {
        console.warn(`Failed to cleanup collection ${name}: ${e}`);
    }
};

export const seedUser = async (requestContext?: APIRequestContext, username: string = "admin") => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    const user = {
        id: username,
        authentication: {
            basic_auth: {
                // hash for "password" (bcrypt cost 12)
                password_hash: "$2a$12$KPRtQETm7XKJP/L6FjYYxuCFpTK/oRs7v9U6hWx9XFnWy6UuDqK/a"
            }
        },
        roles: ["admin"]
    };

    // We use the internal API to seed the user. This request uses HEADERS (API Key) which bypasses auth on backend.
    const response = await context.post('/api/v1/users', { data: { user }, headers: HEADERS });

    if (!response.ok()) {
        const text = await response.text();
        // If user already exists (409), that's fine.
        if (response.status() !== 409) {
             throw new Error(`Failed to seed user ${username}: ${response.status()} ${text}`);
        }
    }
};

export const cleanupUser = async (requestContext?: APIRequestContext, username: string = "admin") => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    try {
        await context.delete(`/api/v1/users/${username}`, { headers: HEADERS });
    } catch (e) {
        console.warn(`Failed to cleanup user: ${e}`);
    }
};
