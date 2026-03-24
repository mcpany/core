import { test, expect } from '@playwright/test';

// Use same backend url as dev server/app
const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:50050';
const API_KEY = process.env.MCPANY_API_KEY || 'test-token';

test.describe('Dashboard Layout', () => {
    test.beforeEach(async ({ request }) => {
        // Setup: Seed the backend with mock traffic to populate the dashboard metrics
        // and tools. Note: Tool failures and top tools depend on prometheus metrics
        // but traffic points can be seeded to verify the request trend metrics.
        const now = Date.now();
        const trafficData = [];
        for (let i = 0; i < 60; i++) {
            trafficData.push({
                time: now - (60 - i) * 60000,
                total: 10 + i,
                errors: 1,
                latency: 50,
                bytes: 1024
            });
        }

        await request.post(`${BACKEND_URL}/api/v1/debug/seed_traffic`, {
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': API_KEY,
            },
            data: trafficData
        });

        // Let's also register a dummy service so the dashboard isn't empty.
        // If it's empty, we get the Onboarding Hero.
        await request.post(`${BACKEND_URL}/api/v1/services`, {
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': API_KEY,
            },
            data: {
                id: "test-dashboard-svc",
                name: "Test Dashboard Service",
                version: "1.0",
                disable: false,
                command_line_service: {
                    command: "echo test",
                    working_directory: ""
                }
            }
        });
    });

    test('should render new opinionated dashboard layout with structured data', async ({ page }) => {
        await page.goto('/');

        // Verify the main title
        await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();

        // Wait for the DashboardLayout to load and replace the generic grid
        // The layout renders metric cards on top
        await expect(page.getByText('Total Requests')).toBeVisible();
        await expect(page.getByText('Active Services')).toBeVisible();
        await expect(page.getByText('Connected Tools')).toBeVisible();

        // Verify the Service Health Table exists
        await expect(page.getByRole('heading', { name: 'Service Health' })).toBeVisible();

        // Verify our test service appears in the Service Health table
        await expect(page.getByRole('cell', { name: 'Test Dashboard Service' })).toBeVisible();

        // Verify Top Tools Table
        await expect(page.getByRole('heading', { name: 'Top Tools' })).toBeVisible();

        // Verify Recent Failures Table
        await expect(page.getByRole('heading', { name: 'Recent Failures' })).toBeVisible();
    });

    test.afterEach(async ({ request }) => {
        // Cleanup service
        await request.delete(`${BACKEND_URL}/api/v1/services/test-dashboard-svc`, {
            headers: {
                'X-API-Key': API_KEY,
            }
        });
    });
});
