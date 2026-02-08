/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stacks List', () => {
  const STACK_NAME = 'e2e-stack-test';

  test.beforeEach(async ({ request }) => {
      // Ensure clean state
      await cleanupCollection(STACK_NAME, request);
      await seedCollection(STACK_NAME, request);
  });

  test.afterEach(async ({ request }) => {
      await cleanupCollection(STACK_NAME, request);
  });

  test('should list created stacks', async ({ page }) => {
    await page.goto('/stacks');

    // Wait for the stack card to appear
    const stackCard = page.locator('.grid').getByText(STACK_NAME);
    await expect(stackCard).toBeVisible({ timeout: 10000 });

    // Verify service count (seeded collection has 1 service)
    // The card contains "1 Services"
    const serviceCount = page.locator('.grid').filter({ hasText: STACK_NAME }).getByText('1 Services');
    await expect(serviceCount).toBeVisible();
  });
});
