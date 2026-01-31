/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Stacks CRUD', () => {
  const seedStackName = 'e2e-stack-seed';
  const newStackName = 'e2e-new-stack';

  test.beforeEach(async ({ request }) => {
    // Ensure clean slate
    await cleanupCollection(seedStackName, request);
    await cleanupCollection(newStackName, request);

    // Seed one stack
    await seedCollection(seedStackName, request);
  });

  test.afterEach(async ({ request }) => {
    await cleanupCollection(seedStackName, request);
    await cleanupCollection(newStackName, request);
  });

  test('should list seeded stacks', async ({ page }) => {
    await page.goto('/stacks');

    // Check if the seeded stack card is visible
    const card = page.locator('.group', { hasText: seedStackName }).first();
    await expect(card).toBeVisible();
    // seedCollection adds 1 service, so we expect "1 Services"
    await expect(card.getByText('1 Services')).toBeVisible();
  });

  test('should create a new stack', async ({ page }) => {
    await page.goto('/stacks');

    // Click Create Stack button
    await page.getByRole('button', { name: 'Create Stack' }).first().click();

    // Fill form
    await page.getByLabel('Stack Name').fill(newStackName);

    // Submit
    await page.getByRole('button', { name: 'Create', exact: true }).click();

    // Verify redirect to editor
    await expect(page).toHaveURL(new RegExp(`/stacks/${newStackName}$`));

    // Go back to list
    await page.goto('/stacks');

    // Verify new stack is listed
    const newCard = page.locator('.group', { hasText: newStackName }).first();
    await expect(newCard).toBeVisible();
    await expect(newCard.getByText('0 Services')).toBeVisible(); // New stack is empty
  });
});
