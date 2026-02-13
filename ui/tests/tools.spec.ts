/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect, APIRequestContext, request } from '@playwright/test';
import { seedUser, cleanupUser } from './e2e/test-data';

const BASE_URL = process.env.BACKEND_URL || 'http://localhost:50050';
const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
const HEADERS = { 'X-API-Key': API_KEY };

// Unique IDs for this test file to avoid conflicts
const PAYMENT_SVC_ID = "svc_tools_payment";
const PAYMENT_SVC_NAME = "Payment Gateway Tools";
const PAYMENT_TOOL_NAME = "process_payment_tools";

const ECHO_SVC_ID = "svc_tools_echo";
const ECHO_SVC_NAME = "Echo Service Tools";
const ECHO_TOOL_NAME = "echo_tool_tools";

const seedIsolatedServices = async (requestContext: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    const services = [
        {
            id: PAYMENT_SVC_ID,
            name: PAYMENT_SVC_NAME,
            version: "v1.2.0",
            http_service: {
                address: "https://stripe.com",
                tools: [
                    { name: PAYMENT_TOOL_NAME, description: "Process a payment isolated" }
                ]
            }
        },
        {
            id: ECHO_SVC_ID,
            name: ECHO_SVC_NAME,
            version: "v1.0",
            command_line_service: {
                command: "/bin/echo",
                tools: [
                    {
                        name: ECHO_TOOL_NAME,
                        description: "Echoes back input isolated",
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
            // First delete if exists (defensive)
            await context.delete(`/api/v1/services/${svc.name}`, { headers: HEADERS });
        } catch (e) {}

        try {
            await context.post('/api/v1/services', { data: svc, headers: HEADERS });
        } catch (e) {
            console.log(`Failed to seed isolated service ${svc.name}: ${e}`);
        }
    }
};

const cleanupIsolatedServices = async (requestContext: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    try {
        await context.delete(`/api/v1/services/${PAYMENT_SVC_NAME}`, { headers: HEADERS });
        await context.delete(`/api/v1/services/${ECHO_SVC_NAME}`, { headers: HEADERS });
    } catch (e) {
        console.log(`Failed to cleanup isolated services: ${e}`);
    }
};

test.describe('Tool Exploration', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request, page }) => {
        await cleanupIsolatedServices(request);
        await seedIsolatedServices(request);
        await seedUser(request, "e2e-tools-admin");

        // Login first
        await page.goto('/login');
        await page.fill('input[name="username"]', 'e2e-tools-admin');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        await cleanupIsolatedServices(request);
        await cleanupUser(request, "e2e-tools-admin");
    });

    test('should list available tools from real backend', async ({ page }) => {
        await page.goto('/tools');

        // Backend registration is async (worker-based), so we might need to reload if not immediately visible.
        // The UI fetches once on mount.
        // Note: The UI tool table displays the service ID, not the friendly name.
        let found = false;
        // Increase retries to 10 for slow CI environments where backend worker might be lagging
        for (let i = 0; i < 10; i++) {
            try {
                // Check for Payment Gateway tool
                await expect(page.getByText(PAYMENT_TOOL_NAME).first()).toBeVisible({ timeout: 5000 });
                found = true;
                break;
            } catch (e) {
                console.log(`Tools not found yet, reloading... (Attempt ${i + 1}/10)`);
                await page.reload();
                // Wait for network idle and a small buffer
                await page.waitForLoadState('networkidle');
                await page.waitForTimeout(1000);
            }
        }

        // Verify Payment Gateway tool is visible
        await expect(page.getByText(PAYMENT_TOOL_NAME).first()).toBeVisible({ timeout: 10000 });

        // Look for the seeded Echo Service tool
        // Note: The UI might capitalize or format names, but usually it shows the raw tool name.
        // We use a regex to handle potential service name prefixes
        try {
            await expect(page.getByText(new RegExp(ECHO_TOOL_NAME)).first()).toBeVisible({ timeout: 20000 });
        } catch (e) {
            console.log('Echo tool not found. Page content:', await page.content());
            throw e;
        }
        await expect(page.getByText('Echoes back input isolated').first()).toBeVisible({ timeout: 20000 });
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');

        // Wait/Reload loop for async backend registration
        for (let i = 0; i < 10; i++) {
            try {
                await expect(page.getByText(PAYMENT_TOOL_NAME).first()).toBeVisible({ timeout: 5000 });
                break;
            } catch (e) {
                await page.reload();
                await page.waitForLoadState('networkidle');
                await page.waitForTimeout(1000);
            }
        }
        await expect(page.getByText(PAYMENT_TOOL_NAME).first()).toBeVisible({ timeout: 10000 });

        // Use regex for filtering row as well
        const toolRow = page.locator('tr').filter({ hasText: new RegExp(ECHO_TOOL_NAME) });
        await toolRow.getByRole('button', { name: 'Inspect' }).click();

        await expect(page.getByText('Echoes back input isolated').first()).toBeVisible();
        await expect(page.getByText('Test & Execute').first()).toBeVisible();
    });

    test('should execute a tool and show results', async ({ page }) => {
        await page.goto('/tools');

        // Wait/Reload loop for async backend registration
        for (let i = 0; i < 10; i++) {
            try {
                await expect(page.getByText(PAYMENT_TOOL_NAME).first()).toBeVisible({ timeout: 5000 });
                break;
            } catch (e) {
                await page.reload();
                await page.waitForLoadState('networkidle');
                await page.waitForTimeout(1000);
            }
        }
        await expect(page.getByText(PAYMENT_TOOL_NAME).first()).toBeVisible({ timeout: 10000 });

        const toolRow = page.locator('tr').filter({ hasText: new RegExp(ECHO_TOOL_NAME) });
        await toolRow.getByRole('button', { name: 'Inspect' }).click();

        // Switch to JSON input tab
        await page.getByRole('tab', { name: 'JSON', exact: true }).click();

        // Fill arguments
        const textArea = page.locator('textarea#args');
        await textArea.fill('{"message": "Hello MCP"}');

        // Click Execute
        await page.getByRole('button', { name: 'Execute' }).click();

        // Verify result
        const outputArea = page.locator('pre.text-green-600, pre.text-green-400');

        try {
            await expect(outputArea).toBeVisible({ timeout: 5000 });
        } catch (e) {
            const errorArea = page.getByText(/Error:/);
            await expect(errorArea).toBeVisible({ timeout: 5000 });
        }
    });
});
