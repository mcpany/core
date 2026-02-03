/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Service Management (Real Data)', () => {
    test.beforeAll(async () => {
        // We use a static config file (server/config.minimal.yaml) for this test to ensure stability.
        // PaymentGateway is pre-configured.
    });

    test('should display rich tables for tools, prompts, and resources', async ({ page }) => {
        // Navigate to Service Detail
        // "PaymentGateway" is seeded in config.minimal.yaml
        await page.goto('/service/PaymentGateway');

        // Verify Title
        try {
             await expect(page.locator('.text-3xl', { hasText: 'PaymentGateway' })).toBeVisible();
        } catch (e) {
             const errorAlert = page.locator('.text-destructive'); // Check for error alert text
             if (await errorAlert.count() > 0) {
                 console.log("Found error alert:", await errorAlert.textContent());
             }
             throw e;
        }

        // 1. Verify Tools Table
        await expect(page.getByText('Tools', { exact: true })).toBeVisible();
        await expect(page.getByText('process_payment')).toBeVisible();

        // Search for tool
        const toolSearch = page.getByPlaceholder('Search tools...');
        await expect(toolSearch).toBeVisible();
        await toolSearch.fill('process');
        await expect(page.getByText('process_payment')).toBeVisible();
        await toolSearch.fill('nonexistent');
        await expect(page.getByText('No tools found matching "nonexistent"')).toBeVisible();
        await toolSearch.clear();

        // Toggle tool check
        const toolRow = page.locator('tr', { hasText: 'process_payment' });
        const toolSwitch = toolRow.getByRole('switch');
        await expect(toolSwitch).toBeVisible();
        await expect(toolSwitch).toBeChecked();
        await toolSwitch.click();
        await expect(toolSwitch).not.toBeChecked();


        // 2. Verify Prompts Table
        await expect(page.getByText('Prompts', { exact: true })).toBeVisible();
        await expect(page.getByText('confirm_payment')).toBeVisible();
        // Check arguments badge
        await expect(page.getByText('amount')).toBeVisible();

        // Search Prompt
        const promptSearch = page.getByPlaceholder('Search prompts...');
        await promptSearch.fill('confirm');
        await expect(page.getByText('confirm_payment')).toBeVisible();


        // 3. Verify Resources Table
        await expect(page.getByText('Resources', { exact: true })).toBeVisible();
        await expect(page.getByText('invoice_template')).toBeVisible();
        await expect(page.getByText('file:///templates/invoice.txt')).toBeVisible();
        await expect(page.getByText('text/plain')).toBeVisible();

        // View Action
        const resourceRow = page.locator('tr', { hasText: 'invoice_template' });
        await expect(resourceRow.getByRole('link', { name: 'View' })).toBeVisible();
    });
});
