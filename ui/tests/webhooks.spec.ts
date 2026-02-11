/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('System Integrations', () => {
  test('should allow configuring alert and audit webhooks', async ({ page }) => {
    await page.goto('/webhooks');

    // Verify Title
    await expect(page.getByRole('heading', { name: 'System Integrations' })).toBeVisible();

    // --- Configure System Alerts ---
    // Locate the card by its title
    // Note: shadcn Card structure puts title in header
    const alertsCard = page.locator('div.rounded-lg.border.bg-card', { hasText: 'System Alerts' }).first();

    // Toggle switch if not enabled
    // We cannot rely on initial state being OFF in a persistent env, so we handle both.
    // Ideally we'd reset state via API before test.
    const alertsSwitch = alertsCard.getByRole('switch');
    const isAlertsEnabled = await alertsSwitch.isChecked();
    if (!isAlertsEnabled) {
        await alertsSwitch.click();
    }

    // Fill URL
    const alertsUrl = `https://hooks.slack.com/test-${Date.now()}`;
    await alertsCard.getByLabel('Webhook URL').fill(alertsUrl);


    // --- Configure Audit Logging ---
    const auditCard = page.locator('div.rounded-lg.border.bg-card', { hasText: 'Audit Logging' }).first();
    const auditSwitch = auditCard.getByRole('switch');
    const isAuditEnabled = await auditSwitch.isChecked();
    if (!isAuditEnabled) {
        await auditSwitch.click();
    }

    // Fill URL
    const auditUrl = `https://audit.example.com/v1/logs-${Date.now()}`;
    await auditCard.getByLabel('Webhook URL').fill(auditUrl);

    // --- Save ---
    await page.getByRole('button', { name: 'Save Changes' }).click();

    // Verify toast confirmation
    await expect(page.getByText('Settings Saved')).toBeVisible();

    // --- Verify Persistence ---
    await page.reload();

    // Wait for data load
    await expect(alertsCard.getByLabel('Webhook URL')).toHaveValue(alertsUrl);
    await expect(auditCard.getByLabel('Webhook URL')).toHaveValue(auditUrl);
  });
});
