/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('playground loads and handles user interaction', async ({ page }) => {
  await page.goto('/playground');

  // Check for title
  await expect(page.getByRole('heading', { name: 'Playground' })).toBeVisible();

  // Type a message
  const input = page.getByPlaceholder('Type a message to interact with your tools...');
  await input.fill('list files');
  await input.press('Enter');

  // Wait for the tool call card (mocked response)
  // We use a slightly longer timeout because of the mock delays
  await expect(page.getByText('Calling Tool: list_files')).toBeVisible({ timeout: 10000 });

  // Wait for the result
  await expect(page.getByText('Tool Output (list_files)')).toBeVisible({ timeout: 10000 });

  // Wait for final response
  await expect(page.getByText("I've listed the files for you")).toBeVisible({ timeout: 10000 });
});
