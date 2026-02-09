/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stack Editor', () => {
  test.beforeEach(async ({ request }) => {
      await seedCollection('default-stack', request);
  });

  test.afterEach(async ({ request }) => {
      await cleanupCollection('default-stack', request);
      await cleanupCollection('new-stack', request);
  });

  test('should list stacks', async ({ page }) => {
    await page.goto('/stacks');
    await expect(page.getByRole('heading', { name: 'Stacks' })).toBeVisible();
    await expect(page.getByText('default-stack')).toBeVisible();
  });

  test('should load the editor for existing stack', async ({ page }) => {
    await page.goto('/stacks/default-stack');
    await expect(page.getByRole('heading', { name: 'Edit Stack: default-stack' })).toBeVisible();
    // Verify YAML editor contains the name
    // Monaco editor content is hard to test directly, but we can check if it loaded without error
    await expect(page.locator('.monaco-editor')).toBeVisible();
  });

  test('should create a new stack', async ({ page }) => {
    await page.goto('/stacks/new');
    await expect(page.getByRole('heading', { name: 'New Stack' })).toBeVisible();

    await page.fill('input[id="name"]', 'new-stack');
    await page.fill('textarea[id="description"]', 'My new stack');

    // We keep default YAML which has "my-stack".
    // Ideally we update YAML to match name, but let's see if save works with default YAML (which has mismatched name).
    // The backend might error "Stack name in config must match URL path" or simply use the ID from URL (which is 'new-stack' for create?).
    // In create mode (POST /collections), the body determines the name.
    // My StackEditor uses `saveStackConfig(stackId || name, parsed)`.
    // If I enter 'new-stack' in input, `name` state is 'new-stack'.
    // YAML has 'my-stack'.
    // `parsed.name = name` sets it to 'new-stack'.
    // So it should work!

    await page.getByRole('button', { name: 'Deploy Stack' }).click();

    // Should redirect to list
    await expect(page).toHaveURL(/\/stacks$/);
    await expect(page.getByText('new-stack')).toBeVisible();
  });
});
