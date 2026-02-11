/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Webhooks Configuration', () => {
  test('should allow configuring system webhooks', async ({ page }) => {
    // Navigate to the Webhooks page
    await page.goto('/webhooks');

    // Verify initial state (assuming default or empty)
    await expect(page.getByText('System Webhooks')).toBeVisible();
    await expect(page.getByText('System Alerts')).toBeVisible();
    await expect(page.getByText('Audit Logging')).toBeVisible();

    // Configure System Alerts
    const alertsToggle = page.getByLabel('Enable Alerts');
    // Ensure it's off initially or handle state. For simplicity assume fresh DB or we check state.
    // If we can't seed, we just test the interaction flow.
    const isAlertsChecked = await alertsToggle.isChecked();
    if (!isAlertsChecked) {
      await alertsToggle.click();
    }

    await page.getByLabel('Webhook URL').first().fill('https://example.com/alerts');

    // Configure Audit Logging
    const auditToggle = page.getByTestId('audit-switch');
    const isAuditChecked = await auditToggle.isChecked();
    if (!isAuditChecked) {
      await auditToggle.click();
    }

    await page.getByLabel('Webhook URL').nth(1).fill('https://example.com/audit');

    // Save
    await page.getByRole('button', { name: 'Save Changes' }).click();

    // Verify Success Toast
    await expect(page.getByText('Settings Saved')).toBeVisible();
    await expect(page.getByText('System webhooks configuration updated.')).toBeVisible();

    // Reload to verify persistence (This verifies backend integration)
    await page.reload();

    // Verify values persisted
    await expect(page.getByLabel('Enable Alerts')).toBeChecked();
    await expect(page.getByLabel('Webhook URL').first()).toHaveValue('https://example.com/alerts');
    await expect(page.getByTestId('audit-switch')).toBeChecked();
    await expect(page.getByLabel('Webhook URL').nth(1)).toHaveValue('https://example.com/audit');
  });
});
