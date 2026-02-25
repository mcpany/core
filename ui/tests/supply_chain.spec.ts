/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Supply Chain Attestation Viewer', () => {
    const verifiedService = {
        name: 'verified-service',
        id: 'verified-service',
        http_service: { address: 'http://localhost:9999' },
        provenance: {
            verified: true,
            signer_identity: 'Acme Corp CA',
            attestation_time: new Date().toISOString(),
            signature_algorithm: 'SHA256-RSA'
        }
    };

    const unverifiedService = {
        name: 'unverified-service',
        id: 'unverified-service',
        http_service: { address: 'http://localhost:9998' }
    };

    test.beforeAll(async ({ request }) => {
        // Ensure cleanup before start in case of previous failure
        await request.delete(`/api/v1/services/${verifiedService.id}`).catch(() => {});
        await request.delete(`/api/v1/services/${unverifiedService.id}`).catch(() => {});

        // Register services
        const res1 = await request.post('/api/v1/services', { data: verifiedService });
        expect(res1.ok()).toBeTruthy();

        const res2 = await request.post('/api/v1/services', { data: unverifiedService });
        expect(res2.ok()).toBeTruthy();
    });

    test.afterAll(async ({ request }) => {
        // Cleanup
        await request.delete(`/api/v1/services/${verifiedService.id}`);
        await request.delete(`/api/v1/services/${unverifiedService.id}`);
    });

    test('should display verified status and details for verified service', async ({ page }) => {
        await page.goto(`/upstream-services/${verifiedService.id}`);

        await page.getByRole('tab', { name: 'Supply Chain' }).click();

        await expect(page.getByText('Verified Source')).toBeVisible();
        await expect(page.getByText('Acme Corp CA')).toBeVisible();
        await expect(page.getByText('SHA256-RSA')).toBeVisible();
        await expect(page.getByText('Certificate Status')).toBeVisible();
    });

    test('should display unverified warning for unverified service', async ({ page }) => {
        await page.goto(`/upstream-services/${unverifiedService.id}`);

        await page.getByRole('tab', { name: 'Supply Chain' }).click();

        await expect(page.getByText('Unverified Source')).toBeVisible();
        await expect(page.getByText('This service lacks provenance information')).toBeVisible();
    });
});
