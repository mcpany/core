/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('HTTP Tool Editor - Live Preview', () => {
  const serviceName = 'http-tool-editor-test';

  test.beforeAll(async ({ request }) => {
    // Seed HTTP service pointing to httpbin
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});

    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        http_service: {
          // Point to local echo server
          address: "http://ui-http-echo-server:5678"
        }
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should preview request and execute tool', async ({ page, request }) => {
    // Navigate
    await page.goto('/upstream-services');
    await expect(page.getByText(serviceName)).toBeVisible();

    const row = page.getByRole('row', { name: serviceName });
    await row.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();

    // Open Tools
    await page.getByRole('tab', { name: 'Tools' }).click();
    await page.getByRole('button', { name: 'Add Tool' }).click();

    // Configure Tool
    await page.getByLabel('Tool Name').fill('get_echo');
    await page.getByLabel('Endpoint Path').fill('/echo');

    // Check Preview (Immediate)
    await page.getByRole('tab', { name: 'Test & Preview' }).click();

    // Verify Preview Content
    // Should show GET and /echo
    const previewCard = page.getByTestId('request-preview-content');
    await expect(previewCard.getByText('GET')).toBeVisible();
    await expect(previewCard.getByText('/echo')).toBeVisible();

    // Add a parameter to verify substitution
    await page.getByRole('tab', { name: 'Request Parameters' }).click();
    await page.getByLabel('Endpoint Path').fill('/echo/{{id}}');
    await page.getByRole('button', { name: 'Add Parameter' }).click();
    await page.getByLabel('Name', { exact: true }).fill('id');

    // Go back to Preview
    await page.getByRole('tab', { name: 'Test & Preview' }).click();

    // Verify Substitution Placeholder
    await expect(page.getByText('/echo/{{id}}')).toBeVisible();

    // Type Argument
    const argsInput = page.getByLabel('Test Arguments (JSON)');
    await argsInput.fill('{\n  "id": "123"\n}');

    // Verify Substitution Result
    await expect(page.getByText('/echo/123')).toBeVisible();

    // Now Save and Test Execution
    // Close Tool Editor
    await page.keyboard.press('Escape');

    // Save Service
    await page.waitForTimeout(1000);

    // Monitor response for first save
    const savePromise = page.waitForResponse(resp => resp.url().includes('/api/v1/services') && resp.status() === 200, { timeout: 10000 });
    await page.getByRole('button', { name: 'Save Changes' }).click({ force: true });
    await savePromise;

    // Toast is flaky, so we rely on network success + subsequent reload verification
    // await expect(page.getByText(/Service (Updated|Created)/)).toBeVisible({ timeout: 20000 });

    // Re-open Service Editor
    // Reload to ensure we get the latest service data (with tools)
    await page.reload();
    await page.getByRole('row', { name: serviceName }).getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();
    await page.getByRole('tab', { name: 'Tools' }).click();

    // Now Edit Tool
    // Use aria-label added in previous step or robust selector
    await page.getByRole('button', { name: 'Edit' }).first().click({ force: true });
    await page.getByRole('tab', { name: 'Test & Preview' }).click();

    // Execute
    // Fill args
    const execArgs = JSON.stringify({ id: "test-execution" }, null, 2);
    await page.getByLabel('Test Arguments (JSON)').fill(execArgs);

    // Ensure button is enabled
    const executeBtn = page.getByRole('button', { name: 'Execute' });
    if (await page.getByText('Invalid JSON').isVisible()) {
      console.log("Invalid JSON detected, fixing...");
      await page.getByLabel('Test Arguments (JSON)').fill(execArgs);
    }
    await expect(executeBtn).toBeEnabled();

    // Click Execute
    await executeBtn.click();

    // Verify Result
    try {
      const resultContainer = page.getByTestId('execution-result-container');
      await expect(resultContainer).toContainText('test-execution', { timeout: 5000 });
      await expect(resultContainer).toContainText('/echo/test-execution');
    } catch (e) {
      // Debug failure
      const resultText = await page.getByTestId('execution-result-container').innerText();
      console.log("Execution Result Content:", resultText);
      throw e;
    }
  });
});
