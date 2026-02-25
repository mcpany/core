/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Supply Chain Attestation', () => {
  const verifiedService = {
    name: 'verified-service',
    id: 'verified-service',
    version: '1.0.0',
    disable: false,
    priority: 0,
    http_service: {
        address: 'https://example.com'
    },
    provenance: {
        verified: true,
        signer_identity: 'Google',
        attestation_time: new Date().toISOString(),
        signature_algorithm: 'SHA256-RSA'
    }
  };

  const unverifiedService = {
    name: 'unverified-service',
    id: 'unverified-service',
    version: '1.0.0',
    disable: false,
    priority: 0,
    http_service: {
        address: 'https://example.com'
    }
  };

  test.beforeAll(async ({ request }) => {
      // Clean up first to ensure clean state
      await request.delete(`/api/v1/services/${verifiedService.name}`).catch(() => {});
      // Small delay to avoid SQLITE_BUSY
      await new Promise(r => setTimeout(r, 500));
      await request.delete(`/api/v1/services/${unverifiedService.name}`).catch(() => {});
      await new Promise(r => setTimeout(r, 500));

      // Seed verified service
      const res1 = await request.post('/api/v1/services', { data: verifiedService });
      expect(res1.ok()).toBeTruthy();
      await new Promise(r => setTimeout(r, 500));

      // Seed unverified service
      const res2 = await request.post('/api/v1/services', { data: unverifiedService });
      expect(res2.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
      await request.delete(`/api/v1/services/${verifiedService.name}`).catch(() => {});
      await request.delete(`/api/v1/services/${unverifiedService.name}`).catch(() => {});
  });

  test('should display verified status for verified service', async ({ page }) => {
    await page.goto(`/upstream-services/${verifiedService.name}`);

    // Check if the tab exists and click it
    const tab = page.getByRole('tab', { name: 'Supply Chain' });
    await expect(tab).toBeVisible();
    await tab.click();

    // Check for "Verified Source" badge
    await expect(page.getByText('Verified Source')).toBeVisible();
    await expect(page.getByText('Trusted')).toBeVisible();

    // Check details
    await expect(page.getByText('Signer Identity')).toBeVisible();
    await expect(page.getByText('Google', { exact: true })).toBeVisible();
    await expect(page.getByText('Valid Signature')).toBeVisible();
  });

  test('should display unverified warning for unverified service', async ({ page }) => {
    await page.goto(`/upstream-services/${unverifiedService.name}`);

    // Click tab
    await page.getByRole('tab', { name: 'Supply Chain' }).click();

    // Check for "Unverified Source"
    await expect(page.getByText('Unverified Source')).toBeVisible();

    // Check for warning content
    await expect(page.getByText('Risk Assessment')).toBeVisible();
    await expect(page.getByText('The author cannot be verified')).toBeVisible();
  });
});
