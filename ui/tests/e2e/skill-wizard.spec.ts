/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Skill Wizard', () => {
  test.beforeEach(async ({ page }) => {
    // Mock the tools API to return expected data
    // This ensures the test validates the UI component logic independent of backend state
    await page.route('**/api/v1/tools', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          tools: [
            {
              name: 'process_payment',
              description: 'Process a payment',
              serviceId: 'Payment Gateway',
            },
            {
              name: 'get_user',
              description: 'Get user details',
              serviceId: 'User Service',
            },
          ],
        }),
      });
    });

    await page.goto('/skills/create');
  });

  test('should allow selecting tools using the new ToolSelector', async ({ page }) => {
    // Step 1: Metadata
    await page.fill('input[id="name"]', 'test-skill');
    await page.fill('textarea[id="description"]', 'A test skill description');

    // Verify ToolSelector is present and works
    // Target specific combobox by id or label since Command component might contain another combobox input
    const toolSelector = page.getByRole('combobox', { name: 'Allowed Tools' });
    await expect(toolSelector).toBeVisible();
    await expect(toolSelector).toContainText('Select tools...');

    // Open ToolSelector
    await toolSelector.click();

    // Check if "process_payment" is visible in the list
    const paymentOption = page.getByRole('option', { name: 'process_payment' });
    await expect(paymentOption).toBeVisible();

    // Select it
    await paymentOption.click();

    // Verify it is selected (Badge appears)
    // We check for the presence of the remove button which is unique to the badge
    await expect(page.getByRole('button', { name: 'Remove process_payment' })).toBeVisible();
    await expect(toolSelector).toContainText('1 tool selected');

    // Select another one "get_user"
    // Popover might stay open or close depending on implementation, but we can click trigger again if needed.
    // If it's already open, clicking might close it.
    // Let's ensure it's open.
    if (await toolSelector.getAttribute('aria-expanded') === 'false') {
        await toolSelector.click();
    }

    const userOption = page.getByRole('option', { name: 'get_user' });
    await expect(userOption).toBeVisible();
    await userOption.click();

    await expect(page.getByRole('button', { name: 'Remove get_user' })).toBeVisible();
    await expect(toolSelector).toContainText('2 tools selected');

    // Remove one via badge close button
    await page.getByRole('button', { name: 'Remove get_user' }).click();

    await expect(page.getByRole('button', { name: 'Remove get_user' })).toBeHidden();
    await expect(toolSelector).toContainText('1 tool selected');

    // Continue Wizard
    // Use exact match to avoid conflict with Next.js dev tools
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // Step 2: Instructions
    await expect(page.getByText('Step 2: Instructions')).toBeVisible();
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // Step 3: Assets
    await expect(page.getByText('Step 3: Assets')).toBeVisible();

    // Note: We don't click "Create Skill" here because it would try to hit the real backend
    // which might fail if not running. The goal is to test the ToolSelector UX.
  });
});
