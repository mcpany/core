/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from './test-data';

test.describe('Stacks (Real Data)', () => {
  const stackName = 'e2e-real-stack';

  test.beforeAll(async ({ request }) => {
    // Seed a real stack in the backend
    await seedCollection(stackName, request);
  });

  test.afterAll(async ({ request }) => {
    // Clean up
    await cleanupCollection(stackName, request);
  });

  test('should display real stacks and allow deletion', async ({ page }) => {
    await page.goto('/stacks');

    // 1. Verify the seeded stack is visible
    const stackCard = page.locator('.card').filter({ hasText: stackName });
    await expect(stackCard).toBeVisible({ timeout: 10000 });

    // Verify some details (assuming seedCollection creates a stack with 1 service)
    await expect(stackCard).toContainText('1 Services');

    // 2. Verify the hardcoded "mcpany-system" is NOT visible (assuming I removed it)
    // Initially this might fail if I haven't removed it yet, but that's the point of TDD.
    // The test expects the final state.
    await expect(page.getByText('mcpany-system')).not.toBeVisible();

    // 3. Delete the stack
    // Find the delete button within the card.
    // Assuming there's a delete button. The mock didn't have one, but I will add it.
    // I'll look for a trash icon or "Delete" text.
    // The previous implementation was a Link wrapping the whole card.
    // I will implementation a specific button.

    // We might need to hover or just click.
    // If the whole card is a link, clicking the delete button requires e.stopPropagation() in the implementation.

    // For this test, I assume I will implement a Delete button.
    // If it's not there yet, the test fails (Red).

    // Wait, if I haven't implemented it, this locator won't find anything.
    // I'll skip the delete interaction in the "Red" phase check effectively, or assert it exists.
    // Use a more generic selector for the icon button inside the footer
    // Since there is only one button in the card (the delete button), getByRole should work
    const deleteBtn = stackCard.getByRole('button');

    await expect(deleteBtn).toBeVisible();

    // Setup dialog handling if I use a confirm dialog
    // We implemented AlertDialog, so we need to click the action button in the dialog
    await deleteBtn.click();

    const confirmBtn = page.getByRole('button', { name: 'Delete' }).filter({ hasText: 'Delete' }).last();
    await expect(confirmBtn).toBeVisible();
    await confirmBtn.click();

    // 4. Verify it disappears
    await expect(stackCard).not.toBeVisible();
  });
});
