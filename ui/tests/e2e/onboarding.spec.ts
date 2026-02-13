/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Onboarding Flow', () => {
  // Test 1: Verify Onboarding View appears when no services exist
  // We mock the API to return an empty list, ensuring isolation from the real backend state.
  test('should show onboarding view when no services exist', async ({ page }) => {
    // Intercept GET /api/v1/services
    // Use glob pattern which is often more reliable
    await page.route('**/api/v1/services*', async route => {
        if (route.request().method() === 'GET') {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({ services: [] })
            });
        } else {
            await route.continue();
        }
    });

    // Mock dashboard metrics to avoid errors if fallback logic triggers (prevent console noise)
    await page.route('**/api/v1/dashboard/metrics*', async route => {
        await route.fulfill({ status: 200, body: JSON.stringify([]) });
    });
    await page.route('**/api/v1/dashboard/traffic*', async route => {
        await route.fulfill({ status: 200, body: JSON.stringify([]) });
    });
    await page.route('**/api/v1/system/status', async route => {
        await route.fulfill({ status: 200, body: JSON.stringify({ uptime_seconds: 100 }) });
    });
    await page.route('**/api/v1/doctor', async route => {
        await route.fulfill({ status: 200, body: JSON.stringify({ checks: {} }) });
    });

    await page.goto('/');

    // Scope to the onboarding container
    const onboarding = page.getByTestId('onboarding-view');
    await expect(onboarding).toBeVisible({ timeout: 10000 });

    // Check for hero text
    await expect(onboarding.getByRole('heading', { name: 'Welcome to MCP Any' })).toBeVisible();
    await expect(onboarding.getByText('Connect your first server')).toBeVisible();

    // Check for cards (CardTitle renders as div, so use getByText with exact match)
    await expect(onboarding.getByText('Memory Server', { exact: true })).toBeVisible();
    await expect(onboarding.getByText('Marketplace', { exact: true })).toBeVisible();
    await expect(onboarding.getByText('Connect Manually', { exact: true })).toBeVisible();
  });

  // Test 2: Verify Registration Flow
  test('should register memory server and redirect to dashboard', async ({ page }) => {
    let serviceRegistered = false;

    // Mock API
    await page.route('**/api/v1/services*', async route => {
        const method = route.request().method();

        if (method === 'GET') {
            if (!serviceRegistered) {
                // Initial state: Empty
                await route.fulfill({
                    status: 200,
                    contentType: 'application/json',
                    body: JSON.stringify({ services: [] })
                });
            } else {
                // Post-registration state: Populated
                await route.fulfill({
                    status: 200,
                    contentType: 'application/json',
                    body: JSON.stringify({
                        services: [{
                            name: 'memory',
                            id: 'memory-123',
                            status: 'healthy'
                        }]
                    })
                });
            }
        } else if (method === 'POST') {
            // Handle registration
            serviceRegistered = true;
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({ name: 'memory', id: 'memory-123' })
            });
        } else {
            await route.continue();
        }
    });

    // Mock dashboard metrics to avoid errors when switching views
    await page.route('**/api/v1/dashboard/metrics*', async route => {
        await route.fulfill({ status: 200, body: JSON.stringify([]) });
    });
    await page.route('**/api/v1/dashboard/traffic*', async route => {
        await route.fulfill({ status: 200, body: JSON.stringify([]) });
    });
    await page.route('**/api/v1/system/status', async route => {
        await route.fulfill({ status: 200, body: JSON.stringify({ uptime_seconds: 100 }) });
    });
    await page.route('**/api/v1/doctor', async route => {
        await route.fulfill({ status: 200, body: JSON.stringify({ checks: {} }) });
    });

    await page.goto('/');

    // Expect Onboarding View
    const onboarding = page.getByTestId('onboarding-view');
    await expect(onboarding).toBeVisible({ timeout: 10000 });

    // Click install on Memory Server card
    const installBtn = onboarding.getByRole('button', { name: 'One-Click Install' });
    await installBtn.click();

    // Wait for success toast
    await expect(page.getByText('Service Installed')).toBeVisible({ timeout: 15000 });

    // The DashboardShell re-fetches services. Since we flipped `serviceRegistered`, it should now get the list.
    // Wait for Onboarding View to disappear
    await expect(onboarding).not.toBeVisible({ timeout: 10000 });
  });
});
