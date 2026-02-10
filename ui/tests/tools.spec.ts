/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Tool Management (Real Data)', () => {
    test('should list available tools from backend', async ({ page }) => {
        await page.goto('/tools');

        // Check for the builtin/config tool (weather-service -> get_weather)
        // This is always present in config.minimal.yaml
        await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 15000 });
        await expect(page.getByText('Get current weather').first()).toBeVisible();
    });

    test('should execute a tool and show results', async ({ page }) => {
        await page.goto('/tools');

        // Use weather-service get_weather which is defined in config.minimal.yaml
        const toolRow = page.locator('tr', { hasText: 'get_weather' });
        await toolRow.getByRole('button', { name: 'Inspect' }).click();

        // Check if Inspector is open
        await expect(page.getByRole('dialog')).toBeVisible();
        await expect(page.getByRole('heading', { name: 'get_weather' })).toBeVisible();

        // Click Execute
        await page.getByRole('button', { name: 'Execute' }).click();

        // Wait for execution to complete (spinner to disappear)
        await expect(page.locator('.lucide-loader-2')).toBeHidden({ timeout: 20000 });

        // Verify Result
        // The output is the result of `echo '{"weather": "sunny"}'`
        // It should contain "sunny" in the stdout field.
        const outputPre = page.locator('pre.text-xs.font-mono');
        await expect(outputPre).toBeVisible({ timeout: 20000 });

        // Debugging: Log content if check fails
        const text = await outputPre.textContent();
        console.log(`Tool Output: ${text}`);

        await expect(outputPre).toContainText('sunny', { timeout: 10000 });
    });
});
