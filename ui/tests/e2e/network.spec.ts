/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Network Topology', () => {
  test.beforeEach(async ({ page, request }) => {
    // Seed data
     try {
        const r1 = await request.post('/api/v1/services', {
            data: {
                id: "svc_01",
                name: "Payment Gateway",
                connection_pool: { max_connections: 100 },
                disable: false,
                version: "v1.2.0",
                http_service: { address: "https://stripe.com", tools: [], resources: [] }
            }
        });
        expect(r1.ok()).toBeTruthy();

        const r2 = await request.post('/api/v1/services', {
            data: {
                id: "svc_02",
                name: "User Service",
                disable: false,
                version: "v1.0",
                grpc_service: { address: "localhost:50051", tools: [], resources: [] }
            }
        });
        expect(r2.ok()).toBeTruthy();
    } catch (e) {
        console.log("Seeding failed or services already exist", e);
        throw e;
    }

    await page.goto('/network');
  });

  test.skip('should display network topology nodes', async ({ page }) => {
    // Locate the header specifically to avoid menu link ambiguity
    await expect(page.locator('.text-lg', { hasText: 'Network Graph' })).toBeVisible();

    // Check for nodes
    await expect(page.getByText('MCP Any')).toBeVisible(); // Core
    // Wait for services to appear (polling might delay them)
    await expect(page.getByText('Payment Gateway').or(page.getByText('User Service'))).toBeVisible();

    // Verify interaction
    await page.getByText('MCP Any').click();
    // Verify sheet opens with correct details
    await expect(page.getByText('Operational Status')).toBeVisible();
    await expect(page.getByText('CORE', { exact: true })).toBeVisible();
  });
});
