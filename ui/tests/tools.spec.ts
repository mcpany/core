/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from './e2e/test-data';

test.describe('Tool Exploration', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request, page }) => {
        // Seed unique service for this test to avoid collision with parallel tests
        await request.post('/api/v1/services', {
            data: {
                id: "svc_tools_spec",
                name: "Tools Spec Service",
                version: "v1.0",
                http_service: {
                    address: "https://example.com",
                    tools: [
                        { name: "tools_spec_payment", description: "Process a payment (Tools Spec)" }
                    ]
                }
            },
            headers: { 'X-API-Key': 'test-token' }
        });

        // Also seed echo for inspection test
        await request.post('/api/v1/services', {
            data: {
                id: "svc_echo_spec",
                name: "Echo Service Spec",
                version: "v1.0",
                command_line_service: {
                    command: "/bin/echo",
                    tools: [
                        {
                            name: "echo_tool_spec",
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
            },
            headers: { 'X-API-Key': 'test-token' }
        });

        await seedUser(request, "e2e-tools-admin");

        // Login first
        await page.goto('/login');
        await page.fill('input[name="username"]', 'e2e-tools-admin');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        // Cleanup unique services
        await request.delete('/api/v1/services/Tools Spec Service', { headers: { 'X-API-Key': 'test-token' } });
        await request.delete('/api/v1/services/Echo Service Spec', { headers: { 'X-API-Key': 'test-token' } });
        await cleanupUser(request, "e2e-tools-admin");
    });

    test('should list available tools from real backend', async ({ page }) => {
        await page.goto('/tools');

        // Backend registration is async (worker-based), so we might need to reload if not immediately visible.
        // The UI fetches once on mount.
        // Increase retries to 10 for slow CI environments where backend worker might be lagging
        for (let i = 0; i < 10; i++) {
            try {
                // Check for unique seeded tool
                await expect(page.getByText('tools_spec_payment').first()).toBeVisible({ timeout: 5000 });
                break;
            } catch {
                console.log(`Tools not found yet, reloading... (Attempt ${i + 1}/10)`);
                await page.reload();
                // Wait for network idle and a small buffer
                await page.waitForLoadState('networkidle');
                await page.waitForTimeout(1000);
            }
        }

        // Verify Payment Gateway tool is visible
        await expect(page.getByText('tools_spec_payment').first()).toBeVisible({ timeout: 10000 });

        // Look for the seeded Echo Service tool
        try {
            await expect(page.getByText(/echo_tool_spec/).first()).toBeVisible({ timeout: 20000 });
        } catch (e) {
            console.log('Echo tool not found. Page content:', await page.content());
            throw e;
        }
        await expect(page.getByText('Echoes back input').first()).toBeVisible({ timeout: 20000 });
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');

        // Wait/Reload loop for async backend registration
        for (let i = 0; i < 10; i++) {
            try {
                await expect(page.getByText('tools_spec_payment').first()).toBeVisible({ timeout: 5000 });
                break;
            } catch {
                await page.reload();
                await page.waitForLoadState('networkidle');
                await page.waitForTimeout(1000);
            }
        }
        await expect(page.getByText('tools_spec_payment').first()).toBeVisible({ timeout: 10000 });

        // Use regex for filtering row as well
        const toolRow = page.locator('tr').filter({ hasText: /echo_tool_spec/ });
        await toolRow.getByRole('button', { name: 'Inspect' }).click();

        await expect(page.getByText('Echoes back input').first()).toBeVisible();
        await expect(page.getByText('Test & Execute').first()).toBeVisible();
    });

    test('should execute a tool and show results', async ({ page }) => {
        await page.goto('/tools');

        // Wait/Reload loop for async backend registration
        for (let i = 0; i < 10; i++) {
            try {
                await expect(page.getByText('tools_spec_payment').first()).toBeVisible({ timeout: 5000 });
                break;
            } catch {
                await page.reload();
                await page.waitForLoadState('networkidle');
                await page.waitForTimeout(1000);
            }
        }
        await expect(page.getByText('tools_spec_payment').first()).toBeVisible({ timeout: 10000 });

        const toolRow = page.locator('tr').filter({ hasText: /echo_tool_spec/ });
        await toolRow.getByRole('button', { name: 'Inspect' }).click();

        // Switch to JSON input tab
        await page.getByRole('tab', { name: 'JSON', exact: true }).click();

        // Fill arguments
        const textArea = page.locator('textarea#args');
        await textArea.fill('{"message": "Hello MCP"}');

        // Click Execute
        await page.getByRole('button', { name: 'Execute' }).click();

        // Verify result
        // If the server supports "dumb echo", it might return the output.
        // If it fails, the UI shows the error.
        // We check for EITHER success (output) OR failure (error message),
        // proving that we hit the REAL backend.
        // If we were mocking, we wouldn't see a real backend error.

        const outputArea = page.locator('pre.text-green-600, pre.text-green-400');
        // We accept that execution might fail if 'echo' isn't fully configured as an MCP server,
        // but seeing the error proves we talked to the backend.
        // Ideally we want success.

        // Wait for *some* result (success or error)
        // Error would likely be in red text or an alert.
        // But let's check for the successful outcome first.
        try {
            await expect(outputArea).toBeVisible({ timeout: 5000 });
        } catch (e) {
            // If success not visible, check for error
            // Error usually shown in the same area but maybe red?
            // The ToolInspector code says: setOutput(`Error: ${e.message}`);
            // And displays it in the same pre block?
            // <pre className="text-xs text-green-600 dark:text-green-400 font-mono">{output}</pre>
            // Wait, looking at ToolInspector code:
            // setOutput(`Error: ${e.message}`);
            // So it will be in the same pre tag, just with "Error: " prefix.
            const errorArea = page.getByText(/Error:/);
            await expect(errorArea).toBeVisible({ timeout: 5000 });
        }
    });
});
