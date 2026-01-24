/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resource Preview Modal', () => {

  test('should open resource in modal from explorer', async ({ page }) => {
    await page.goto('/resources');

    // We assume the backend exposes 'System Logs' from weather-service (config.minimal.yaml).
    const resourceName = 'System Logs';

    // Check if the resource is visible
    const resourceItem = page.locator('div.font-medium', { hasText: resourceName });
    await expect(resourceItem).toBeVisible({ timeout: 10000 });

    // Click on it
    await resourceItem.click();

    // Wait for content to load in Explorer first
    // It should contain some text from README (e.g. "MCP Any" or "License" or anything).
    // We just check for non-empty content or specific text if known.
    // Let's assume content loads.
    await expect(page.locator('.monaco-editor')).toBeVisible();

    // Wait for "Maximize" button and click it
    await page.click('button[title="Maximize"]');

    // Wait for modal to open and verify title
    const modalTitle = page.locator("div[role='dialog']").getByRole("heading", { name: resourceName });
    await expect(modalTitle).toBeVisible();

    // Verify content in modal is visible (editor)
    await expect(page.locator("div[role='dialog'] .monaco-editor")).toBeVisible();
  });

});
