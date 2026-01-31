/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Stacks Management', () => {
  const stackName = `e2e-stack-${Date.now()}`;

  test('should create, edit, deploy and delete a stack', async ({ page }) => {
    // 1. Navigate to Stacks page
    await page.goto('/stacks');
    await expect(page.getByRole('heading', { name: 'Stacks' })).toBeVisible();

    // 2. Create Stack
    await page.getByRole('button', { name: 'Create Stack' }).click();
    await page.getByLabel('Name').fill(stackName);
    await page.getByRole('button', { name: 'Create', exact: true }).click();

    // 3. Verify redirection to Detail page
    await expect(page).toHaveURL(new RegExp(`/stacks/${stackName}`));
    await expect(page.getByRole('heading', { name: stackName })).toBeVisible();

    // 4. Edit Stack
    await page.getByRole('tab', { name: 'Editor' }).click();
    // Ensure editor is loaded
    await expect(page.getByText('config.yaml')).toBeVisible();

    // Save (even if empty/default) to verify backend persistence
    await page.getByRole('button', { name: 'Save Changes' }).click();
    await expect(page.getByText('Configuration saved successfully')).toBeVisible();

    // 5. Deploy
    await page.getByRole('button', { name: 'Deploy Stack' }).click();
    await expect(page.getByText('Stack deployment initiated')).toBeVisible();

    // 6. Delete
    page.on('dialog', dialog => dialog.accept());
    // The delete button is an icon button with trash icon.
    // It might not have text. We can target by the icon or class, or title/aria-label if I added it.
    // I didn't add aria-label but it's a ghost button with Trash2.
    // Let's assume it's the last button in the header.
    // Or add aria-label in a followup?
    // I'll target by locator `button:has(.lucide-trash-2)` or similar if possible, or just the order.
    // In my code:
    // <Button variant="ghost" size="icon" onClick={handleDelete} ...> <Trash2 ... /> </Button>
    // I'll try to find it by icon class if visible, or just use `page.locator('.lucide-trash-2').click()`.
    await page.locator('.lucide-trash-2').click();

    // Wait for navigation back to list
    await expect(page).toHaveURL(/\/stacks$/);

    // Verify it's gone
    await expect(page.getByText(stackName)).not.toBeVisible();
  });
});
