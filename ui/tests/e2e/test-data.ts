/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { request, APIRequestContext } from '@playwright/test';

// Instead of importing the compiled proto messages which are failing to resolve in playwright environment,
// we construct the JSON objects directly as any. Playwright just sends these to the backend /seed endpoint.
const BASE_URL = process.env.BACKEND_URL || 'http://localhost:50050';
const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
const ECHO_SERVER_BASE_URL = process.env.UI_HTTP_ECHO_BASE_URL || 'http://ui-http-echo-server:5678';
const HEADERS = { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };

export const seedGlobalState = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });

    const services = [
        {
            id: "svc_01",
            name: "Payment Gateway",
            version: "v1.2.0",
            httpService: {
                address: "https://stripe.com",
                tools: [
                    { name: "process_payment", description: "Process a payment", callId: "process_payment_call" }
                ],
                calls: {
                    process_payment_call: {
                        method: "HTTP_METHOD_POST",
                        endpointPath: "/v1/charges"
                    }
                }
            }
        },
        {
            id: "svc_02",
            name: "User Service",
            version: "v1.0",
            httpService: {
                address: "http://localhost:50051", // Dummy address
                tools: [
                    { name: "get_user", description: "Get user details", callId: "get_user_call" }
                ],
                calls: {
                    get_user_call: {
                        method: "HTTP_METHOD_GET",
                        endpointPath: "/users/{id}"
                    }
                }
            }
        },
        {
            id: "svc_03",
            name: "Math",
            version: "v1.0",
            httpService: {
                address: ECHO_SERVER_BASE_URL,
                tools: [
                    { name: "calculator", description: "calc", callId: "calc_call" }
                ],
                prompts: [
                    {
                        name: "Calculate Sum",
                        description: "Adds two numbers together",
                        inputSchema: {
                            type: "object",
                            properties: {
                                a: { type: "number", description: "First number" },
                                b: { type: "number", description: "Second number" }
                            },
                            required: ["a", "b"]
                        }
                    }
                ],
                calls: {
                    calc_call: {
                        method: "HTTP_METHOD_POST",
                        endpointPath: "/calc"
                    }
                }
            }
        },
        {
            id: "svc_echo",
            name: "Echo Service",
            version: "v1.0",
            commandLineService: {
                command: "echo",
                tools: [
                    {
                        name: "echo_tool",
                        description: "Echoes back input",
                        inputSchema: { type: "object" },
                        callId: "echo_call"
                    }
                ],
                calls: {
                    echo_call: {
                        args: ["echoed_output"]
                    }
                }
            }
        },
        {
            id: "svc_resource",
            name: "Resource Service",
            version: "v1.0",
            commandLineService: {
                command: "cat",
                resources: [
                    {
                        uri: "file:///test.json",
                        name: "test.json",
                        mimeType: "application/json"
                    }
                ],
                calls: {
                    resource_call: {
                        args: ["{\"key\": \"value\", \"long\": \"content to test modal view\"}"]
                    }
                }
            }
        }
    ];

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
                    apiKey: { value: { plainText: "" } }
                },
                openapiService: {
                    specUrl: "https://raw.githubusercontent.com/linear/linear/master/api/openapi.yaml"
                }
            }
        }
    ];

    const users = [
        {
            id: "e2e-admin-core",
            authentication: {
                basicAuth: {
                    username: "e2e-admin-core",
                    // hash for "password" (bcrypt cost 12)
                    passwordHash: "$2a$12$KPRtQETm7XKJP/L6FjYYxuCFpTK/oRs7v9U6hWx9XFnWy6UuDqK/a"
                }
            },
            roles: ["admin"],
            profileIds: ["dev", "prod"]
        }
    ];

    const credentials = [
        {
            id: 'cred-1',
            name: 'Test Credential',
            authentication: {
                apiKey: {
                    paramName: 'Authorization',
                    in: 0,
                    value: { plainText: 'secret' }
                }
            }
        }
    ];

    const seedRequest = {
        upstream_services: services,
        service_templates: templates,
        users: users,
        credentials: credentials,
        secrets: [],
        profiles: []
    };

    try {
        const res = await context.post('/api/v1/debug/seed', { data: seedRequest, headers: HEADERS });
        if (!res.ok()) {
            const text = await res.text();
            throw new Error(`Failed to seed global state: ${res.status()} ${text}`);
        }
        console.log("Global state seeded successfully.");
    } catch (e) {
        console.log(`Failed to seed global state: ${e}`);
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

// Backward compatibility wrappers to ensure other tests don't break
export const seedServices = async (requestContext?: APIRequestContext) => {
    // Calling seedGlobalState ensures services are present.
    await seedGlobalState(requestContext);
};

export const seedUser = async (requestContext: APIRequestContext | undefined, username: string) => {
    // We create a specific user if requested, in addition to the core user.
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    const user = {
        id: username,
        authentication: {
            basicAuth: {
                username: username,
                passwordHash: "$2a$12$KPRtQETm7XKJP/L6FjYYxuCFpTK/oRs7v9U6hWx9XFnWy6UuDqK/a" // password
            }
        },
        roles: ["admin"], // Default to admin for e2e tests
        profileIds: ["dev"]
    };

    try {
        const res = await context.post('/api/v1/users', { data: user, headers: HEADERS });
        if (!res.ok() && res.status() !== 409) { // Ignore conflict if user exists
            // If user creation fails, we might create it via seed?
            // But seed clears everything.
            // If this is called AFTER seedGlobalState, it adds a user.
            console.log(`Failed to create user ${username}: ${res.status()}`);
        }
    } catch (e) {
        console.log(`Error seeding user ${username}: ${e}`);
    }
};

export const cleanupServices = async (requestContext?: APIRequestContext) => {
    // No-op
};

export const cleanupUser = async (requestContext: APIRequestContext | undefined, username: string) => {
    // No-op
};

export const seedProfiles = async (requestContext?: APIRequestContext) => {
    // Included in seedGlobalState (empty profiles list currently, but we can add if needed)
};

export const cleanupProfiles = async (requestContext?: APIRequestContext) => {
    // No-op
};

export const seedPrompts = async (requestContext?: APIRequestContext) => {
    // No-op
};

export const cleanupPrompts = async (requestContext?: APIRequestContext) => {
    // No-op
};

export const seedWebhooks = async (requestContext?: APIRequestContext) => {
    // No-op
};

export const seedCollection = async (name?: string, requestContext?: APIRequestContext) => {
    if (!name) return;
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    try {
        const res = await context.post('/api/v1/collections', {
            data: {
                name,
                description: `Test collection: ${name}`,
                version: '1.0.0',
                services: [
                    {
                        name: 'weather-service',
                        commandLineService: {
                            command: 'echo weather'
                        }
                    }
                ]
            },
            headers: HEADERS,
        });
        if (!res.ok()) {
            const text = await res.text();
            console.log(`seedCollection: POST /api/v1/collections => ${res.status()} ${text}`);
        }
    } catch (e) {
        console.log(`seedCollection failed: ${e}`);
    }
};

export const cleanupCollection = async (name?: string, requestContext?: APIRequestContext) => {
    if (!name) return;
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    try {
        await context.delete(`/api/v1/collections/${name}`, { headers: HEADERS });
    } catch (e) {
        // Ignore cleanup errors (collection may not exist)
    }
};
