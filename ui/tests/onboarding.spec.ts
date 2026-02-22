/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Onboarding Flow', () => {
  test.beforeEach(async ({ request }) => {
    try { await request.delete('/api/v1/services/wizard-test-service', { headers: { 'X-API-Key': 'test-token' } }); } catch (e) { }
  });

  test.afterEach(async ({ request }) => {
    try { await request.delete('/api/v1/services/wizard-test-service', { headers: { 'X-API-Key': 'test-token' } }); } catch (e) { }
  });

  test('shows setup wizard when no services exist', async ({ page }) => {
    // Mock services to be empty
    await page.route('**/api/v1/services', async route => {
        if (route.request().method() === 'GET') {
            await route.fulfill({ json: [] });
        } else {
            await route.continue();
        }
    });

    await page.goto('/');

    // Wait for the app to load
    await page.waitForLoadState('networkidle');

    const welcome = page.getByText('Welcome to MCP Any');
    await expect(welcome).toBeVisible();
    await expect(page.getByRole('button', { name: /Get Started/i })).toBeVisible();
  });

  test('complete setup wizard flow', async ({ page }) => {
    let mockServices = true;
    // Mock services to be empty initially
    await page.route('**/api/v1/services', async route => {
        if (route.request().method() === 'GET' && mockServices) {
             await route.fulfill({ json: [] });
        } else {
            await route.continue();
        }
    });

    await page.goto('/');

    // Check if we are on wizard
    const welcome = page.getByText('Welcome to MCP Any');
    await expect(welcome).toBeVisible();

    // Step 1: Welcome
    await page.getByRole('button', { name: /Get Started/i }).click();

    // Step 2: Path Selection
    await expect(page.getByText('How do you want to connect?')).toBeVisible();
    await page.getByText('Remote API').click();
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // Step 3: Details
    await expect(page.getByText('Configure Remote Service')).toBeVisible();
    await page.getByLabel('Service Name').fill('wizard-test-service');
    await page.getByLabel('URL').fill('http://localhost:8080'); // Dummy URL

    // Disable mocking so reload sees real services
    mockServices = false;

    await page.getByRole('button', { name: /Connect Service/i }).click();

    // Step 4: Verify Success and Redirect
    await expect(page.getByText('Service Connected', { exact: true })).toBeVisible();

    // Wait for reload (simulated by checking for dashboard)
    // The reload happens after 1000ms.
    // Dashboard should appear because backend has services (defaults + new one)
    await expect(page.getByRole('heading', { name: /Dashboard/i })).toBeVisible({ timeout: 10000 });
  });
});
