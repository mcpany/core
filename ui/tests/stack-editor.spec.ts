/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Stack Editor', () => {
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
  test('should load the editor and show initial config', async ({ page }) => {
    // Navigate to a stack page (mocking the ID)
    await page.goto('/stacks/default-stack');

    // Check if the editor container is present
    const editor = page.locator('.monaco-editor');
    // await expect(editor).toBeVisible({ timeout: 15000 });

    const visualizer = page.locator('.stack-visualizer-container');
    await expect(visualizer.getByText('weather-service', { exact: true })).toBeVisible();
  });

  // Enabled previously flaky interaction test

  test('should update visualizer when template added', async ({ page }) => {
    // Override the beforeEach mock with a stateful one or just allow the update to reflect
    await page.route('**/api/v1/stacks/*/config', async (route) => {
        if (route.request().method() === 'POST' || route.request().method() === 'PUT') {
             // Accept the update
             const body = route.request().postData();
             await route.fulfill({ status: 200, contentType: 'application/yaml', body: body || '' });
        } else {
             // For GET, we should ideally return the updated config, but for this test checking optimistic UI or initial add,
             // we can just return the base config BUT if the UI refetches immediately, we need to return the new one.
             // Simpler: Return a config that HAS redis for the GETs (simulating successful add)
             // But we don't know WHEN it GETs.
             // If the UI is optimistic, the mock is irrelevant unless it errors.
             // If the UI waits for response, the headers/status matter.
             // Let's assume the UI sends a POST/PUT and updates.
             // If we just continue, it hits the real backend which might not have the stack.
             // Let's try to just fulfill with success for PUT/POST.
             // And for GET?
             // If I use `route.continue()`, it hits backend.
             // Let's use a mock that returns "redis" in the body if we think it's the update.
             // OR: Since we just want to verify the visualizer updates, maybe we just wait for the element?
             // The failure was `toBeVisible` timeout.
             // This suggests the element NEVER appeared.
             // Maybe the click didn't work?
             // Or the visualizer needs the Backend response to render?
             // I'll make the mock return a config WITH redis for ALL requests in this test.
             // This ensures visualizer has it.
             await route.fulfill({
                status: 200,
                contentType: 'application/yaml',
                body: `version: "1.0"
services:
  weather-service:
    image: mcpany/weather-service:latest
  redis:
    image: redis:latest
`
             });
        }
    });

    await page.goto('/stacks/default-stack');
    // We already mocked it to HAVE redis, so it should be visible immediately?
    // But the test says "when template added".
    // If we mock it to already have it, the test proves nothing about "adding".
    // But if the UI loads from the mock, it will show Redis.
    // If we want to test "Adding", we should start with NO redis, then click, then expect Redis.
    // The previous mock had NO redis.
    // The test failed.
    // So likely the visualizer DOES NOT update optimistically, or it re-fetches and gets the old mock.
    // So the fix is to make the mock return Redis AFTER the click (or statefully).
    // Let's try the stateful approach.

    let config = `version: "1.0"
services:
  weather-service:
    image: mcpany/weather-service:latest
`;
    await page.route('**/api/v1/stacks/*/config', async (route) => {
        if (route.request().method() === 'GET') {
            await route.fulfill({ status: 200, contentType: 'application/yaml', body: config });
        } else if (route.request().method() === 'POST' || route.request().method() === 'PUT') {
            // Update config
            const sent = route.request().postData();
            if (sent) config = sent;
            await route.fulfill({ status: 200, contentType: 'application/yaml', body: config });
        }
    });

    await page.goto('/stacks/default-stack');
    await expect(page.locator('.stack-visualizer-container').getByText('weather-service', { exact: true })).toBeVisible({ timeout: 15000 });

    await page.getByText('Redis').click();
    const visualizer = page.locator('.stack-visualizer-container');
    await expect(visualizer.getByText('redis-cache', { exact: true })).toBeVisible();
  });
});
