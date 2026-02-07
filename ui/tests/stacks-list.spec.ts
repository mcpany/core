/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stacks List', () => {
  const stackName = 'e2e-test-stack-list';

  test.beforeEach(async ({ request }) => {
      await seedCollection(stackName, request);
  });

  test.afterEach(async ({ request }) => {
      await cleanupCollection(stackName, request);
  });

  test('should list the seeded stack', async ({ page }) => {
    await page.goto('/stacks');

    // Wait for the stack card to appear
    // The card should contain the stack name
    // Use .first() to avoid strict mode violation as name appears twice (title and id)
    await expect(page.getByText(stackName).first()).toBeVisible({ timeout: 10000 });

    // Optionally check for "Online" or "Services" count if possible
    // But visibility of name is enough for "The Strategic Pivot"
  });
});
