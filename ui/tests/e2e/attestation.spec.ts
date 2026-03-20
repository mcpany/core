/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Supply Chain Attestation Viewer', () => {
  test('should display verified and unverified services correctly', async ({ page }) => {
    // Navigate to the Attestation page
    await page.goto('/attestation');

    // Wait for the page to load
    await expect(page.getByRole('heading', { name: 'Supply Chain Attestation' })).toBeVisible();

    // The backend seeds `mcp-verified-service` (verified) and `weather-service` (unverified)

    // Verify the summary cards
    await expect(page.getByText('Verified Services', { exact: true })).toBeVisible();
    await expect(page.getByText('Unverified Services', { exact: true })).toBeVisible();
    await expect(page.getByText('Total Services', { exact: true })).toBeVisible();

    // Verify mcp-verified-service is present and marked as Verified
    const verifiedRow = page.locator('tr').filter({ hasText: 'mcp-verified-service' });
    await expect(verifiedRow).toBeVisible();
    await expect(verifiedRow.getByText('Verified', { exact: true })).toBeVisible();
    await expect(verifiedRow.locator('text=MCP Community Verified')).toBeVisible();
    await expect(verifiedRow.locator('text=ecdsa-p256-sha256')).toBeVisible();

    // Verify weather-service is present and marked as Unverified
    const unverifiedRow = page.locator('tr').filter({ hasText: 'weather-service' });
    await expect(unverifiedRow).toBeVisible();
    await expect(unverifiedRow.getByText('Unverified', { exact: true }).first()).toBeVisible();
  });
});
