/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices } from './test-data';

test.describe('Skill Wizard', () => {
    test.beforeEach(async ({ page, request }) => {
        await seedServices(request);
        await page.goto('/skills/create');
    });

    test.afterEach(async ({ request }) => {
        // Cleanup the skill we created
        // We need a way to cleanup skill. apiClient.deleteSkill?
        // We can just rely on fresh environment or manual cleanup if needed.
        // For now just clean services.
        await cleanupServices(request);
    });

    test('should allow selecting tools using the new ToolSelector', async ({ page }) => {
        // Step 1: Metadata
        await page.fill('input[id="name"]', 'test-skill');
        await page.fill('textarea[id="description"]', 'A test skill description');

        // Verify ToolSelector is present and works
        const toolSelector = page.getByRole('combobox');
        await expect(toolSelector).toBeVisible();
        await expect(toolSelector).toContainText('Select tools...');

        // Open ToolSelector
        await toolSelector.click();

        // Check if "process_payment" (from seedServices) is visible in the list
        // Note: seedServices creates "Payment Gateway" with id "svc_01"
        const paymentOption = page.getByRole('option', { name: 'process_payment' });
        await expect(paymentOption).toBeVisible();

        // Select it
        await paymentOption.click();

        // Verify it is selected (Badge appears)
        // We look for the badge text
        await expect(page.getByText('process_payment', { exact: true })).toBeVisible();
        await expect(toolSelector).toContainText('1 tool selected');

        // Select another one "get_user" from "User Service" (svc_02)
        await toolSelector.click();
        const userOption = page.getByRole('option', { name: 'get_user' });
        await expect(userOption).toBeVisible();
        await userOption.click();

        await expect(page.getByText('get_user', { exact: true })).toBeVisible();
        await expect(toolSelector).toContainText('2 tools selected');

        // Remove one via badge close button
        // The close button has accessible name "Remove {toolName}"
        await page.getByRole('button', { name: 'Remove get_user' }).click();

        await expect(page.getByText('get_user', { exact: true })).toBeHidden();
        await expect(toolSelector).toContainText('1 tool selected');

        // Continue Wizard
        await page.getByRole('button', { name: 'Next' }).click();

        // Step 2: Instructions
        await expect(page.getByText('Step 2: Instructions')).toBeVisible();
        await page.getByRole('button', { name: 'Next' }).click();

        // Step 3: Assets
        await expect(page.getByText('Step 3: Assets')).toBeVisible();

        // Save
        await page.getByRole('button', { name: 'Create Skill' }).click();

        // Verify redirection to /skills
        await expect(page).toHaveURL(/\/skills$/);

        // Verify skill is listed
        await expect(page.getByText('test-skill')).toBeVisible();
    });
});
