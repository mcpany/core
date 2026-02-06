/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Service Diagnostics', () => {
  const serviceName = 'e2e-diagnostics-test-service';

  test.beforeAll(async ({ request }) => {
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        http_service: {
            address: "http://localhost:50050/health"
        },
        priority: 10
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should run diagnostics including browser connectivity check', async ({ page }) => {
    await page.goto(`/upstream-services/${serviceName}`);
    await page.getByRole('tab', { name: 'Diagnostics' }).click();

    const runButton = page.getByRole('button', { name: 'Run Diagnostics' });
    await expect(runButton).toBeVisible();
    await runButton.click();

    await expect(page.getByText('Configuration Validation')).toBeVisible();
    await expect(page.getByText('Browser Connectivity')).toBeVisible();

    const success = page.getByText('Browser successfully reached the service URL.');
    const failure = page.getByText('Browser failed to reach the service URL.');

    await expect(success.or(failure)).toBeVisible({ timeout: 10000 });

    if (await failure.isVisible()) {
        console.log("Browser Connectivity Failed (Expected in some CI envs). Verifying Heuristic Analysis...");
        await expect(page.getByText('Analysis: Client Network Issue')).toBeVisible();
    } else {
        await expect(success).toBeVisible();
    }

    const valid = page.getByText('Configuration is valid.');
    const invalid = page.getByText('Configuration is invalid.');
    await expect(valid.or(invalid)).toBeVisible();

    // Take screenshot for visual verification
    await page.screenshot({ path: 'diagnostics-screenshot.png', fullPage: true });
  });
});
