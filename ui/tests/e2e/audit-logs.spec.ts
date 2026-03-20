import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Audit Logs Viewer', () => {
    test.beforeAll(async ({ request }) => {
        // Initial setup
        await seedGlobalState(request);
    });

    test('should render tabular data correctly using RichResultViewer instead of raw JSON', async ({ request, page }) => {
        // 1. Seed the database with a rich audit log using the new AuditLogsRaw field
        const auditLogs = [
            {
                timestamp: new Date().toISOString(),
                tool_name: "fetch_rich_data",
                user_id: "admin",
                profile_id: "default",
                trace_id: "trace-12345",
                span_id: "span-12345",
                arguments: { query: "select * from users" },
                // This is the rich data we want to verify is rendered in a table
                result: [
                    { id: 1, name: "Alice", role: "Admin" },
                    { id: 2, name: "Bob", role: "User" }
                ],
                duration: "150ms",
                duration_ms: 150
            }
        ];

        const seedRequest = {
            upstream_services: [],
            service_templates: [],
            users: [],
            credentials: [],
            secrets: [],
            profiles: [],
            audit_logs: auditLogs
        };

        const res = await request.post('/api/v1/debug/seed', {
            data: seedRequest,
            headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token', 'Content-Type': 'application/json' }
        });

        expect(res.ok()).toBeTruthy();

        // 2. Login
        await page.goto('/login');
        await page.fill('input[type="text"]', 'e2e-admin-core');
        await page.fill('input[type="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/dashboard');

        // 3. Navigate to alerts page
        await page.goto('/alerts');

        // 4. Look for the seeded audit log and open its details
        // Wait for the table to load
        await expect(page.locator('text="fetch_rich_data"')).toBeVisible();

        // Click the 'View' button for the first matching row
        const row = page.locator('tr').filter({ hasText: 'fetch_rich_data' }).first();
        await row.getByRole('button', { name: 'View' }).click();

        // 5. Verify that RichResultViewer rendered the table headers correctly (id, name, role)
        const dialog = page.locator('[role="dialog"]');
        await expect(dialog).toBeVisible();

        // Wait for RichResultViewer to process
        await page.waitForTimeout(500);

        // We expect to see table headers 'id', 'name', 'role'
        await expect(dialog.getByRole('columnheader', { name: 'id' })).toBeVisible();
        await expect(dialog.getByRole('columnheader', { name: 'name' })).toBeVisible();
        await expect(dialog.getByRole('columnheader', { name: 'role' })).toBeVisible();

        // Verify table cells
        await expect(dialog.getByRole('cell', { name: 'Alice' })).toBeVisible();
        await expect(dialog.getByRole('cell', { name: 'Admin' })).toBeVisible();
    });
});
