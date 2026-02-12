/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('System Integrations (Webhooks)', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/webhooks');
  });

  test('should configure system alerts (Real Data)', async ({ page, request }) => {
    // Enable Alerts
    const alertsSwitch = page.getByLabel('Enable Notifications');
    await expect(alertsSwitch).toBeVisible();

    // Ensure it starts unchecked (default state)
    if (await alertsSwitch.isChecked()) {
        await alertsSwitch.click();
        await page.getByRole('button', { name: 'Save Changes' }).click();
        await page.waitForTimeout(1000); // Wait for save
    }

    await alertsSwitch.click();

    // Enter URL
    const alertsUrl = page.locator('#alerts-url');
    await expect(alertsSwitch).toBeChecked();
    await expect(alertsUrl).toBeEnabled();
    const testUrl = 'https://api.example.com/alerts-' + Date.now();
    await alertsUrl.fill(testUrl);

    // Save
    const saveButton = page.getByRole('button', { name: 'Save Changes' });
    await saveButton.click();

    // Verify Success Toast
    await expect(page.getByText('Configuration Saved', { exact: true }).first()).toBeVisible();

    // Verify Backend State via API
    // We access the backend directly to confirm persistence
    const response = await request.get('/api/v1/settings');
    expect(response.ok()).toBeTruthy();
    const settings = await response.json();

    expect(settings.alerts.enabled).toBe(true);
    expect(settings.alerts.webhook_url).toBe(testUrl);
  });

  test('should configure audit stream (Real Data)', async ({ page, request }) => {
    // Enable Audit Stream
    const auditSwitch = page.getByLabel('Enable Streaming');
    await expect(auditSwitch).toBeVisible();

    // Reset state if needed
    if (await auditSwitch.isChecked()) {
        await auditSwitch.click();
        await page.getByRole('button', { name: 'Save Changes' }).click();
        await page.waitForTimeout(1000);
    }

    await auditSwitch.click();

    // Enter URL
    const auditUrl = page.locator('#audit-url');
    await expect(auditUrl).toBeEnabled();
    const testUrl = 'https://splunk.example.com/hec-' + Date.now();
    await auditUrl.fill(testUrl);

    // Save
    await page.getByRole('button', { name: 'Save Changes' }).click();

    // Verify Success Toast
    await expect(page.getByText('Configuration Saved', { exact: true }).first()).toBeVisible();

    // Verify Backend State
    const response = await request.get('/api/v1/settings');
    expect(response.ok()).toBeTruthy();
    const settings = await response.json();

    // Verify Audit Config: enabled=true, storage_type=4 (WEBHOOK)
    // Note: storage_type might be returned as enum string or int depending on JSON mapping.
    // proto JSON mapping usually uses strings by default, or ints if emitDefaultValues?
    // Let's check what we get. `settings.audit`
    expect(settings.audit.enabled).toBe(true);
    // storage_type is 4.
    // Backend returns enum string by default in JSON
    expect(settings.audit.storage_type).toBe("STORAGE_TYPE_WEBHOOK");
    expect(settings.audit.webhook_url).toBe(testUrl);
  });
});
