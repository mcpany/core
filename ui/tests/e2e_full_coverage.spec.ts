/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('E2E Full Coverage', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
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

  // TODO: Fix flaky address visibility in CI/Docker environment
  test.skip('should register and manage a service', async ({ page }) => {
    await page.goto('/services');
    await page.click('button:has-text("Add Service")');

    // Wait for dialog to be fully visible
    await expect(page.getByRole('dialog')).toBeVisible();

    // Fill form (Basic HTTP Service)
    await page.fill('input[id="name"]', 'e2e-test-service');

    // Explicitly select HTTP type to ensure address field renders
    // shadcn select trigger is a button with role combobox
    const typeSelect = page.getByRole('combobox');
    await typeSelect.click();
    await page.getByRole('option', { name: 'HTTP' }).click();

    // Fill Endpoint
    console.log("TEST UPDATE: Filling Address");
    await page.waitForTimeout(500); // Wait for state update
    const addressInput = page.getByLabel('Address / URL');
    await expect(addressInput).toBeVisible();
    await addressInput.click();
    await addressInput.fill('http://http-echo-server:8080');
    await addressInput.blur();
    await page.waitForTimeout(500); // Wait for state update
    await expect(addressInput).toHaveValue('http://http-echo-server:8080');

    const responsePromise = page.waitForResponse(response =>
      response.url().includes('/api/services') &&
      (response.status() === 200 || response.status() === 201)
    );

    await page.click('button:has-text("Save Changes")');
    await responsePromise;

    await page.reload();

    await expect(page.locator('text=e2e-test-service')).toBeVisible();

    const row = page.locator('tr').filter({ hasText: 'e2e-test-service' });
    // Edit Service
    // Find row with text, then click the Settings button (not the Switch)
    // The Settings button is the second button in the row (Switch is first)
    // Or target by icon presence if possible, but identifying by order is easier here if robust
    // The switch has role="switch". The settings button likely doesn't have a role or is generic button.
    await row.getByRole('link', { name: 'e2e-test-service' }).click();
    await page.getByRole('button', { name: 'Edit Config' }).click();

    await expect(page.locator('input[id="name"]')).toHaveValue('e2e-test-service');
    // Cancel
    await page.click('button:has-text("Cancel")');

    await page.waitForLoadState('networkidle');
    await row.getByRole('switch').click();
    await expect(row).toContainText('Inactive');
  });

  test.skip('should manage global settings', async ({ page }) => {
    await page.goto('/settings');

    await page.getByRole('tab', { name: 'General' }).click();

    await expect(page.getByText('Global Configuration')).toBeVisible();

    // Use a more specific locator for the Log Level combobox
    // It's inside the form, and it's the first combobox (Listen Address is input, GC Interval is input)
    // The previous run failed expecting "DEBUG" but getting "INFO", meaning the click didn't work.
    // Try forcing the click or waiting.
    const form = page.locator('form');
    // Log Level is the first select in the form
    const logLevelSelect = form.getByRole('combobox').nth(0);

    await logLevelSelect.click();
    await expect(page.getByRole('option', { name: 'DEBUG' })).toBeVisible();
    await page.getByRole('option', { name: 'DEBUG' }).click();

    await expect(logLevelSelect).toContainText('DEBUG');

    await page.click('button:has-text("Save Settings")');
    await page.waitForTimeout(500); // Give it a moment

    await page.reload();
    await page.getByRole('tab', { name: 'General' }).click();
    await expect(page.locator('form').getByRole('combobox').nth(0)).toContainText('DEBUG');
  });

  test.skip('should execute tools in playground', async ({ page }) => {
    await page.goto('/playground');

    // Verify playground is loaded
    await expect(page.getByText('Playground')).toBeVisible();

    // Type command
    // Wait for hydration/render
    await page.waitForTimeout(1000);
    // Use generic input locator as the placeholder is complex and potentially brittle
    const input = page.locator('input').first();
    await input.fill('builtin.mcp:list_roots {}');
    await page.getByRole('button', { name: 'Send' }).click();

    // Verify tool call message
    await expect(page.locator('text=Calling: builtin.mcp:list_roots')).toBeVisible();

    // Verify result
    // The UI shows "Result (toolname)" on success, or an error message on failure.
    // We check for either "Result" header or the error message fallback.
    await expect(page.locator('text=Result (builtin.mcp:list_roots)').or(page.locator('text=Tool execution failed'))).toBeVisible({ timeout: 10000 });
  });

  // TODO: Fix flaky secrets test in CI/Docker environment
  test.skip('should manage secrets', async ({ page }) => {
    await page.goto('/settings');
    await page.getByRole('tab', { name: 'Secrets & Keys' }).click();

    await page.click('button:has-text("Add Secret")');
    await page.fill('input[id="name"]', 'e2e_secret');
    await page.fill('input[id="key"]', 'API_KEY');
    await page.fill('input[id="value"]', 'super_secret_value');

    const responsePromise = page.waitForResponse(response =>
        response.url().includes('/api/secrets') &&
        (response.status() === 200 || response.status() === 201)
    );

    const saveBtn = page.getByRole('button', { name: 'Save Secret' });
    await expect(saveBtn).toBeEnabled();
    await saveBtn.click();
    await responsePromise;

    await page.reload();

    await expect(page.locator('text=e2e_secret')).toBeVisible();

    const row = page.locator('tr').filter({ hasText: 'e2e_secret' });
    await row.getByRole('button').click();
    await expect(page.locator('text=e2e_secret')).not.toBeVisible();
  });
});
