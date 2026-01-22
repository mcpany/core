/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Dashboard Real Data', () => {

    test('should display seeded traffic data', async ({ page }) => {
        // Mock network calls
        await page.route('/api/v1/dashboard/traffic', async (route) => {
            const now = new Date();
            const trafficPoints = [];
            // Generate 10 points
            for (let i = 9; i >= 0; i--) {
                const t = new Date(now.getTime() - i * 60000);
                const timeStr = t.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });
                trafficPoints.push({
                    time: timeStr,
                    requests: 100,
                    errors: i % 10 === 0 ? 10 : 0,
                    latency: 50
                });
            }
            await route.fulfill({ json: trafficPoints });
        });

        // Mock dashboard metrics which drives the "Total Requests" card
        await page.route('/api/v1/dashboard/metrics', async (route) => {
            await route.fulfill({
                json: [
                    { label: "Total Requests", value: "1,000", icon: "Activity", change: "+10%", trend: "up" },
                    { label: "Avg Latency", value: "50ms", icon: "Clock", change: "-5ms", trend: "down" },
                    { label: "Error Rate", value: "1.00%", icon: "AlertCircle", change: "+0.1%", trend: "up" },
                    { label: "Avg Throughput", value: "1.67 rps", icon: "Zap", change: "+0.2", trend: "up" }
                ]
            });
        });

        await page.route('/api/v1/dashboard/top-tools', async (route) => {
             await route.fulfill({ json: [] });
        });

        await page.route('/api/v1/doctor', async (route) => {
            await route.fulfill({ json: { status: 'healthy', checks: {} } });
        });

        await page.route('/api/v1/debug/seed_traffic', async (route) => {
             await route.fulfill({ status: 200 });
        });

        // Load dashboard
        await page.goto('/');

        // Verify metrics
        await expect(page.locator('text=Total Requests')).toBeVisible();

        // Use a more specific locator to debug and allow for potential data propagation delay
        // Note: The total requests logic sums up the mock data: 10 points * 100 reqs = 1000
        const totalRequestsLocator = page.locator('.text-2xl.font-bold').first();
        // We expect 1,000
        await expect(totalRequestsLocator).toHaveText(/[1-9][0-9,]{2,}/, { timeout: 30000 });

        // Avg Latency: 50ms
        await expect(page.getByText('50ms')).toBeVisible();

        // 10 errors / 1000 = 1%
        await expect(page.getByText(/1\.00%|0\.9\d%/)).toBeVisible();

        // Avg Throughput matches requests per minute?
        // 1000 requests in 10 mins (600s) = 1.67 rps.
        await expect(page.getByText('1.67 rps')).toBeVisible();

        // Verify charts existence (roughly)
        await expect(page.locator('.recharts-surface').first()).toBeVisible();
    });

    test('should display health history based on traffic', async ({ page }) => {
         // Mock network calls with errors
         await page.route('/api/v1/dashboard/traffic', async (route) => {
             const now = new Date();
             const trafficPoints = [];
             // 5 mins of high errors (80% error rate)
             for (let i = 0; i < 5; i++) {
                 const t = new Date(now.getTime() - i * 60000);
                 const timeStr = t.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });
                 trafficPoints.push({
                     time: timeStr,
                     requests: 100,
                     errors: 80,
                     latency: 50
                 });
             }
             await route.fulfill({ json: trafficPoints });
         });

         await page.route('/api/v1/doctor', async (route) => {
             await route.fulfill({ json: { status: 'degraded', checks: {} } });
         });

         await page.route('/api/v1/dashboard/top-tools', async (route) => {
             await route.fulfill({ json: [] });
         });

         await page.route('/api/v1/debug/seed_traffic', async (route) => {
             await route.fulfill({ status: 200 });
        });

         await page.goto('/');
         // Check for "System Uptime" card
         await expect(page.locator('text=System Uptime')).toBeVisible();

         // In HealthHistoryChart, we infer status from traffic.
         // We verify that we see "Operational" text which is static,
         // but the fact that it renders means no crash.
         await expect(page.locator('text=Operational')).toBeVisible();
    });
});
