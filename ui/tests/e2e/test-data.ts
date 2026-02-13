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
                address: "https://stripe.com",
                tools: [
                    { name: "process_payment", description: "Process a payment", call_id: "pay" }
                ],
                calls: {
                    pay: {
                        method: "HTTP_METHOD_POST",
                        endpoint_path: "/v1/charges"
                    }
                }
            }
        },
        {
            id: "svc_02",
            name: "User Service",
            version: "v1.0",
            http_service: {
                address: "http://localhost:50051",
                tools: [
                     { name: "get_user", description: "Get user details", call_id: "get" }
                ],
                calls: {
                    get: {
                        method: "HTTP_METHOD_GET",
                        endpoint_path: "/users/{id}"
                    }
                }
            }
        },
        {
            id: "svc_03",
            name: "Math",
            version: "v1.0",
            http_service: {
                // Point to the echo server running in the docker network
                address: "http://ui-http-echo-server:5678",
                tools: [
                    { name: "calculator", description: "calc", call_id: "calc" }
                ],
                calls: {
                    calc: {
                        method: "HTTP_METHOD_POST",
                        endpoint_path: "/echo"
                    }
                }
            }
        },
        {
            id: "svc_echo",
            name: "Echo Service",
            version: "v1.0",
            http_service: {
                // Reuse the same echo server for the Echo Service
                address: "http://ui-http-echo-server:5678",
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
                        method: "HTTP_METHOD_POST",
                        endpoint_path: "/echo",
                        // Pass arguments as JSON body
                        input_transformer: {
                            template: "{{ toJson . }}"
                        }
                    }
                }
            }
        }
    ];

    for (const svc of services) {
        try {
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
        await context.delete('/api/v1/services/Payment Gateway', { headers: HEADERS });
        await context.delete('/api/v1/services/User Service', { headers: HEADERS });
        await context.delete('/api/v1/services/Math', { headers: HEADERS });
        await context.delete('/api/v1/services/Echo Service', { headers: HEADERS });
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

export const seedTemplates = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    const templates = [
        {
            id: "google-calendar",
            name: "Google Calendar",
            description: "Manage events and calendars.",
            icon: "calendar",
            tags: ["google", "productivity"],
            serviceConfig: {
                name: "google_calendar",
                upstreamAuth: {
                    oauth2: {
                        tokenUrl: "https://oauth2.googleapis.com/token",
                        clientId: { plainText: "" },
                        clientSecret: { plainText: "" },
                        scopes: "https://www.googleapis.com/auth/calendar"
                    }
                },
                openapiService: {
                    specUrl: "https://api.apis.guru/v2/specs/googleapis.com/calendar/v3/openapi.yaml"
                }
            }
        },
        {
            id: "github",
            name: "GitHub",
            description: "Interact with repositories, issues, and PRs.",
            icon: "github",
            tags: ["dev", "git"],
            serviceConfig: {
                name: "github",
                upstreamAuth: {
                    bearerToken: { token: { plainText: "" } }
                },
                openapiService: {
                    address: "https://api.github.com",
                    specUrl: "https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.yaml"
                }
            }
        },
        {
            id: "linear",
            name: "Linear",
            description: "Issue tracking and project management.",
            icon: "linear",
            tags: ["dev", "pm"],
            serviceConfig: {
                name: "linear",
                upstreamAuth: {
                    apiKey: { plainText: "" }
                },
                openapiService: {
                    specUrl: "https://raw.githubusercontent.com/linear/linear/master/api/openapi.yaml"
                }
            }
        }
    ];

    for (const tmpl of templates) {
        try {
            const res = await context.post('/api/v1/templates', { data: tmpl, headers: HEADERS });
            if (!res.ok()) {
                console.log(`Failed to seed template ${tmpl.name}: ${res.status()} ${await res.text()}`);
            }
        } catch (e) {
            console.log(`Failed to seed template ${tmpl.name}: ${e}`);
        }
    }
};

export const cleanupTemplates = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    const ids = ["google-calendar", "github", "linear"];
    for (const id of ids) {
        try {
            await context.delete(`/api/v1/templates/${id}`, { headers: HEADERS });
        } catch (e) {
            console.log(`Failed to cleanup template ${id}: ${e}`);
        }
    }
};

export const seedWebhooks = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    try {
        const res = await context.post('/api/v1/alerts/webhook', {
            data: { url: "https://example.com/webhook" },
            headers: HEADERS
        });
        if (!res.ok()) {
            console.log(`Failed to seed webhook: ${res.status()} ${await res.text()}`);
        }
    } catch (e) {
        console.log(`Failed to seed webhook: ${e}`);
    }
};
