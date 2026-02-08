/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Stacks Management', () => {
  test.beforeEach(async ({ page }) => {
    // Perform authentication if not already done by global setup
    // Since we don't have a global setup file visible, we'll try to set the token manually
    // or navigate to login.
    // Assuming Playwright config sets X-API-Key header, that might be enough for direct requests
    // but client-side app uses localStorage.

    // Seed the token into localStorage
    await page.goto('/'); // Navigate to a page to set localStorage
    await page.evaluate(() => {
        localStorage.setItem('mcp_auth_token', 'test-token');
    });
  });

  test('should create and verify a new stack', async ({ page }) => {
    await page.goto('/stacks');

    // Verify the page title
    await expect(page.getByRole('heading', { name: 'Stacks', exact: true })).toBeVisible();

    // Create a new stack
    // Using first() because sometimes there might be hidden elements or multiple if responsiveness triggers
    await page.getByRole('button', { name: 'Create Stack' }).first().click();

    // Fill in name
    const stackName = 'e2e-test-stack-' + Date.now();
    await page.getByLabel('Name').fill(stackName);

    // Click Create
    await page.getByRole('button', { name: 'Create', exact: true }).click();

    // Verify redirect to editor
    // URL should contain encoded stack name
    await expect(page).toHaveURL(new RegExp(`/stacks/${encodeURIComponent(stackName)}`));

    // Verify editor loaded
    await expect(page.getByText('config.yaml')).toBeVisible();

    // Go back to list and verify it appears
    await page.goto('/stacks');
    await expect(page.getByText(stackName)).toBeVisible();
  });
});
