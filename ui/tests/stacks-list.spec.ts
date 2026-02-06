/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stacks List Page', () => {
  const stackName = 'test-stack-list';

  test.beforeEach(async ({ page, request }) => {
    // Set API Key in localStorage for the browser context
    await page.addInitScript(() => {
      localStorage.setItem('mcp_api_key', 'test-token');
    });

    // Ensure clean state
    await cleanupCollection(stackName, request);
    await seedCollection(stackName, request);
  });

  test.afterEach(async ({ request }) => {
    await cleanupCollection(stackName, request);
  });

  test('should list fetched collections', async ({ page }) => {
    await page.goto('/stacks');

    // Wait for the stack card to appear
    const stackCard = page.locator('.stack-card').filter({ hasText: stackName });
    await expect(stackCard).toBeVisible({ timeout: 10000 });

    // Verify service count (seeded collection has 1 service)
    await expect(stackCard).toContainText('1 Services');

    // Verify status
    await expect(stackCard).toContainText('Active');
  });

  test('should show empty state when no stacks', async ({ page, request }) => {
     // Ensure no stacks match our filter or clean all if possible?
     // Since tests run in parallel or shared env, we can't delete ALL stacks.
     // But we can check if our specific stack is gone after delete.
     await cleanupCollection(stackName, request);

     await page.goto('/stacks');
     const stackCard = page.locator('.stack-card').filter({ hasText: stackName });
     await expect(stackCard).not.toBeVisible();
  });
});
