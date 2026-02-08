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

    // Select the specific card link for the stack
    // The grid structure is div.grid > Link > Card
    // We target the Link containing the text
    const stackLink = page.locator('.grid > a').filter({ hasText: STACK_NAME });

    await expect(stackLink).toBeVisible({ timeout: 10000 });

    // Verify service count within that specific card
    await expect(stackLink.getByText('1 Services')).toBeVisible();
  });
});
