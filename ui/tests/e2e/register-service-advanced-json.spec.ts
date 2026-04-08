/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Register Service Dialog - Advanced JSON', () => {
    test.beforeEach(async ({ request, page }) => {
        // Seed the database with a test service and users
        await seedGlobalState(request);

        // Login before each test
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

    test('should allow entering and saving config via Monaco Editor in Advanced tab', async ({ page, request }) => {
        // Navigate to upstream services
        await page.goto('/upstream-services');

        // Click the Register Service button
        await page.getByRole('button', { name: 'Add Service' }).click();

        // Start from Scratch (Blank template)
        await page.getByRole('button', { name: 'Start from Scratch' }).click();

        // Switch to the Advanced (JSON) tab
        await page.getByRole('tab', { name: 'Advanced (JSON)' }).click();

        // Focus the Monaco editor and enter new config JSON
        // Monaco editor uses a hidden textarea for input, but we can also evaluate script to set the value.
        // It's cleaner to evaluate script to set Monaco's model value, but Playwright might not have direct access to Monaco instance easily.
        // The most robust way to interact with Monaco in Playwright is via keyboard typing or pasting.
        // Let's locate the monaco editor lines area and click it, then select all and paste.
        const editorLocator = page.locator('.monaco-editor');
        await expect(editorLocator).toBeVisible();

        const testServiceId = 'advanced-json-test-service';
        const configPayload = {
            name: testServiceId,
            httpService: {
                address: "https://api.advanced-json-test.com"
            }
        };
        const configString = JSON.stringify(configPayload, null, 2);

        // Click into the editor
        await editorLocator.click();

        // Select all text (Ctrl+A / Cmd+A) and delete it
        const modifier = process.platform === 'darwin' ? 'Meta' : 'Control';
        await page.keyboard.press(`${modifier}+a`);
        await page.keyboard.press('Backspace');

        // Type or paste the new config.
        // Since paste involves clipboard permissions which can be tricky in headles, we type it out.
        // Alternatively, use evaluate to set the value via React state, but typing is closer to user behavior.
        // Typing a large string might be slow, so we can paste by using `page.evaluate` to set clipboard data and then paste.
        // Actually, since Monaco sets `value` prop, let's just type it.
        await page.keyboard.insertText(configString);


        // Click Register Service
        await page.getByRole('button', { name: 'Register Service' }).click();

        // Verify toast notification
        await expect(page.getByText('Service Registered')).toBeVisible();

        // Verify the backend received the config by querying the API
        const response = await request.get(`/api/v1/services/${testServiceId}`);
        expect(response.ok()).toBeTruthy();

        const service = await response.json();
        const svcData = service.service || service; // handle different API response structures

        expect(svcData.name).toBe(testServiceId);
        expect(svcData.http_service).toBeDefined();
        expect(svcData.http_service.address).toBe("https://api.advanced-json-test.com");
    });
});
