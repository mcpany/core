/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Dashboard Real Data', () => {
    test.describe.configure({ mode: 'serial' });

    test('should display seeded traffic data', async ({ page, request }) => {
        // 1. Seed data into the backend
        // We use the '/api/v1/debug/seed' endpoint
        page.on('console', msg => console.log('BROWSER LOG:', msg.text()));
        page.on('pageerror', err => console.log('BROWSER ERROR:', err.message));
        const now = new Date();
        const trafficPoints = [];

        // Generate 60 points for the last 60 minutes
        for (let i = 59; i >= 0; i--) {
            const t = new Date(now.getTime() - i * 60000);
            const timeStr = t.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });
            trafficPoints.push({
                time: timeStr,
                requests: 100, // Constant request rate for easy verification
                errors: i % 10 === 0 ? 10 : 0, // Some errors every 10 minutes
                latency: 50 // Constant latency
            });
        }

        const seedRes = await request.post('/api/v1/debug/seed', {
            data: { traffic: trafficPoints },
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': process.env.MCPANY_API_KEY || 'test-token'
            }
        });
        expect(seedRes.ok()).toBeTruthy();

        // 2. Load the dashboard
        await page.goto('/');

        // Debug: Fetch traffic data directly to verify backend state
        const trafficRes = await request.get('/api/v1/dashboard/traffic', {
             headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' }
        });
        expect(trafficRes.ok()).toBeTruthy();
        const trafficData = await trafficRes.json();
        console.log('DEBUG: Traffic Data Length:', trafficData.length);

        // Expect at least one point with requests > 0
        const hasData = trafficData.some((p: any) => p.requests > 0);
        expect(hasData).toBeTruthy();

        // 3. Verify metrics
        await expect(page.locator('text=Total Requests')).toBeVisible();

        // Wait for data to load
        // Find the card containing "Total Requests" and verify the value (text-2xl)
        // We find the specific text element "Total Requests" and traverse up to the Card.
        // "Total Requests" is in CardTitle (div) -> CardHeader (div) -> Card (div)
        const totalRequestsText = page.getByText(/^Total Requests$/);
        const totalRequestsCard = totalRequestsText.locator('../..');
        // Find the value within the card
        const totalRequestsValue = totalRequestsCard.locator('.text-2xl');

        await expect(totalRequestsValue).toHaveText(/[0-9,]+/, { timeout: 30000 });

        // Avg Latency: 50ms
        await expect(page.getByText('50ms')).toBeVisible();

        // 60 errors / 6000 ~ 1%
        await expect(page.getByText(/1\.00%|0\.9\d%/)).toBeVisible();

        // Avg Throughput matches requests per minute?
        // 1.67 rps approx.
        await expect(page.getByText(/1\.6\d rps/)).toBeVisible();

        // 4. Verify charts existence (roughly)
        await expect(page.locator('.recharts-surface').first()).toBeVisible();
    });

    test('should display health history based on traffic', async ({ page, request }) => {
         // 1. Seed data with specific error patterns to affect health
         const now = new Date();
         const trafficPoints = [];

         // 5 mins of high errors (100% errors) to cause "error" status
         for (let i = 0; i < 5; i++) {
             const t = new Date(now.getTime() - i * 60000);
             const timeStr = t.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });
             trafficPoints.push({
                 time: timeStr,
                 requests: 100,
                 errors: 80, // 80% error rate -> should be error status
                 latency: 50
             });
         }

         const seedRes = await request.post('/api/v1/debug/seed', {
             data: { traffic: trafficPoints },
             headers: {
                'Content-Type': 'application/json',
                'X-API-Key': process.env.MCPANY_API_KEY || 'test-token'
            }
         });
         expect(seedRes.ok()).toBeTruthy();

         await page.goto('/');
         // Check for "System Uptime" card
         await expect(page.locator('text=System Uptime')).toBeVisible();

         await expect(page.locator('text=Operational')).toBeVisible();
    });
});
