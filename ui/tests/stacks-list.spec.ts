/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stacks List', () => {
  const stackName = 'e2e-stack-test';

  test.beforeEach(async ({ request }) => {
      // Ensure clean state
      await cleanupCollection(stackName, request);
      await seedCollection(stackName, request);
  });

  test.afterEach(async ({ request }) => {
      await cleanupCollection(stackName, request);
  });

  test('should list created stacks', async ({ page }) => {
    await page.goto('/stacks');

    // Wait for loading to finish (Loader2 should disappear)
    await expect(page.locator('.lucide-loader-2')).toBeHidden();

    // Check if the stack card is visible
    const stackCard = page.locator('.grid').getByText(stackName).first();
    await expect(stackCard).toBeVisible();

    // Check service count (seedCollection adds 1 service)
    await expect(page.getByText('1 Services')).toBeVisible();
  });
});
