/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Stack Status', () => {
    const stackName = 'test-status-stack';
    const service1 = {
        name: 'test-svc-1',
        http_service: { address: 'http://localhost:9991' },
        disable: false
    };
    const service2 = {
        name: 'test-svc-2',
        http_service: { address: 'http://localhost:9992' },
        disable: true
    };

    test.beforeEach(async ({ request }) => {
        const collection = {
            name: stackName,
            services: [service1, service2]
        };

        const res = await request.put(`/api/v1/collections/${stackName}`, {
            data: collection,
            headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' }
        });
        expect(res.ok()).toBeTruthy();
    });

    test.afterEach(async ({ request }) => {
        await request.delete(`/api/v1/collections/${stackName}`, {
            headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' }
        });
        // Also cleanup services if they persist
        await request.delete(`/api/v1/services/${service1.name}`, { headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' } }).catch(() => {});
        await request.delete(`/api/v1/services/${service2.name}`, { headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' } }).catch(() => {});
    });

    test('should display stack status and allow bulk actions', async ({ page }) => {
        await page.goto(`/stacks/${stackName}`);

        // Click on "Overview & Status" tab (it might be default, but let's be safe)
        await page.getByRole('tab', { name: 'Overview & Status' }).click();

        // Check if services are listed
        await expect(page.getByText(service1.name)).toBeVisible();
        await expect(page.getByText(service2.name)).toBeVisible();

        // Verify "Start All" and "Stop All" buttons are present
        await expect(page.getByRole('button', { name: 'Start All' })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Stop All' })).toBeVisible();

        // Test interaction
        let startCalls = 0;
        page.on('request', req => {
            if (req.method() === 'PUT' && req.url().includes('/api/v1/services/') && !req.url().includes('collections')) {
                 const postData = req.postDataJSON();
                 if (postData && postData.disable === false) {
                     startCalls++;
                 }
            }
        });

        await page.getByRole('button', { name: 'Start All' }).click();

        // Wait a bit for requests
        await page.waitForTimeout(1000);

        // We expect requests to enable services
        expect(startCalls).toBeGreaterThan(0);
    });
});
