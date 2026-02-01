/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';

test.describe('Service Config Edit', () => {
    test('Can edit service config via Sheet', async ({ page }) => {
        const serviceId = "svc-test-edit";
        const serviceName = "Test Service";
        const newServiceName = "Test Service Updated";

        // Debug requests
        page.on('console', msg => console.log('PAGE LOG:', msg.text()));

        // Abort gRPC to force REST fallback
        await page.route('**/mcpany.api.v1.RegistrationService/GetService', async (route) => {
             await route.abort();
        });

        // Mock credentials
        await page.route(/\/api\/v1\/credentials/, async (route) => {
             await route.fulfill({ json: { credentials: [] } });
        });

        // Mock all services API
        await page.route(/\/api\/v1\/services/, async (route) => {
             const method = route.request().method();
             const url = route.request().url();

             if (url.endsWith('/services') && method === 'GET') {
                 // List services
                 await route.fulfill({ json: { services: [ { id: serviceId, name: serviceName } ] } });
                 return;
             }

             const decodedUrl = decodeURIComponent(url);
             if (decodedUrl.includes(serviceId) || decodedUrl.includes(serviceName) || decodedUrl.includes(newServiceName)) {
                 if (method === 'GET') {
                    await route.fulfill({
                        json: {
                            service: {
                                id: serviceId,
                                name: serviceName,
                                disable: false,
                                version: "1.0.0",
                                http_service: { address: "http://example.com", tools: [], resources: [] }
                            }
                        }
                    });
                 } else if (method === 'PUT' || method === 'POST') {
                     // Update Service
                     await route.fulfill({ json: { success: true, name: newServiceName } });
                 } else {
                    await route.continue();
                 }
                 return;
             }

             await route.continue();
        });

        // Go to service detail page
        await page.goto(`/service/${serviceId}`);

        // Click Edit Config
        await page.getByRole('button', { name: 'Edit Config' }).click();

        // Wait for Sheet
        await expect(page.getByRole('dialog')).toBeVisible();
        await expect(page.getByText('Edit Service Configuration')).toBeVisible();

        // Change name (General tab is default)
        await page.getByLabel('Service Name').fill(newServiceName);

        // Click Save Changes
        await page.getByRole('button', { name: 'Save Changes' }).click();

        // Verify success toast
        // Note: The toast message depends on what ServiceDetail calls in toast()
        // In my code: toast({ title: "Service Updated", description: "Configuration saved successfully." });
        await expect(page.getByText('Service Updated').first()).toBeVisible();
        await expect(page.getByText('Configuration saved successfully').first()).toBeVisible();
    });
});
