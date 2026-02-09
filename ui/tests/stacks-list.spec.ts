/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stacks List', () => {
  test.describe.configure({ mode: 'serial' });

  const stackName = 'e2e-test-stack';
  const newStackName = 'ui-created-stack';

  test.beforeEach(async ({ request }) => {
    // Ensure clean state
    await cleanupCollection(stackName, request);
    await cleanupCollection(newStackName, request);
    // Seed a stack
    await seedCollection(stackName, request);
  });

  test.afterEach(async ({ request }) => {
    await cleanupCollection(stackName, request);
    await cleanupCollection(newStackName, request);
  });

  test('should display seeded stack in the list', async ({ page }) => {
    await page.goto('/stacks');

    // The current hardcoded page has "mcpany-system".
    // We expect our seeded stack to be there.
    await expect(page.getByText(stackName)).toBeVisible();
  });

  test('should navigate to stack detail on click', async ({ page }) => {
    await page.goto('/stacks');
    await page.getByText(stackName).click();
    await expect(page).toHaveURL(new RegExp(`/stacks/${stackName}`));
  });

  test('should create a new stack via dialog', async ({ page }) => {
    await page.goto('/stacks');

    // Use specific role selector to avoid ambiguity
    await page.getByRole('button', { name: 'Create Stack' }).first().click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();

    await dialog.getByLabel('Name').fill(newStackName);
    await dialog.getByLabel('Description').fill('Created via UI test');

    // Click the submit button inside the dialog footer
    await dialog.getByRole('button', { name: 'Create Stack' }).click();

    // Dialog should close
    await expect(dialog).toBeHidden();

    // Verify stack appears in list
    await expect(page.getByText(newStackName)).toBeVisible();
    await expect(page.getByText('Created via UI test')).toBeVisible();
  });
});
