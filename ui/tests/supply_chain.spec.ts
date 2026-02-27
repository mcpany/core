/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Supply Chain Attestation Viewer', () => {
    const verifiedServiceName = 'verified-service-test';
    const unverifiedServiceName = 'unverified-service-test';

    test.beforeEach(async ({ request }) => {
        // Cleanup potential leftovers
        await request.delete(`/api/v1/services/${verifiedServiceName}`).catch(() => {});
        await request.delete(`/api/v1/services/${unverifiedServiceName}`).catch(() => {});
    });

    test.afterAll(async ({ request }) => {
        // Cleanup
        await request.delete(`/api/v1/services/${verifiedServiceName}`).catch(() => {});
        await request.delete(`/api/v1/services/${unverifiedServiceName}`).catch(() => {});
    });

    test('should display verified status for verified service', async ({ page, request }) => {
        // Seed Verified Service
        const verifiedService = {
            name: verifiedServiceName,
            id: verifiedServiceName,
            version: '1.0.0',
            http_service: { address: 'https://example.com' },
            provenance: {
                verified: true,
                signer_identity: 'MCP Authority',
                attestation_time: new Date().toISOString(),
                signature_algorithm: 'SHA256-RSA'
            }
        };

        const res = await request.post('/api/v1/services', {
            data: verifiedService
        });

        if (!res.ok()) {
            console.error('Failed to create verified service:', await res.text());
        }
        expect(res.ok()).toBeTruthy();

        // Navigate
        await page.goto(`/upstream-services/${verifiedServiceName}`);

        // Wait for page to be ready
        await expect(page.getByRole('heading', { name: verifiedServiceName })).toBeVisible();

        // Click Tab
        await page.getByRole('tab', { name: 'Supply Chain' }).click();

        // Assert
        // Use exact: true to avoid matching the JSON debug dump
        await expect(page.getByRole('heading', { name: 'Verified Service' })).toBeVisible();
        // The signer identity is in a span, also in JSON.
        // We use a locator that targets the value specifically or just strict mode check.
        // Using text with strict mode caused issues, so we use first() or specific container.
        await expect(page.locator('span').filter({ hasText: 'MCP Authority' }).first()).toBeVisible();
        await expect(page.getByText('SHA256-RSA').first()).toBeVisible();
    });

    test('should display unverified warning for unverified service', async ({ page, request }) => {
        // Seed Unverified Service
        const unverifiedService = {
            name: unverifiedServiceName,
            id: unverifiedServiceName,
            version: '1.0.0',
            http_service: { address: 'https://example.com' },
            // No provenance
        };

        const res = await request.post('/api/v1/services', {
            data: unverifiedService
        });

        if (!res.ok()) {
            console.error('Failed to create unverified service:', await res.text());
        }
        expect(res.ok()).toBeTruthy();

        // Navigate
        await page.goto(`/upstream-services/${unverifiedServiceName}`);

        // Wait for page to be ready
        await expect(page.getByRole('heading', { name: unverifiedServiceName })).toBeVisible();

        // Click Tab
        await page.getByRole('tab', { name: 'Supply Chain' }).click();

        // Assert
        await expect(page.getByText('Unverified Service')).toBeVisible();

        // It might render the Alert OR the Card depending on how the backend returns null/empty provenance.
        // We check for either text indicating unverified status.
        const alertText = page.getByText('Proceed with caution');
        const cardText = page.getByText('The signature for this service could not be verified');

        await expect(alertText.or(cardText)).toBeVisible();
    });
});
