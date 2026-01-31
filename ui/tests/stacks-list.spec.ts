/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stacks List Page', () => {
  const stackName = 'e2e-test-stack';
  const newStackName = 'new-stack';

  test.beforeEach(async ({ request }) => {
      await seedCollection(stackName, request);
  });

  test.afterEach(async ({ request }) => {
      await cleanupCollection(stackName, request);
      await cleanupCollection(newStackName, request);
  });

  test('should display seeded stack and allow creating new stack', async ({ page }) => {
    // 1. Visit Stacks List
    await page.goto('/stacks');

    // 2. Verify seeded stack is visible
    // Using first() because the name appears in both title and ID
    await expect(page.getByText(stackName).first()).toBeVisible();

    // 3. Create New Stack
    await page.getByRole('button', { name: 'Create Stack' }).click();

    // Wait for Dialog
    await expect(page.getByRole('dialog')).toBeVisible();

    // Fill Name
    // We assume the input will have a label "Stack Name" or placeholder
    await page.getByLabel('Stack Name').fill(newStackName);

    // Click Create
    await page.getByRole('button', { name: 'Create', exact: true }).click();

    // 4. Verify Redirection
    await expect(page).toHaveURL(`/stacks/${newStackName}`);

    // 5. Verify Back to List
    await page.goto('/stacks');
    await expect(page.getByText(newStackName).first()).toBeVisible();
  });
});
