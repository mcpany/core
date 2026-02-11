/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Onboarding Flow', () => {
  // Ensure we start with a clean slate for this test
  test.beforeEach(async ({ request }) => {
    // Fetch all services
    const response = await request.get('/api/v1/services');
    if (!response.ok()) return;

    const data = await response.json();
    const services = Array.isArray(data) ? data : (data.services || []);

    // Delete all services to force the onboarding wizard to appear
    for (const s of services) {
        await request.delete(`/api/v1/services/${s.name}`);
    }
  });

  test('should show onboarding wizard when no services exist and deploy demo', async ({ page }) => {
    await page.goto('/');

    // Verify Wizard is visible (Welcome message)
    await expect(page.getByText('Welcome to MCP Any')).toBeVisible();
    await expect(page.getByText('Deploy Demo Service')).toBeVisible();

    // Deploy Demo
    await page.getByRole('button', { name: 'Start with One Click' }).click();

    // Verify Deploying State
    await expect(page.getByText('Deploying...')).toBeVisible();

    // Verify Success
    await expect(page.getByText("You're All Set!")).toBeVisible({ timeout: 30000 }); // Give it time to register

    // Go to Dashboard
    await page.getByRole('button', { name: 'Go to Dashboard' }).click();

    // Verify Wizard is GONE
    await expect(page.getByText('Welcome to MCP Any')).not.toBeVisible();

    // Navigate to Services to confirm
    await page.goto('/upstream-services');

    // Verify weather service exists
    await expect(page.getByText('weather')).toBeVisible();
    // Verify tag (from template if any, wttrin usually doesn't have tags in template but let's check name)

    // Navigate to Tools
    await page.goto('/tools');

    // Verify get_weather tool
    await expect(page.getByText('get_weather')).toBeVisible();
  });
});
