/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('Tools render MCP Rich Content correctly', async ({ page }) => {
  // 1. Go to Tools page
  await page.goto('/tools');

  // 2. Search for the tool
  const searchInput = page.getByPlaceholder('Search tools...');
  await searchInput.fill('get_rich_content');

  // 3. Wait for the tool card to appear and click Inspect
  // Note: Depending on layout (table vs grid), the selector might need adjustment.
  // Assuming table layout as per default.
  // The row contains the tool name.
  await expect(page.getByRole('cell', { name: 'rich-result-test-service.get_rich_content', exact: true })).toBeVisible();

  // Find the button within the row/card
  // We can just click the first "Inspect" button if we filtered correctly,
  // or scope it to the container.
  await page.getByRole('button', { name: 'Inspect' }).first().click();

  // 4. Wait for Inspector dialog
  const dialog = page.getByRole('dialog');
  await expect(dialog).toBeVisible();
  // Check title exists (use first to avoid strict mode if multiple headers present)
  await expect(dialog.getByRole('heading').filter({ hasText: 'get_rich_content' }).first()).toBeVisible();

  // 5. Execute the tool
  await dialog.getByRole('button', { name: 'Execute' }).click();

  // 6. Verify "Rendered" tab is active and content is displayed
  // Wait for the "Rendered" tab trigger to be visible
  const renderedTabTrigger = dialog.getByRole('tab', { name: 'Rendered' });
  await expect(renderedTabTrigger).toBeVisible();
  await expect(renderedTabTrigger).toHaveAttribute('data-state', 'active');

  // 7. Verify Content
  // H1 header
  await expect(dialog.locator('h1', { hasText: 'Rich Content Test' })).toBeVisible();
  // Bold text
  await expect(dialog.locator('strong', { hasText: 'markdown' })).toBeVisible();
  // List item
  await expect(dialog.locator('li', { hasText: 'Item 1' })).toBeVisible();
  // Image
  await expect(dialog.locator('img[alt="Tool Result Image"]')).toBeVisible();
});
