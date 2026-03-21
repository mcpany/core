/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resource Preview Modal', () => {

  test('should open resource in modal from explorer', async ({ page }) => {
    // Mock resources list
    await page.route('**/api/v1/resources', async route => {
        await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
                resources: [{ uri: 'file:///test.json', name: 'test.json', mimeType: 'application/json' }]
            })
        });
    });

    // Mock resource read with regex to handle encoded URI
    await page.route(/\/api\/v1\/resources\/read.*/, async route => {
        await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
                contents: [{
                    uri: 'file:///test.json',
                    mimeType: 'application/json',
                    text: '{"key": "value", "long": "content to test modal view"}'
                }]
            })
        });
    });

    await page.goto('/resources');

    // Wait for resources to load and click
    // Use first() to avoid ambiguity or specific class selector
    const resourceItem = page.locator('div.font-medium', { hasText: 'test.json' });
    await expect(resourceItem).toBeVisible();

    // Click on the resource to select it
    await resourceItem.click();

    // Wait for the inline preview to render before opening the modal.
    // RichResultViewer wraps the content, so we look for the JSON button/tab.
    const inlinePreviewTab = page.getByRole('button', { name: 'JSON' }).first();
    await expect(inlinePreviewTab).toBeVisible();
    // In RichResultViewer tree view, strings are quoted and separated from keys
    await expect(page.getByText('"content to test modal view"').first()).toBeVisible();

    // Wait for "Maximize" button and click it
    await page.click('button[title="Maximize"]');

    // Wait for modal to open and verify title
    const modalTitle = page.locator("div[role='dialog']").getByRole("heading", { name: "test.json" });
    await expect(modalTitle).toBeVisible();

    // Verify content in modal
    const modalContentTab = page.locator("div[role='dialog']").getByRole('button', { name: 'JSON' }).first();
    await expect(modalContentTab).toBeVisible();
    await expect(page.locator("div[role='dialog']").getByText('"content to test modal view"').first()).toBeVisible();
  });

});
