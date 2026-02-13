/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices } from './e2e/test-data';

test.describe('Playground Collections (Real Data)', () => {
  // Use seeded services (echo_tool) for robust testing

  test.beforeEach(async ({ request }) => {
      await seedServices(request);
  });

  test.afterEach(async ({ request }) => {
      await cleanupServices(request);
  });

  test('should allow saving requests to a collection and replaying them', async ({ page }) => {
    // 1. Navigate to Playground
    await page.goto('/playground');

    // 2. Wait for 'echo_tool' to appear in the library sidebar
    const sidebar = page.locator('.border-r', { hasText: 'Library' });
    await expect(sidebar.getByText('echo_tool')).toBeVisible({ timeout: 15000 });

    // 3. Select the tool
    await page.getByPlaceholder('Search tools...').fill('echo_tool');
    // Click the tool card in the sidebar
    await sidebar.getByText('echo_tool').click();

    // 4. Build Command
    // Opens the configuration dialog
    await expect(page.getByRole('dialog')).toBeVisible();
    await page.getByRole('button', { name: /build command/i }).click();

    // 5. Execute Tool
    const sendBtn = page.getByLabel('Send');
    await expect(sendBtn).toBeEnabled();
    await sendBtn.click();

    // 6. Verify Execution Result
    await expect(page.getByText('Tool Execution')).toBeVisible();
    // Use first() to avoid ambiguity if it appears multiple times
    await expect(page.getByText('echo_tool', { exact: true }).first()).toBeVisible();
    await expect(page.getByText('Result: echo_tool').first()).toBeVisible();

    // 7. Save to Collection
    // The save button is on the tool call card header.
    // Target the specific tool call card (last one)
    // The button has aria-label="Save to collection"
    const saveBtn = page.getByLabel('Save to collection').last();
    await expect(saveBtn).toBeVisible();
    await saveBtn.click();

    // 8. Configure Save Dialog
    await expect(page.getByRole('dialog', { name: 'Save Request' })).toBeVisible();
    await page.getByLabel('Name').fill('My Echo Test');
    // Note: The logic auto-creates "My Collection" if none exists.
    await page.getByRole('button', { name: 'Save' }).click();

    // 9. Verify Collections Sidebar
    // Switch tab
    await page.getByRole('button', { name: 'Collections' }).click();

    // Check Collection exists
    const collectionsSidebar = page.locator('.border-r', { hasText: 'Collections' });
    await expect(collectionsSidebar.getByText('My Collection')).toBeVisible();

    // Expand/Check Request exists (it might be auto-expanded)
    await expect(collectionsSidebar.getByText('My Echo Test')).toBeVisible();

    // 10. Run from Collection
    // First clear the chat to be sure we see new execution
    await page.getByRole('button', { name: 'Clear' }).click();
    await expect(page.queryByText('Result: echo_tool')).toBeNull();

    // Click Run on the saved request item
    const requestRow = collectionsSidebar.locator('div', { hasText: 'My Echo Test' }).first();
    // Hover to reveal button
    await requestRow.hover();
    await requestRow.locator('button[title="Run"]').click();

    // 11. Verify Re-execution
    // Should see the result again
    await expect(page.getByText('Result: echo_tool')).toBeVisible();
  });
});
