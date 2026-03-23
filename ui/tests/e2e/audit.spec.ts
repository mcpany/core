/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

const BASE_URL = process.env.BACKEND_URL || 'http://localhost:50050';
const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
const HEADERS = { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };

const seedAuditLogs = async (requestContext: any) => {
    // Generate a test audit log with a simple object argument and result
    const log = {
        timestamp: new Date().toISOString(),
        toolName: "test_tool_properties_view",
        userId: "test-user",
        profileId: "default",
        arguments: JSON.stringify({ target: "google.com", port: 443, use_tls: true }),
        result: JSON.stringify({ status: "success", resolved_ip: "142.250.190.46", latency_ms: 25 }),
        error: "",
        duration: "100ms",
        durationMs: 100,
        traceId: "test-trace-id",
        spanId: "test-span-id",
        parentId: "test-parent-id"
    };

    try {
        await requestContext.post('/api/v1/mcp/call', {
            headers: HEADERS,
            data: {
                tool_name: "echo",
                arguments: { target: "google.com", port: 443, use_tls: true }
            }
        });
    } catch (e) {
        console.log(`Failed to seed audit log: ${e}`);
    }
};

test.describe('Audit Logs Properties View', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request, page }) => {
        await seedGlobalState(request);
        // We will seed an audit log by calling a tool
        await request.post('/mcp', {
             headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${API_KEY}` },
             data: {
                 jsonrpc: "2.0",
                 id: 1,
                 method: "tools/call",
                 params: {
                     name: "process_payment",
                     arguments: { target: "google.com", port: 443, use_tls: true }
                 }
             }
        });

        await page.goto('/login');
        await page.waitForLoadState('networkidle');
        await page.fill('input[name="username"]', 'e2e-admin-core');
        await page.fill('input[name="password"]', 'password');
        await Promise.all([
            page.waitForURL('/', { timeout: 30000 }),
            page.click('button[type="submit"]', { force: true })
        ]);
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test('should display simple JSON objects as a Properties Table in Audit Log Viewer', async ({ page }) => {
        // Go to audit page
        await page.goto('/audit');

        // Wait for the table to load
        await page.waitForSelector('table');

        // Click the first "View" button
        const viewButton = page.locator('button:has-text("View")').first();
        await expect(viewButton).toBeVisible();
        await viewButton.click();

        // Dialog should open
        const dialog = page.locator('[role="dialog"]');
        await expect(dialog).toBeVisible();

        // Find the "Arguments" section
        const argsSection = dialog.locator('h4:has-text("Arguments") + div');

        // Instead of raw JSON, we should see a Table with the keys "target", "port", "use_tls"
        await expect(argsSection.locator('td', { hasText: 'target' })).toBeVisible();
        await expect(argsSection.locator('td', { hasText: 'google.com' })).toBeVisible();
        await expect(argsSection.locator('td', { hasText: 'port' })).toBeVisible();
        await expect(argsSection.locator('td', { hasText: '443' })).toBeVisible();

        // Verify it's a PropertiesTable (has 'Props' button and renders the table cells)
        // We verify that the "Raw" view is not active by default and we see the table cells.
        await expect(argsSection.locator('button:has-text("Props")')).toBeVisible();

        // Click the close button
        await page.keyboard.press('Escape');
        await expect(dialog).not.toBeVisible();
    });
});
