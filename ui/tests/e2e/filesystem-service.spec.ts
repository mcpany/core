/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Filesystem Service Feature', () => {
  test('should proxy api requests', async ({ page }) => {
    const response = await page.goto('/api/v1/services');
    expect(response?.status()).toBe(200);
    const json = await response?.json();
    console.log("Services JSON:", json);
    expect(Array.isArray(json)).toBeTruthy();
  });

  const serviceName = `e2e-fs-${Date.now()}`;

  test.skip('should create a native filesystem service via UI', async ({ page }) => {
    // Navigate to Services Page
    await page.goto('/upstream-services');
    await expect(page.getByRole('heading', { name: 'Upstream Services' })).toBeVisible();

    // Open Add Service Dialog
    const addButton = page.getByRole('button', { name: 'Add Service' });
    await expect(addButton).toBeVisible();

    await addButton.click({ force: true });

    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 10000 });

    // Select "Custom Service" to create a fresh config
    await page.getByText('Custom Service').click();

    // Fill Basic Information
    await page.fill('input[id="name"]', serviceName);

    // Switch to Connection Tab
    await page.getByRole('tab', { name: 'Connection' }).click();

    // Select "Filesystem" from Service Type Dropdown
    await page.getByRole('combobox').click();
    await page.getByRole('option', { name: 'Filesystem' }).click();

    // Verify Filesystem Configuration Options are displayed
    await expect(page.getByText('Filesystem Type')).toBeVisible();
    await expect(page.getByText('Root Paths (Mounts)')).toBeVisible();

    // Configure Root Path
    await page.getByRole('button', { name: 'Add Path' }).click();

    // Fill Virtual Path
    const virtualInput = page.locator('input[placeholder="/virtual/path"]');
    await expect(virtualInput).toBeVisible();
    await virtualInput.fill('/mnt/e2e');

    // Fill Real Path
    const realInput = page.locator('input[placeholder="/real/path"]');
    await expect(realInput).toBeVisible();
    await realInput.fill('/tmp');

    // Save the Service
    await page.getByRole('button', { name: 'Save Changes' }).click();

    // Verify Success Message
    // Wait for dialog to close
    await expect(page.getByRole('dialog')).toBeHidden({ timeout: 10000 });

    // Verify Service appears in the list
    // We reload to ensure we are fetching fresh data from backend
    await page.reload();
    await expect(page.getByText(serviceName)).toBeVisible();

    // Verify Service Type is displayed correctly in the list
    const row = page.locator('tr').filter({ hasText: serviceName });
    await expect(row.getByText('Filesystem')).toBeVisible();

    // Verify Address column shows the mount
    await expect(row.getByText('/tmp')).toBeVisible();
  });
});
