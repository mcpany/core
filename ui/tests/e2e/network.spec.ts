/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Network Topology', () => {
  test.beforeEach(async ({ page, request }) => {
    page.on('console', msg => console.log(`[BROWSER-NETWORK] ${msg.text()}`));
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
        if (!r1.ok() && r1.status() !== 409) {
            console.error(`Failed to seed svc_01: ${r1.status()} ${await r1.text()}`);
        }
        expect(r1.status() === 200 || r1.status() === 201 || r1.status() === 409).toBeTruthy();

        const r2 = await request.post('/api/v1/services', {
            data: {
                id: "svc_02",
                name: "User Service",
                disable: false,
                version: "v1.0",
                grpc_service: { address: "localhost:50051", tools: [], resources: [] }
            }
        });
        if (!r2.ok() && r2.status() !== 409) {
            console.error(`Failed to seed svc_02: ${r2.status()} ${await r2.text()}`);
        }
        expect(r2.status() === 200 || r2.status() === 201 || r2.status() === 409).toBeTruthy();
    } catch (e) {
        console.log("Seeding interaction failed", e);
        throw e;
    }

    await page.goto('/network');
  });

  test('should display network topology nodes', async ({ page }) => {
    // Locate the header specifically to avoid menu link ambiguity
    await expect(page.locator('.text-lg', { hasText: 'Network Graph' })).toBeVisible();

    // Check for nodes
    await expect(page.locator('.react-flow').getByText('MCP Any').first()).toBeVisible(); // Core
    // Wait for services to appear (polling might delay them)
    await expect(page.locator('.react-flow').getByText('Payment Gateway').first().or(page.locator('.react-flow').getByText('User Service').first()).first()).toBeVisible();

    // Verify interaction
    await page.locator('.react-flow').getByText('MCP Any').first().click();
    // Verify sheet opens with correct details
    await expect(page.getByText('Operational Status')).toBeVisible();
    await expect(page.getByText('CORE', { exact: true })).toBeVisible();
  });
});
