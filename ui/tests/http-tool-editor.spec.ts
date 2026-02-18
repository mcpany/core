/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('HTTP Tool Editor - Live Test', () => {
  const serviceName = 'e2e-http-tool-preview';

  test.beforeAll(async ({ request }) => {
    // Seed HTTP service
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});

    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        http_service: {
            address: "https://httpbin.org" // Use a real echo service for execution test
        }
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should preview request and execute tool', async ({ page }) => {
    // Navigate to upstream services
    await page.goto('/upstream-services');

    // Wait for list to load
    await expect(page.getByText(serviceName)).toBeVisible();

    // Find row, click edit.
    const row = page.getByRole('row', { name: serviceName });
    await row.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();

    // Wait for Sheet to open.
    await expect(page.getByRole('dialog', { name: 'Edit Service' })).toBeVisible();

    // Click "Tools" tab.
    await page.getByRole('tab', { name: 'Tools' }).click();

    // Click "Add Tool"
    await page.getByRole('button', { name: 'Add Tool' }).click();

    // Sheet for Tool Editor should open
    await expect(page.getByText('Edit new_tool')).toBeVisible();

    // Fill details
    await page.getByLabel('Tool Name').fill('get_uuid');
    await page.getByLabel('Description').fill('Get UUID');

    // Call details
    await page.getByLabel('Endpoint Path').fill('/uuid');

    // Go to "Test & Preview" tab
    await page.getByRole('tab', { name: 'Test & Preview' }).click();

    // Verify Preview shows the correct path
    // Limit scope to the preview card
    const previewCard = page.locator('.space-y-4 > .h-full');
    await expect(previewCard.getByText('/uuid')).toBeVisible();
    await expect(previewCard.getByText('GET')).toBeVisible();

    // Close Tool Editor (Escape) - this saves to local state
    await page.keyboard.press('Escape');

    // Save Service - this persists to backend
    await page.getByRole('button', { name: 'Save Changes' }).click();
    await expect(page.getByText('Service Updated')).toBeVisible();

    // Re-open Service Editor (since Save closes it)
    const row2 = page.getByRole('row', { name: serviceName });
    await row2.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();

    await page.getByRole('tab', { name: 'Tools' }).click();

    // Find the tool 'get_uuid' and click edit icon
    const toolCard = page.locator('.space-y-4 > .grid > div').filter({ hasText: 'get_uuid' });
    await toolCard.getByRole('button').nth(0).click();

    // Go to Test Tab
    await page.getByRole('tab', { name: 'Test & Preview' }).click();

    // Verify Preview
    await expect(previewCard.getByText('/uuid')).toBeVisible();

    // Execute
    await page.getByRole('button', { name: 'Execute Tool' }).click();

    // Verify Result
    // Wait for the header "Execution Result" to appear anywhere
    const resultHeader = page.getByText('Execution Result');
    await expect(resultHeader).toBeVisible({ timeout: 15000 });

    // Get content
    const preTag = page.locator('pre.text-green-400');
    await expect(preTag).toBeVisible();
    const content = await preTag.textContent();
    console.log("Execution Result Content:", content);

    expect(content).toContain('uuid');
  });
});
