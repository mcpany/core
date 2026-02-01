/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Service Onboarding', () => {
  // Use a unique ID suffix or clear all services to ensure clean state
  test.beforeEach(async ({ request }) => {
    // List all services
    const listRes = await request.get('/api/v1/services');
    if (listRes.ok()) {
      const data = await listRes.json();
      const services = Array.isArray(data) ? data : (data.services || []);

      // Delete each service to ensure empty state
      for (const service of services) {
        await request.delete(`/api/v1/services/${service.name}`);
      }
    }
  });

  test('should display onboarding screen when no services are registered', async ({ page }) => {
    await page.goto('/upstream-services');

    // Verify Onboarding Hero
    await expect(page.getByText('Connect your first service')).toBeVisible();
    await expect(page.getByText('MCP Any acts as a universal adapter')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Connect Manually' })).toBeVisible();

    // Verify Quick Start Cards are present
    // We check for some of the featured templates
    await expect(page.getByText('GitHub', { exact: true })).toBeVisible();
    await expect(page.getByText('PostgreSQL', { exact: true })).toBeVisible();
    await expect(page.getByText('Google Maps', { exact: true })).toBeVisible();

    // Test Interaction: Click GitHub Card
    // Find the card by text and click it
    const githubCard = page.locator('.group', { hasText: 'GitHub' }).first();
    await githubCard.click();

    // Verify Sheet Opens with GitHub Configuration
    await expect(page.getByRole('dialog')).toBeVisible();
    // The sheet title should update based on selection
    await expect(page.getByRole('heading', { name: 'Configure GitHub' })).toBeVisible();

    // Verify the specific field for GitHub template is visible
    await expect(page.getByLabel('GitHub Personal Access Token')).toBeVisible();

    // Close the sheet
    await page.getByRole('button', { name: 'Close' }).click();

    // Verify we are back to onboarding
    await expect(page.getByText('Connect your first service')).toBeVisible();
  });

  test('should allow manual connection from onboarding', async ({ page }) => {
    await page.goto('/upstream-services');

    // Click "Connect Manually"
    await page.getByRole('button', { name: 'Connect Manually' }).click();

    // Verify Sheet Opens with "New Service" (Custom Service template logic)
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'New Service' })).toBeVisible();

    // Verify standard service editor fields are present
    await expect(page.getByLabel('Service Name')).toBeVisible();
    await expect(page.getByLabel('Service Type')).toBeVisible();
  });
});
