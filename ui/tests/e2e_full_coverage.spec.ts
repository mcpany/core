/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('E2E Full Coverage', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test.beforeEach(async ({ page }) => {
    page.on('console', msg => console.log(`[BROWSER] ${msg.text()}`));
  });

  test('should navigate to all pages and verify content', async ({ page }) => {
    // Dashboard (Home)
    await expect(page).toHaveTitle(/MCPAny/);
    await expect(page.locator('h1')).toContainText('Dashboard');

    // Services
    await page.getByRole('link', { name: 'Services' }).click();
    await expect(page.locator('h2')).toContainText('Services');
    await expect(page.getByRole('button', { name: 'Add Service' })).toBeVisible();

    // Settings
    await page.getByRole('link', { name: 'Settings' }).click();
    await expect(page.locator('h2')).toContainText('Settings');
    // Default tab is profiles
    await expect(page.getByText('Execution Profiles')).toBeVisible();

    // Tools
    await page.goto('/tools');
    await expect(page.locator('h2')).toContainText('Tools');

    // Prompts
    await page.goto('/prompts');
    await expect(page.locator('h2')).toContainText('Prompts');

    // Resources
    await page.goto('/resources');
    await expect(page.locator('h2')).toContainText('Resources');
  });

  test('should execute tools in playground', async ({ page }) => {
    await page.goto('/playground');

    // Verify playground is loaded
    await expect(page.getByRole('heading', { name: 'Playground' })).toBeVisible();

    // Type command
    // Wait for hydration/render
    await page.waitForTimeout(1000);
    const input = page.getByRole('textbox');
    // We use weather-service.get_weather because list_roots requires a session (unavailable in HTTP playground)
    await input.fill('weather-service.get_weather {}');
    await page.getByRole('button', { name: 'Send' }).click();

    // Verify tool call message
    await expect(page.locator('text=Calling: weather-service.get_weather')).toBeVisible();

    // Verify result contains the expected weather output
    await expect(page.getByText('sunny')).toBeVisible();

    // Verify result
    // The UI shows "Result (toolname)" on success, or an error message on failure.
    // We check for either "Result" header or the error message fallback.
    await expect(page.locator('text=Result (weather-service.get_weather)').or(page.locator('text=Tool execution failed'))).toBeVisible({ timeout: 30000 });
  });

});
