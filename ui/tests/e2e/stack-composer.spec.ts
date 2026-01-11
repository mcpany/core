/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Stack Composer', () => {

  // Mock the API response for getStackConfig
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/v1/stacks/*/config', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'text/plain',
        body: `version: "1.0"
services:
  weather-service:
    image: mcpany/weather-service:latest
`
      });
    });
  });
  test('should load the editor and visualize configuration', async ({ page }) => {
    // Navigate to a stack detail page
    await page.goto('/stacks/e2e-test-stack');
    await page.getByRole('tab', { name: 'Editor' }).click();

    // Verify Editor is loaded
    await expect(page.locator('text=config.yaml')).toBeVisible();

    // The component used fallback data: `weather-service`

    // Verify Visualizer shows the existing service
    // Visualizer renders card titles. Use precise selector to avoid matching textarea content.
    await expect(page.locator('span.truncate', { hasText: 'weather-service' })).toBeVisible({ timeout: 10000 });

    // "Valid Configuration" badge in visualizer
    await expect(page.locator('.stack-visualizer-container').getByText('Valid Configuration').first()).toBeVisible();
  });
  test('should insert template from palette', async ({ page }) => {
    await page.goto('/stacks/e2e-test-stack');

    // Ensure we are on Editor tab (page level)
    await page.getByRole('tab', { name: 'Editor' }).click();

    // Wait for the editor to load even the initial content to avoid race conditions
    // await expect(page.locator('.monaco-editor')).toContainText('Stack Configuration', { timeout: 15000 });

    // Verify the Side Palette is visible
    await expect(page.getByText('Service Palette')).toBeVisible({ timeout: 10000 });

    // Click a template (e.g., Redis)
    await page.getByText('Redis').first().click();

    // Verify Visualizer updates
    await expect(page.locator('.stack-visualizer-container').getByText('redis-cache')).toBeVisible({ timeout: 15000 });
  });

  test.skip('should validate invalid YAML', async ({ page }) => {
    await page.goto('/stacks/e2e-test-stack');

    // Ensure we are on Editor tab (page level)
    await page.getByRole('tab', { name: 'Editor' }).click();

    // Wait for editor to fully load initial content
    // await expect(page.locator('.monaco-editor')).toContainText('Stack Configuration', { timeout: 15000 });

    // Focus editor and inject invalid YAML at the top
    // await page.locator('.monaco-editor').click();
    // await page.keyboard.press('Control+Home');
    // await page.waitForTimeout(500);

    // Type definitely invalid syntax at the very beginning
    await page.keyboard.type('!!!! invalid !!!!\n', { delay: 100 });

    // Look for "Invalid YAML" badge - this is the most direct indicator
    await expect(page.getByText('Invalid YAML')).toBeVisible({ timeout: 20000 });

    // Also check for the error detail
    await expect(page.getByText('YAML Syntax Error')).toBeVisible({ timeout: 10000 });
  });
});
