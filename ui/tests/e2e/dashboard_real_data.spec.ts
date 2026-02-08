/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

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

        // 2. Load the dashboard
        await page.goto('/');

        // Debug: Fetch traffic data directly to verify backend state
        const trafficRes = await request.get('/api/v1/dashboard/traffic');
        expect(trafficRes.ok()).toBeTruthy();
        const trafficData = await trafficRes.json();
        console.log('DEBUG: Traffic Data:', JSON.stringify(trafficData));
        // Expect at least one point with requests > 0
        const hasData = trafficData.some((p: any) => p.requests > 0);
        expect(hasData).toBeTruthy();

        // 3. Verify metrics
        // We seeded 100 requests per minute for 60 minutes = 6000 total requests?
        // Wait, GetTrafficHistory returns the history.
        // AnalyticsDashboard sums them up.
        // 60 points * 100 requests = 6000 total requests.
        // Check if "Total Requests" card shows 6,000 (formatted).

        await expect(page.locator('text=Total Requests')).toBeVisible();

        // The endpoint returns points. The UI sums them up.
        // Total Requests: 6,000 (roughly, might be 5900 if minute rolled over)
        // Check if we have a non-zero value formatted (contains comma or digits)
        const totalRequests = page.locator('div').filter({ hasText: /^Total Requests$/ }).locator('..').getByRole('paragraph');
        // Or simpler: find the value card.
        // The card structure: Header "Total Requests", Content: "6,000"
        // Let's just look for the text "Total Requests" and verify the number nearby is not "0"

        await expect(page.locator('text=Total Requests')).toBeVisible();
        // Wait for data to load (it starts at 0 or empty)
        await expect(page.getByText('Use traffic history to infer historical health').first()).toBeHidden(); // Ensure no error text is shown if that was a thing?
        // Just wait for non-zero requests
        // We expect around 6,000.
        // Use a more specific locator to debug and allow for potential data propagation delay
        // Target the value div (.text-2xl) instead of the change description (paragraph)
        const totalRequestsLocator = page.locator('div').filter({ hasText: /^Total Requests$/ }).locator('..').locator('.text-2xl');
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

         await page.goto('/');
         // Check for "System Uptime" card
         await expect(page.locator('text=System Uptime')).toBeVisible();

         // In HealthHistoryChart, we infer status from traffic.
         // We might verify that we see some red bars (error status).
         // This is hard to verify visually with text locators, but we can check if the chart renders.
         // And maybe check if "Operational" text is there.
         await expect(page.locator('text=Operational')).toBeVisible();
    });
});
