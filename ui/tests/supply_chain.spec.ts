/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Supply Chain Attestation', () => {
  const verifiedServiceId = 'verified-service-test';
  const unverifiedServiceId = 'unverified-service-test';

  test.beforeAll(async ({ request }) => {
    // 1. seed verified service
    await request.post('/api/v1/services', {
      data: {
        name: verifiedServiceId,
        id: verifiedServiceId,
        http_service: { address: 'http://example.com' },
        provenance: {
          verified: true,
          signer_identity: 'ACME Corp Security Team',
          attestation_time: new Date().toISOString(),
          signature_algorithm: 'SHA256-RSA'
        }
      }
    });

    // 2. seed unverified service
    await request.post('/api/v1/services', {
      data: {
        name: unverifiedServiceId,
        id: unverifiedServiceId,
        http_service: { address: 'http://example.org' }
        // No provenance
      }
    });
  });

  test.afterAll(async ({ request }) => {
    // Cleanup
    await request.delete(`/api/v1/services/${verifiedServiceId}`);
    await request.delete(`/api/v1/services/${unverifiedServiceId}`);
  });

  test('should display verified status for verified service', async ({ page }) => {
    await page.goto(`/upstream-services/${verifiedServiceId}`);

    // Check if the Supply Chain tab is visible
    const tabTrigger = page.getByRole('tab', { name: 'Supply Chain' });
    await expect(tabTrigger).toBeVisible();
    await tabTrigger.click();

    // Check content
    await expect(page.getByText('Verified Service')).toBeVisible();
    await expect(page.getByText('ACME Corp Security Team')).toBeVisible();
    await expect(page.getByText('SHA256-RSA')).toBeVisible();
  });

  test('should display warning for unverified service', async ({ page }) => {
    await page.goto(`/upstream-services/${unverifiedServiceId}`);

    // Check if the Supply Chain tab is visible
    const tabTrigger = page.getByRole('tab', { name: 'Supply Chain' });
    await expect(tabTrigger).toBeVisible();
    await tabTrigger.click();

    // Check content
    await expect(page.getByText('Unverified Service')).toBeVisible();
    await expect(page.getByText('Proceed with caution')).toBeVisible();
  });
});
