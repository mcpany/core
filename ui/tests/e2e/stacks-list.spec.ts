/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './test-data';

const TEST_STACK_NAME = 'e2e-test-stack';
const NEW_STACK_NAME = 'new-created-stack';

test.describe('Stacks List', () => {
  test.beforeEach(async ({ request }) => {
      // Ensure clean state
      await cleanupCollection(TEST_STACK_NAME, request);
      await cleanupCollection(NEW_STACK_NAME, request);
      // Seed a stack
      await seedCollection(TEST_STACK_NAME, request);
  });

  test.afterEach(async ({ request }) => {
      await cleanupCollection(TEST_STACK_NAME, request);
      await cleanupCollection(NEW_STACK_NAME, request);
  });

  test('should display seeded stack', async ({ page }) => {
    await page.goto('/stacks');

    // Check for the stack card
    // Note: The card displays name and service count
    await expect(page.getByText(TEST_STACK_NAME)).toBeVisible();
    // seedCollection adds 1 service, so we expect "1 Services"
    await expect(page.getByText('1 Services')).toBeVisible();
  });

  test('should create a new stack', async ({ page }) => {
    await page.goto('/stacks');

    // Open Dialog
    await page.getByRole('button', { name: 'Create Stack' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    // Fill Name
    await page.getByLabel('Name').fill(NEW_STACK_NAME);

    // Submit
    // Use exact match to avoid matching "Create Stack" trigger button if visible
    await page.getByRole('button', { name: 'Create', exact: true }).click();

    // Expect Dialog to close
    await expect(page.getByRole('dialog')).toBeHidden();

    // Expect new stack to appear
    // Use locator to avoid matching the toast notification
    await expect(page.locator('.text-2xl', { hasText: NEW_STACK_NAME })).toBeVisible();
    // Empty stack should have 0 services
    await expect(page.getByText('0 Services')).toBeVisible();
  });
});
