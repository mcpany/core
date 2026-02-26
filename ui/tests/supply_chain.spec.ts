/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Supply Chain Attestation Viewer', () => {
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
      signer_identity: 'Acme Corp Official',
      attestation_time: '2024-01-01T12:00:00Z',
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
    // Seed services
    // Ignore errors if they already exist (or delete first)
    await request.delete(`/api/v1/services/${verifiedService.name}`).catch(() => {});
    await request.delete(`/api/v1/services/${unverifiedService.name}`).catch(() => {});

    await request.post('/api/v1/services', { data: verifiedService });
    await request.post('/api/v1/services', { data: unverifiedService });
  });

  test.afterAll(async ({ request }) => {
    // Cleanup
    await request.delete(`/api/v1/services/${verifiedService.name}`).catch(() => {});
    await request.delete(`/api/v1/services/${unverifiedService.name}`).catch(() => {});
  });

  test('should display verified badge and details for verified service', async ({ page }) => {
    await page.goto(`/upstream-services/${verifiedService.name}`);

    // Click Supply Chain tab
    await page.getByRole('tab', { name: 'Supply Chain' }).click();

    // Assert "Verified Service" card
    await expect(page.getByText('Verified Service', { exact: true })).toBeVisible();
    await expect(page.getByText('Trusted')).toBeVisible();

    // Assert Details
    await expect(page.getByText('Acme Corp Official')).toBeVisible();
    await expect(page.getByText('SHA256-RSA')).toBeVisible();
  });

  test('should display warning for unverified service', async ({ page }) => {
    await page.goto(`/upstream-services/${unverifiedService.name}`);

    // Click Supply Chain tab
    await page.getByRole('tab', { name: 'Supply Chain' }).click();

    // Assert "Unverified Service" warning
    await expect(page.getByText('Unverified Service')).toBeVisible();
    await expect(page.getByText('Use with caution')).toBeVisible();
  });
});
