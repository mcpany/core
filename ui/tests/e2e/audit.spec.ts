/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Audit Logs Viewer', () => {
  test.beforeEach(async ({ page, request }) => {
    // Seed global state (users, etc)
    await seedGlobalState(request);

    // Mock API for /api/v1/audit/logs to return a rich result.
    await page.route('**/api/v1/audit/logs*', async route => {
      await route.fulfill({
        json: {
          entries: [
            {
              timestamp: new Date().toISOString(),
              toolName: 'complex_tool',
              userId: 'e2e-user',
              profileId: 'default',
              arguments: JSON.stringify({ prompt: "hello" }),
              result: JSON.stringify({
                content: [
                  {
                    type: "text",
                    text: "# Rich Markdown Result\n\nThis is a **markdown** response from a simulated tool."
                  }
                ]
              }),
              error: '',
              duration: '1.2s',
              durationMs: 1200
            }
          ]
        }
      });
    });

    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.fill('input[name="username"]', 'e2e-admin-core');
    await page.fill('input[name="password"]', 'password');
    await Promise.all([
      page.waitForURL('/', { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true })
    ]);
    await expect(page).toHaveURL('/', { timeout: 15000 });
  });

  test('should render rich results in tabs', async ({ page }) => {
    await page.goto('/audit');
    await page.waitForLoadState('networkidle');

    // Ensure the table loaded with the mock data
    const firstRow = page.locator('table tbody tr').first();
    await expect(firstRow).toBeVisible();
    await expect(firstRow).toContainText('complex_tool');

    // Click "View"
    await firstRow.getByRole('button', { name: 'View' }).click();

    // The dialog should open
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
    await expect(dialog).toContainText('Audit Log Detail');
    await expect(dialog).toContainText('complex_tool');

    // Check if the tabs exist
    const tabList = page.getByRole('tablist');
    await expect(tabList).toBeVisible();

    const overviewTab = page.getByRole('tab', { name: 'Overview' });
    const resultTab = page.getByRole('tab', { name: 'Result' });
    await expect(overviewTab).toBeVisible();
    await expect(resultTab).toBeVisible();

    // Check Overview Tab is default
    await expect(overviewTab).toHaveAttribute('aria-selected', 'true');
    await expect(dialog.getByText('User ID')).toBeVisible();

    // Click Result Tab
    await resultTab.click();
    await expect(resultTab).toHaveAttribute('aria-selected', 'true');

    // Verify RichResultViewer rendered markdown
    // Look for the rendered markdown heading
    const markdownHeading = dialog.locator('h1', { hasText: 'Rich Markdown Result' });
    await expect(markdownHeading).toBeVisible();

    const markdownParagraph = dialog.locator('p', { hasText: 'This is a markdown response' });
    await expect(markdownParagraph).toBeVisible();
  });
});
