/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Agent Flow Visualizer E2E', () => {
  // Use serial mode to avoid backend contention if we were seeding,
  // but now we rely on static config so parallel might be fine.
  // Keeping serial for safety in CI.
  test.describe.configure({ mode: 'serial' });

  test.beforeEach(async ({ page }) => {
      // Login as the default admin user configured in docker-compose.test.yml
      // MCPANY_ADMIN_INIT_USERNAME=e2e-admin
      // MCPANY_ADMIN_INIT_PASSWORD=password
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      await page.fill('input[name="username"]', 'e2e-admin');
      await page.fill('input[name="password"]', 'password');
      await page.click('button[type="submit"]', { force: true });
      await page.waitForURL('/', { timeout: 30000 });
  });

  // No cleanup needed as we are not seeding dynamic resources

  test('should visualize live trace of tool execution', async ({ page }) => {
    // 1. Go to Playground
    await page.goto('/playground');

    // 2. Select tool 'get_weather' (from config.minimal.yaml)
    // Open sidebar if closed (it's open by default usually but verify)
    // The sidebar lists tools.
    // Wait for tools list to populate with retry/reload
    await expect(async () => {
        await page.reload();
        // Wait for sidebar to load
        await expect(page.getByRole('heading', { name: 'Console' })).toBeVisible();
        // Use loose matching or look for 'get_weather'
        // The tool name in log is 'weather-service.get_weather', but UI might show simple name or full name.
        await expect(page.getByText('get_weather', { exact: false }).first()).toBeVisible({ timeout: 5000 });
    }).toPass({ timeout: 30000 });

    // Click on get_weather to configure
    await page.getByText('get_weather', { exact: false }).first().click();

    // 3. Dialog opens
    await expect(page.getByRole('dialog')).toBeVisible();
    // Heading might be 'weather-service.get_weather'
    await expect(page.getByRole('heading', { name: /get_weather/ })).toBeVisible();

    // 4. Fill JSON (empty is fine for this tool, or dummy)
    await page.getByRole('tab', { name: 'JSON' }).click();
    await page.locator('textarea').fill('{}');

    // 5. Build Command
    await page.getByRole('button', { name: /build command/i }).click();

    // 6. Send
    await page.getByLabel('Send').click();

    // 7. Wait for result
    // "tool-result" type message.
    await expect(page.getByText('Result:', { exact: false })).toBeVisible({ timeout: 10000 });

    // 8. Navigate to Visualizer
    await page.goto('/visualizer');

    // 9. Verify Graph
    // Should see "MCP Core"
    await expect(page.locator('.react-flow__node').filter({ hasText: 'MCP Core' })).toBeVisible({ timeout: 10000 });

    // Should see "weather-service"
    await expect(async () => {
         const nodes = await page.locator('.react-flow__node').allInnerTexts();
         const hasNode = nodes.some(n => /weather-service/i.test(n));
         expect(hasNode).toBeTruthy();
    }).toPass({ timeout: 10000 });
  });
});
