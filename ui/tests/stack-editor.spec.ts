/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stack Editor', () => {
  test.beforeEach(async ({ request }) => {
      await seedCollection('default-stack', request);
      // Wait a bit for potential backend sync (though seedCollection awaits response)
  });

  test.afterEach(async ({ request }) => {
      await cleanupCollection('default-stack', request);
  });

  test('should load the YAML editor', async ({ page }) => {
    await page.goto('/stacks/default-stack');

    // Check for Monaco Editor
    await expect(page.locator('.monaco-editor')).toBeVisible({ timeout: 30000 });

    // Check for Save button
    await expect(page.getByRole('button', { name: 'Save & Deploy' })).toBeVisible();
  });
});
