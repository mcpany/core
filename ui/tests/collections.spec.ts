/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Collections', () => {
  test('should create a collection, add a test case, and run it', async ({ page }) => {
    // 1. Navigate to Playground
    await page.goto('/playground');

    // 2. Open Collections Tab
    await page.click('button:has-text("Collections")');

    // 3. Create Collection
    await page.locator('.flex.items-center.justify-between', { hasText: 'Collections' })
              .locator('button')
              .click();

    await page.fill('input[placeholder*="Collection Name"]', 'E2E Suite');
    await page.click('button:text("Create")');

    // 4. Verify Collection Exists
    await expect(page.locator('button', { hasText: 'E2E Suite' })).toBeVisible();

    // 5. Switch to Library
    await page.click('button:has-text("Library")');

    const input = page.locator('input[placeholder="Enter command or select a tool..."]');
    await input.fill('weather-service.get_weather {}');

    // 6. Execute Tool
    await page.keyboard.press('Enter');

    const toolCallHeader = page.locator('.p-3.flex.flex-row.items-center', { hasText: 'Tool Execution' }).last();
    await expect(toolCallHeader).toBeVisible();

    // 7. Save to Collection
    await toolCallHeader.hover();
    await toolCallHeader.locator('button[aria-label="Save to Collection"]').click();

    // 8. Fill Save Dialog
    await expect(page.locator('div[role="dialog"]', { hasText: 'Save to Collection' })).toBeVisible();

    // Select E2E Suite if not selected (Seed data "My First Collection" might be first)
    await page.click('button[role="combobox"]');
    await page.click('div[role="option"]:has-text("E2E Suite")');

    await page.fill('input[placeholder="e.g. Valid Input Test"]', 'Weather Test');
    await page.click('button:text("Save")');

    // 8.5 Reload to verify persistence
    await page.reload();

    // 9. Verify Test Case in Collection
    await page.click('button:has-text("Collections")');

    const collectionBtn = page.locator('button', { hasText: 'E2E Suite' });
    await expect(collectionBtn).toBeVisible();

    await collectionBtn.click();

    const testCaseLabel = page.locator('span[title="Weather Test"]');
    await expect(testCaseLabel).toBeVisible();

    // 10. Run Test Case
    const container = page.locator('div.group\\/item:has(span[title="Weather Test"])');
    await container.hover();
    await container.locator('button[title="Run"]').click();

    // 11. Verify Input populated
    await expect(input).toHaveValue('weather-service.get_weather {}');
  });
});
