/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('Render MCP Rich Content (Markdown & Image)', async ({ page }) => {
  // Go to tools page
  await page.goto('/tools');

  // Search for the tool
  const searchBox = page.getByPlaceholder('Search tools...');
  await searchBox.click();
  await searchBox.fill('get_rich_content');

  // Wait for the tool to appear
  // Use a more specific selector to ensure we find the exact tool
  const toolRow = page.locator('tr').filter({ hasText: 'get_rich_content' });
  await expect(toolRow).toBeVisible({ timeout: 10000 });

  // Open inspector
  await toolRow.getByRole('button', { name: 'Inspect' }).click();

  // Wait for dialog
  await expect(page.getByRole('dialog')).toBeVisible();

  // Execute
  await page.getByRole('button', { name: 'Execute' }).click();

  // Verify Rendered Tab is active
  // The tab list should contain "Rendered"
  const renderedTab = page.getByRole('tab', { name: 'Rendered' });
  await expect(renderedTab).toBeVisible({ timeout: 10000 });
  await expect(renderedTab).toHaveAttribute('data-state', 'active');

  // Verify Markdown Content
  // "Hello World" h1
  const prose = page.locator('.prose');
  await expect(prose.locator('h1')).toHaveText('Hello World');
  // "markdown" italic
  await expect(prose.locator('em')).toHaveText('markdown');
  // "test" bold
  await expect(prose.locator('strong')).toHaveText('test');

  // Verify Image Content
  const img = page.locator('img[alt="Tool Output"]');
  await expect(img).toBeVisible();
  await expect(img).toHaveAttribute('src', /^data:image\/png;base64,/);
});
