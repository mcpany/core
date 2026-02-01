/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';

const DATE = new Date().toISOString().split('T')[0];
const AUDIT_DIR = path.join(__dirname, `../../.audit/ui/${DATE}`);

test.describe('Service Config Edit', () => {
    test('Updates service config on detail page', async ({ page }) => {
        const serviceId = "svc-test-diff";
        const serviceName = "Test Service";
        const newServiceName = "Test Service Updated";

        // Debug requests
        page.on('request', request => console.log('>>', request.method(), request.url()));
        page.on('response', response => console.log('<<', response.status(), response.url()));
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
        // Use regex to match both /services and /services/id
        await page.route(/\/api\/v1\/services/, async (route) => {
             const method = route.request().method();
             const url = route.request().url();

             console.log(`Intercepted ${method} ${url}`);

             if (url.endsWith('/services') && method === 'GET') {
                 // List services (for siblings)
                 await route.fulfill({ json: { services: [ { id: serviceId, name: serviceName } ] } });
                 return;
             }

             // Match by ID OR Name (since update uses Name in URL)
             // Decode URL to ensure we match "Test Service Updated" regardless of encoding (e.g. %20 vs +)
             const decodedUrl = decodeURIComponent(url);
             if (decodedUrl.includes(serviceId) || decodedUrl.includes(serviceName) || decodedUrl.includes(newServiceName)) {
                 if (method === 'GET') {
                    // Get Service
                    // Must return snake_case properties as the client expects REST response to be snake_case
                    // and maps it to camelCase.
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
                     const postData = route.request().postDataJSON();
                     console.log("Update Body:", JSON.stringify(postData));
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
        await page.getByRole('button', { name: 'Edit Config' }).click({ timeout: 10000 });

        // Wait for Sheet
        await expect(page.getByText('Edit Service')).toBeVisible();

        // Change name
        await page.getByLabel('Service Name').fill(newServiceName);
        await page.getByLabel('Service Name').blur();

        // Verify "Save Changes" button is visible
        await expect(page.getByRole('button', { name: 'Save Changes' })).toBeVisible();

        // Click Save Changes
        await page.getByRole('button', { name: 'Save Changes' }).click();

        // Verify success
        // The toast appears. We check for the generic success or the specific text if toast contains it.
        // The updated ServiceDetail uses: toast({ title: "Service Updated", ... })
        await expect(page.getByText('Service Updated', { exact: true })).toBeVisible();

        // Sheet should close
        await expect(page.getByText('Edit Service')).not.toBeVisible();
    });
});
