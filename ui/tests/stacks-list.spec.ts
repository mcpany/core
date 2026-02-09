/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stacks List', () => {
  const stackName = 'e2e-test-stack';

  test.beforeEach(async ({ request }) => {
    // Ensure clean state
    await cleanupCollection(stackName, request);
    // Seed a stack
    await seedCollection(stackName, request);
  });

  test.afterEach(async ({ request }) => {
    await cleanupCollection(stackName, request);
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
    const newStackName = 'new-ui-stack';
    await cleanupCollection(newStackName, page.request);

    await page.goto('/stacks');
    await page.getByRole('button', { name: 'Create Stack' }).click();

    await page.getByLabel('Name').fill(newStackName);
    await page.getByLabel('Description').fill('Created via UI test');
    await page.getByRole('button', { name: 'Create', exact: true }).click();

    await expect(page.getByText(newStackName)).toBeVisible();
    await cleanupCollection(newStackName, page.request);
  });
});
