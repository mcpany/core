/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Webhooks Integration', () => {
  test('should load settings, update them, and verify persistence', async ({ page }) => {
    // 1. Navigate to Webhooks page
    await page.goto('/webhooks');

    // 2. Check key elements
    await expect(page.getByRole('heading', { name: 'Integrations' })).toBeVisible();
    await expect(page.getByText('Alerts Webhook')).toBeVisible();
    await expect(page.getByText('Audit Log Stream')).toBeVisible();

    // 3. Configure Alert Webhook
    const alertUrlInput = page.locator('#alert-url');
    const testAlertUrl = `https://hooks.slack.com/services/T00000000/B00000000/${Date.now()}`;

    // Enable switch if disabled
    const alertSwitch = page.locator('#alert-enabled');
    const isAlertChecked = await alertSwitch.isChecked();
    if (!isAlertChecked) {
        await alertSwitch.click();
    }

    await alertUrlInput.fill(testAlertUrl);

    // 4. Configure Audit Webhook
    const auditUrlInput = page.locator('#audit-url');
    const testAuditUrl = `https://splunk-collector.example.com/hec/${Date.now()}`;

    // Enable switch if disabled
    const auditSwitch = page.locator('#audit-enabled');
    const isAuditChecked = await auditSwitch.isChecked();
    if (!isAuditChecked) {
        await auditSwitch.click();
    }

    await auditUrlInput.fill(testAuditUrl);

    // 5. Save
    await page.getByRole('button', { name: 'Save Changes' }).click();

    // Expect toast success
    await expect(page.getByText('Settings Saved').first()).toBeVisible();

    // 6. Reload to verify persistence
    await page.reload();

    // Wait for loading to finish
    await expect(page.getByRole('heading', { name: 'Integrations' })).toBeVisible();

    // Verify values match
    await expect(alertUrlInput).toHaveValue(testAlertUrl);
    await expect(auditUrlInput).toHaveValue(testAuditUrl);
    await expect(alertSwitch).toBeChecked();
    await expect(auditSwitch).toBeChecked();
  });
});
