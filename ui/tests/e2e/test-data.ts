/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { request, APIRequestContext } from '@playwright/test';

const BASE_URL = process.env.BACKEND_URL || 'http://localhost:50050';
const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
const HEADERS = { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };

export const seedGlobalState = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });

    const services = [
        {
            id: "svc_01",
            name: "Payment Gateway",
            version: "v1.2.0",
            http_service: {
                address: "https://stripe.com",
                tools: [
                    { name: "process_payment", description: "Process a payment", call_id: "process_payment_call", input_schema: { type: "object", properties: { amount: { type: "number", description: "Payment amount in cents" } } } }
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
            grpc_service: {
                address: "localhost:50051",
                tools: [
                    { name: "get_user", description: "Get a user", call_id: "GetUser" }
                ],
                calls: {
                    "GetUser": { method_path: "/UserService/GetUser" }
                }
            }
        }
    ];

    try {
        const payload = JSON.stringify({ upstream_services: services });
        const response = await context.post(`/api/v1/debug/seed`, {
            data: payload,
            headers: HEADERS
        });
        if (!response.ok()) {
            const body = await response.text();
            console.log(`Failed to seed data: ${response.status()} ${response.statusText()} ${body}`);
        }
    } catch (error) {
        console.log(`Error seeding data: ${error}`);
    }
};

export const seedServices = async (requestContext?: APIRequestContext) => {
    await seedGlobalState(requestContext);
};

export const seedUser = async (requestContext: APIRequestContext | undefined, username: string) => {
    // No-op for now to keep it simple, or implement simple call if needed
};

export const cleanupUser = async (requestContext: APIRequestContext | undefined, username: string) => {
    // No-op
};

export const seedProfiles = async (requestContext?: APIRequestContext) => {
    // No-op
};

export const cleanupProfiles = async (requestContext?: APIRequestContext) => {
    // No-op
};
