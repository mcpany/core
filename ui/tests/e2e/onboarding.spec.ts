/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Onboarding Flow', () => {
  // Cleanup before each test to ensure zero state
  test.beforeEach(async ({ request }) => {
    // List services
    const listRes = await request.get('/api/v1/services');
    if (listRes.ok()) {
        const data = await listRes.json();
        const services = Array.isArray(data) ? data : (data.services || []);
        for (const s of services) {
            await request.delete(`/api/v1/services/${s.name}`);
        }
    }
  });

  test('shows onboarding when no services exist', async ({ page }) => {
    await page.goto('/');
    // Check for hero text
    await expect(page.getByText('Welcome to MCP Any')).toBeVisible({ timeout: 10000 });
    // Check for quick start cards
    await expect(page.getByText('Connect Filesystem')).toBeVisible();
    await expect(page.getByText('Connect GitHub')).toBeVisible();
  });

  test('completes quick start and redirects to dashboard', async ({ page }) => {
    await page.goto('/');

    // Use Manual Configuration to avoid npx timeout/issues in CI environment
    await page.getByText('Manual Configuration').click();

    // Dialog should open
    await expect(page.getByRole('dialog')).toBeVisible();

    // Select Command Line type
    // Using more generic selector as accessibility label might vary in shadcn implementation
    await page.locator('button:has-text("HTTP")').click(); // Default is HTTP
    await page.getByRole('option', { name: 'Command Line' }).click();

    // Fill details
    await page.getByLabel('Service Name').fill('test-service');
    await page.getByLabel('Command').fill('echo "hello"'); // Simple command that always works

    // Click Register
    await page.getByRole('button', { name: 'Register Service' }).click();

    // Wait for Toast
    await expect(page.getByText('Service Registered')).toBeVisible();

    // Should transition to Dashboard automatically (or after reload if needed, but app should be reactive)
    // We wait for "Welcome" to disappear
    await expect(page.getByText('Welcome to MCP Any')).not.toBeVisible({ timeout: 20000 });

    // Dashboard specific element should appear
    await expect(page.getByText('System Health')).toBeVisible();
  });
});
