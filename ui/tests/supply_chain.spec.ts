/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Supply Chain Attestation', () => {
    const verifiedServiceId = 'verified-service-test';
    const unverifiedServiceId = 'unverified-service-test';

    test.beforeAll(async ({ request }) => {
        // Seed Verified Service
        const res1 = await request.post('/api/v1/services', {
            data: {
                id: verifiedServiceId,
                name: verifiedServiceId,
                version: "1.0.0",
                http_service: {
                    address: "http://example.com"
                },
                provenance: {
                    verified: true,
                    signer_identity: "ACME Corp",
                    attestation_time: new Date().toISOString(),
                    signature_algorithm: "SHA256-RSA"
                }
            }
        });
        expect(res1.ok()).toBeTruthy();

        // Seed Unverified Service
        const res2 = await request.post('/api/v1/services', {
            data: {
                id: unverifiedServiceId,
                name: unverifiedServiceId,
                version: "1.0.0",
                http_service: {
                    address: "http://example.com"
                }
            }
        });
        expect(res2.ok()).toBeTruthy();
    });

    test('should display verification details for a verified service', async ({ page }) => {
        await page.goto(`/upstream-services/${verifiedServiceId}`);

        // Click Supply Chain tab
        await page.getByRole('tab', { name: 'Supply Chain' }).click();

        // Check for Verified Service Card
        await expect(page.getByText('Verified Service', { exact: true })).toBeVisible();
        await expect(page.getByText('This service has a valid cryptographic signature and attestation.')).toBeVisible();

        // Check Details
        await expect(page.getByText('ACME Corp')).toBeVisible();
        await expect(page.getByText('SHA256-RSA')).toBeVisible();
        await expect(page.getByText(verifiedServiceId)).toBeVisible(); // Service ID Hash check
    });

    test('should display warning for an unverified service', async ({ page }) => {
        await page.goto(`/upstream-services/${unverifiedServiceId}`);

        // Click Supply Chain tab
        await page.getByRole('tab', { name: 'Supply Chain' }).click();

        // Check for Unverified Service Card
        await expect(page.getByText('Unverified Service')).toBeVisible();
        await expect(page.getByText('This service does not have a valid supply chain attestation.')).toBeVisible();
        await expect(page.getByText('Proceed with caution')).toBeVisible();
    });

    test.afterAll(async ({ request }) => {
        // Cleanup
        await request.delete(`/api/v1/services/${verifiedServiceId}`);
        await request.delete(`/api/v1/services/${unverifiedServiceId}`);
    });
});
