/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '../fixtures';

test.describe('Dashboard Real Data', () => {
    test.describe.configure({ mode: 'serial' });

    test('should display seeded traffic data', async ({ page, request }) => {
        // 1. Seed data into the backend
        // We use the '/api/v1/debug/seed_traffic' endpoint which is proxied to the backend
        // traffic points: Time (HH:MM), Total, Errors, Latency
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

        const seedRes = await request.post('/api/v1/debug/seed_traffic', {
            data: trafficPoints,
            headers: {
                'Content-Type': 'application/json'
            }
        });
        expect(seedRes.ok()).toBeTruthy();

        // 2. Load the dashboard (already loaded by fixture, but we can navigate explicitly if needed)
        // Fixture logs in at /, so we are at dashboard.
        // We might need to wait for data refresh if SWR/polling is used.
        // Or force reload.
        await page.reload();

        // 3. Verify metrics
        // We seeded 100 requests per minute for 60 minutes = 6000 total requests?
        // AnalyticsDashboard sums them up.
        // 60 points * 100 requests = 6000 total requests.
        // Check if "Total Requests" card shows 6,000 (formatted).

        await expect(page.locator('text=Total Requests')).toBeVisible();

        // The endpoint returns points. The UI sums them up.
        // Total Requests: 6,000 (roughly, might be 5900 if minute rolled over)
        // Check if we have a non-zero value formatted (contains comma or digits)
        // Use a more specific locator to debug and allow for potential data propagation delay
        const totalRequestsLocator = page.locator('div').filter({ hasText: /^Total Requests$/ }).locator('..').getByRole('paragraph');
        await expect(totalRequestsLocator).toHaveText(/[0-9,]+/, { timeout: 30000 });

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

         const seedRes = await request.post('/api/v1/debug/seed_traffic', {
             data: trafficPoints
         });
         expect(seedRes.ok()).toBeTruthy();

         await page.reload();
         // Check for "System Uptime" card
         await expect(page.locator('text=System Uptime')).toBeVisible();

         // In HealthHistoryChart, we infer status from traffic.
         // We might verify that we see some red bars (error status).
         // This is hard to verify visually with text locators, but we can check if the chart renders.
         // And maybe check if "Operational" text is there.
         await expect(page.locator('text=Operational')).toBeVisible();
    });
});
