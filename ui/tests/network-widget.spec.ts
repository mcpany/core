/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('dashboard network topology widget', async ({ page }) => {
  // Go to dashboard
  await page.goto('/');

  // Ensure dashboard title is visible first
  await expect(page.getByText('Dashboard', { exact: true }).first()).toBeVisible({ timeout: 15000 });

  // The widget might take a moment to be added to the layout or rendered
  await expect(async () => {
    // If the widget is missing, try to add it via the "Add Widget" button
    const networkWidget = page.locator('.react-flow');
    if (!(await networkWidget.isVisible())) {
      const trigger = page.getByTestId('add-widget-trigger').first();
      if (await trigger.isVisible()) {
        await trigger.click();
        await page.getByText('Network Topology').first().click();
      }
    }
    // Check for the React Flow container
    await expect(page.locator('.react-flow')).toBeVisible({ timeout: 15000 });

    // Check for the presence of nodes
    await expect(page.locator('.react-flow__node').first()).toBeVisible({ timeout: 10000 });
  }).toPass({ timeout: 60000, intervals: [2000, 5000, 10000] });
});
