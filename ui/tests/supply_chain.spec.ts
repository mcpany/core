/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Supply Chain Attestation', () => {
  const verifiedServiceName = `verified-service-${Date.now()}`;
  const unverifiedServiceName = `unverified-service-${Date.now()}`;

  test.beforeAll(async ({ request }) => {
    // 1. Seed Verified Service
    const verifiedPayload = {
      name: verifiedServiceName,
      id: verifiedServiceName,
      version: '1.0.0',
      disable: false,
      priority: 0,
      load_balancing_strategy: 0,
      http_service: {
        address: 'https://example.com'
      },
      provenance: {
        verified: true,
        signer_identity: 'ACME Corp Security Team',
        attestation_time: new Date().toISOString(),
        signature_algorithm: 'SHA256withRSA'
      }
    };

    const res1 = await request.post('/api/v1/services', {
      data: verifiedPayload,
      headers: { 'Content-Type': 'application/json' }
    });
    if (!res1.ok()) {
        console.error("Failed to register verified service:", res1.status(), await res1.text());
    }
    expect(res1.ok()).toBeTruthy();

    // 2. Seed Unverified Service
    const unverifiedPayload = {
      name: unverifiedServiceName,
      id: unverifiedServiceName,
      version: '1.0.0',
      disable: false,
      priority: 0,
      load_balancing_strategy: 0,
      http_service: {
        address: 'https://example.org'
      }
      // No provenance
    };

    const res2 = await request.post('/api/v1/services', {
      data: unverifiedPayload,
      headers: { 'Content-Type': 'application/json' }
    });
    if (!res2.ok()) {
        console.error("Failed to register unverified service:", res2.status(), await res2.text());
    }
    expect(res2.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    // Cleanup
    await request.delete(`/api/v1/services/${verifiedServiceName}`);
    await request.delete(`/api/v1/services/${unverifiedServiceName}`);
  });

  test('should display verified badge and details for verified service', async ({ page }) => {
    await page.goto(`/upstream-services/${verifiedServiceName}`);

    // Wait for load
    await expect(page.getByRole('heading', { level: 1, name: verifiedServiceName })).toBeVisible();

    // Click Supply Chain tab
    await page.getByRole('tab', { name: 'Supply Chain' }).click();

    // Assert Verified State
    await expect(page.getByText('Verified Service')).toBeVisible();
    await expect(page.getByText('ACME Corp Security Team').first()).toBeVisible();
    await expect(page.getByText('SHA256withRSA')).toBeVisible();
    await expect(page.getByText('Signature Valid')).toBeVisible();
  });

  test('should display warning for unverified service', async ({ page }) => {
    await page.goto(`/upstream-services/${unverifiedServiceName}`);

    // Wait for load
    await expect(page.getByRole('heading', { level: 1, name: unverifiedServiceName })).toBeVisible();

    // Click Supply Chain tab
    await page.getByRole('tab', { name: 'Supply Chain' }).click();

    // Assert Unverified State
    await expect(page.getByText('Unverified Service')).toBeVisible();
    await expect(page.getByText('could not be verified')).toBeVisible();
  });
});
