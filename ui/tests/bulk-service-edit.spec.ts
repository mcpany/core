/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from './e2e/test-data';

test.describe('Bulk Service Edit', () => {

    const service1 = {
        name: "BulkTest1",
        version: "1.0",
        command_line_service: {
            command: "echo",
            env: { "EXISTING": { plain_text: "val1" } }
        }
    };
    const service2 = {
        name: "BulkTest2",
        version: "1.0",
        command_line_service: {
            command: "ls",
            env: { "EXISTING": { plain_text: "val2" } }
        }
    };

    test.beforeEach(async ({ request, page }) => {
        // Seed services
        await request.post('/api/v1/services', { data: service1, headers: { 'X-API-Key': 'test-token' } });
        await request.post('/api/v1/services', { data: service2, headers: { 'X-API-Key': 'test-token' } });
        await seedUser(request, "e2e-admin");

        // Login
        await page.goto('/login');
        await page.fill('input[name="username"]', 'e2e-admin');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
    });

    test.afterEach(async ({ request }) => {
        await request.delete('/api/v1/services/BulkTest1', { headers: { 'X-API-Key': 'test-token' } });
        await request.delete('/api/v1/services/BulkTest2', { headers: { 'X-API-Key': 'test-token' } });
        await cleanupUser(request, "e2e-admin");
    });

    test('can bulk add environment variables', async ({ page, request }) => {
        await page.goto('/upstream-services');

        // Select both services
        const row1 = page.locator('tr', { hasText: 'BulkTest1' });
        const row2 = page.locator('tr', { hasText: 'BulkTest2' });
        await expect(row1).toBeVisible();
        await expect(row2).toBeVisible();

        // Checkboxes are in first cell
        await row1.locator('input[type="checkbox"]').check();
        await row2.locator('input[type="checkbox"]').check();

        // Click Bulk Edit
        await page.click('button:has-text("Bulk Edit")');

        // Go to Environment tab
        await page.click('button[role="tab"]:has-text("Environment")');

        // Add Env Var
        await page.click('button:has-text("Add Variable")');
        await page.fill('input[placeholder="KEY"]', 'NEW_VAR');
        await page.fill('input[placeholder="VALUE"]', 'new_value');

        // Apply
        await page.click('button:has-text("Apply Changes")');

        // Verify toast
        await expect(page.locator('text=Services Updated')).toBeVisible();

        // Verify backend state
        const res1 = await request.get('/api/v1/services/BulkTest1', { headers: { 'X-API-Key': 'test-token' } });
        const json1 = await res1.json();
        const s1 = json1.service || json1; // handle wrapped or flat

        expect(s1.command_line_service.env['NEW_VAR'].plain_text).toBe('new_value');
        expect(s1.command_line_service.env['EXISTING'].plain_text).toBe('val1');

        const res2 = await request.get('/api/v1/services/BulkTest2', { headers: { 'X-API-Key': 'test-token' } });
        const json2 = await res2.json();
        const s2 = json2.service || json2;

        expect(s2.command_line_service.env['NEW_VAR'].plain_text).toBe('new_value');
        expect(s2.command_line_service.env['EXISTING'].plain_text).toBe('val2');
    });
});
