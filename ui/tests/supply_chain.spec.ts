/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Supply Chain Attestation Viewer', () => {
  const verifiedServiceName = 'e2e-verified-service';
  const unverifiedServiceName = 'e2e-unverified-service';

  test.beforeEach(async ({ request }) => {
    // Clean up potentially stale data with retries
    const cleanup = async (name: string) => {
        for (let i = 0; i < 3; i++) {
            const res = await request.delete(`/api/v1/services/${name}`);
            if (res.ok() || res.status() === 404) return;
            await new Promise(r => setTimeout(r, 500));
        }
    };

    await cleanup(verifiedServiceName);
    await cleanup(unverifiedServiceName);

    // Seed Verified Service
    const verifiedPayload = {
      name: verifiedServiceName,
      http_service: {
        address: "https://verified.example.com"
      },
      provenance: {
        verified: true,
        signer_identity: "Test Authority",
        attestation_time: "2024-01-01T12:00:00Z",
        signature_algorithm: "ECDSA-SHA256"
      }
    };

    // Seed Unverified Service
    const unverifiedPayload = {
      name: unverifiedServiceName,
      http_service: {
        address: "https://unverified.example.com"
      }
      // No provenance provided
    };

    // Use API to register services with retry to handle SQLITE_BUSY
    const createService = async (payload: any) => {
        for (let i = 0; i < 5; i++) {
            const res = await request.post('/api/v1/services', { data: payload });
            if (res.ok()) return;
            // console.log(`Retry ${i+1} for ${payload.name}: ${res.status()} ${await res.text()}`);
            await new Promise(r => setTimeout(r, 1000));
        }
        throw new Error(`Failed to create service ${payload.name}`);
    };

    await createService(verifiedPayload);
    await createService(unverifiedPayload);
  });

  test.afterEach(async ({ request }) => {
    // Clean up
    await request.delete(`/api/v1/services/${verifiedServiceName}`).catch(() => {});
    await request.delete(`/api/v1/services/${unverifiedServiceName}`).catch(() => {});
  });

  test('should display verified status and details for attested service', async ({ page }) => {
    await page.goto(`/upstream-services/${verifiedServiceName}`);

    // Wait for page to load
    await expect(page.getByRole('heading', { level: 1, name: verifiedServiceName })).toBeVisible();

    // Click Supply Chain Tab
    await page.getByRole('tab', { name: 'Supply Chain' }).click();

    // Verify Badge and Status
    await expect(page.getByText('Verified Service', { exact: true })).toBeVisible();
    await expect(page.getByText('This service has been cryptographically verified')).toBeVisible();

    // Verify Details
    await expect(page.getByText('Test Authority')).toBeVisible();
    await expect(page.getByText('ECDSA-SHA256')).toBeVisible();

    // Verify Timestamp (Formatted)
    // "2024-01-01 12:00:00" depends on local time zone, so might be tricky.
    // We check for year/month at least.
    await expect(page.getByText('2024-01-01')).toBeVisible();
  });

  test('should display unverified warning for unattested service', async ({ page }) => {
    await page.goto(`/upstream-services/${unverifiedServiceName}`);

    // Wait for page to load
    await expect(page.getByRole('heading', { level: 1, name: unverifiedServiceName })).toBeVisible();

    // Click Supply Chain Tab
    await page.getByRole('tab', { name: 'Supply Chain' }).click();

    // Verify Warning
    await expect(page.getByText('Unverified Service')).toBeVisible();
    await expect(page.getByText('This service does not have a valid supply chain attestation')).toBeVisible();

    // Verify Missing Details
    await expect(page.getByText('Unknown Identity')).toBeVisible();
  });
});
