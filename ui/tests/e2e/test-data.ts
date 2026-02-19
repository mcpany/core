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
                    { name: "process_payment", description: "Process a payment", call_id: "process_payment_call" }
                ],
                calls: {
                    process_payment_call: {
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
                address: "http://localhost:50051", // Dummy address
                tools: [
                    { name: "get_user", description: "Get user details", call_id: "get_user_call" }
                ],
                calls: {
                    get_user_call: {
                        method: "HTTP_METHOD_GET",
                        endpoint_path: "/users/{id}"
                    }
                }
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
                    { name: "calculator", description: "calc", call_id: "calc_call" }
                ],
                calls: {
                    calc_call: {
                        method: "HTTP_METHOD_POST",
                        endpoint_path: "/calc"
                    }
                }
            }
        },
        {
            id: "svc_echo",
            name: "Echo Service",
            version: "v1.0",
            command_line_service: {
                command: "echo",
                tools: [
                    {
                        name: "echo_tool",
                        description: "Echoes back input",
                        input_schema: { type: "object" },
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
            const res = await context.post('/api/v1/services', { data: svc, headers: HEADERS });
            if (!res.ok()) {
                const text = await res.text();
                throw new Error(`${res.status()} ${text}`);
            }
        } catch (e) {
            console.log(`Failed to seed service ${svc.name}: ${e}`);
            throw e;
        }
    }
};

export const seedDebugData = async (data: any, requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    const res = await context.post('/api/v1/debug/seed', {
        data: data,
        headers: HEADERS
    });
    if (!res.ok()) {
        const text = await res.text();
        throw new Error(`Failed to seed debug data: ${res.status()} ${text}`);
    }
};

export const seedCollection = async (name: string, requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    const collection = {
        name: name,
        services: [
            {
                name: "weather-service",
                // Use command_line_service matching config.minimal.yaml to avoid Docker issues in E2E
                command_line_service: {
                    command: "echo",
                    tools: [
                        {
                            name: "get_weather",
                            description: "Get current weather",
                            call_id: "get_weather"
                        }
                    ],
                    calls: {
                        get_weather: {
                            args: ['{"weather": "sunny"}']
                        }
                    }
                }
            }
        ]
    };
    try {
        const res = await context.post('/api/v1/collections', { data: collection, headers: HEADERS });
        if (!res.ok()) {
            const text = await res.text();
            throw new Error(`Failed to seed collection ${name}: ${res.status()} ${text}`);
        }
    } catch (e) {
        console.log(`Failed to seed collection ${name}: ${e}`);
        throw e;
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

export const seedTemplates = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    const templates = [
        {
            id: "google-calendar",
            name: "Google Calendar",
            description: "Manage events and calendars.",
            icon: "calendar",
            tags: ["google", "productivity"],
            service_config: {
                name: "google_calendar",
                upstream_auth: {
                    oauth2: {
                        token_url: "https://oauth2.googleapis.com/token",
                        client_id: { plain_text: "" },
                        client_secret: { plain_text: "" },
                        scopes: "https://www.googleapis.com/auth/calendar"
                    }
                },
                openapi_service: {
                    spec_url: "https://api.apis.guru/v2/specs/googleapis.com/calendar/v3/openapi.yaml"
                }
            }
        },
        {
            id: "github",
            name: "GitHub",
            description: "Interact with repositories, issues, and PRs.",
            icon: "github",
            tags: ["dev", "git"],
            service_config: {
                name: "github",
                upstream_auth: {
                    bearer_token: { token: { plain_text: "" } }
                },
                openapi_service: {
                    address: "https://api.github.com",
                    spec_url: "https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.yaml"
                }
            }
        },
        {
            id: "linear",
            name: "Linear",
            description: "Issue tracking and project management.",
            icon: "linear",
            tags: ["dev", "pm"],
            service_config: {
                name: "linear",
                upstream_auth: {
                    api_key: { plain_text: "" }
                },
                openapi_service: {
                    spec_url: "https://raw.githubusercontent.com/linear/linear/master/api/openapi.yaml"
                }
            }
        }
    ];

    for (const tmpl of templates) {
        try {
            await context.post('/api/v1/templates', { data: tmpl, headers: HEADERS });
        } catch (e) {
            console.log(`Failed to seed template ${tmpl.name}: ${e}`);
        }
    }
};

export const seedWebhooks = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });

    // Seed global webhook
    try {
        await context.post('/api/v1/alerts/webhook', {
            data: { url: "https://example.com/webhook" },
            headers: HEADERS
        });
    } catch (e) {
        console.log(`Failed to seed global webhook: ${e}`);
    }

    // Seed alert rules (which might be displayed as webhooks or alerts)
    const alerts = [
        {
            id: "alert_1",
            name: "Critical Service Down",
            condition: "service.status == 'down' && service.priority <= 1",
            action: { webhook: { url: "https://pagerduty.com/webhook" } }
        }
    ];

    for (const alert of alerts) {
        try {
            // Note: If /api/v1/alerts supports POST for creating alerts/rules
            // api_alerts.go has handleAlerts (POST -> CreateAlert) and handleAlertRules (POST -> CreateRule).
            // The object above looks more like a Rule.
            await context.post('/api/v1/alerts/rules', { data: alert, headers: HEADERS });
        } catch (e) {
            console.log(`Failed to seed alert ${alert.name}: ${e}`);
        }
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

export const cleanupWebhooks = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    // Cleanup global webhook (set to empty)
    try {
        await context.post('/api/v1/alerts/webhook', { data: { url: "" }, headers: HEADERS });
    } catch (e) {
        console.log(`Failed to cleanup global webhook: ${e}`);
    }
    // Cleanup alerts/rules
    try {
        await context.delete(`/api/v1/alerts/rules/alert_1`, { headers: HEADERS });
    } catch (e) {
        console.log(`Failed to cleanup alert rule: ${e}`);
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
        roles: ["admin"],
        profile_ids: ["dev", "prod"]
    };
    // We don't delete first anymore to avoid race conditions in parallel tests
    // await cleanupUser(context, username);
    try {
        // We use the internal API to seed the user. This request uses HEADERS (API Key) which bypasses auth on backend.
        const res = await context.post('/api/v1/users', { data: user, headers: HEADERS });
        if (!res.ok() && res.status() !== 409) {
            const text = await res.text();
            throw new Error(`${res.status()} ${text}`);
        }
    } catch (e) {
        console.log(`Failed to seed user: ${e}`);
        throw e;
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

export const seedProfiles = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    const profiles = [
        {
            name: "dev",
            required_roles: ["admin"],
            service_config: {
                "svc_01": { enabled: true }
            }
        },
        {
            name: "prod",
            required_roles: ["admin"],
            service_config: {
                "svc_01": { enabled: false }
            }
        }
    ];

    for (const p of profiles) {
        try {
            await context.post('/api/v1/profiles', { data: p, headers: HEADERS });
        } catch (e) {
            console.log(`Failed to seed profile ${p.name}: ${e}`);
        }
    }
};

export const cleanupProfiles = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    const names = ["dev", "prod"];
    for (const name of names) {
        try {
            await context.delete(`/api/v1/profiles/${name}`, { headers: HEADERS });
        } catch (e) {
            console.log(`Failed to cleanup profile ${name}: ${e}`);
        }
    }
};
