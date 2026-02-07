import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from './e2e/test-data';

const serviceName = 'e2e-http-service';

test.describe('HTTP Tool Editor', () => {
    test.beforeEach(async ({ request, page }) => {
        // Seed User
        await seedUser(request, "e2e-admin");

        // Seed Service
        // Delete first to be clean
        await request.delete(`/api/v1/services/${serviceName}`, { headers: { 'X-API-Key': 'test-token' } }).catch(() => {});
        await request.post('/api/v1/services', {
            headers: { 'X-API-Key': 'test-token' },
            data: {
                name: serviceName,
                version: "1.0.0",
                http_service: {
                    address: "https://httpbin.org"
                }
            }
        });

        // Login
        await page.goto('/login');
        await page.fill('input[name="username"]', 'e2e-admin');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        await request.delete(`/api/v1/services/${serviceName}`, { headers: { 'X-API-Key': 'test-token' } });
        await cleanupUser(request, "e2e-admin");
    });

    test('should allow adding and editing tools for HTTP service', async ({ page, request }) => {
        // Navigate to services
        await page.goto('/upstream-services');

        // Find the service row and click Edit
        // We use filter to find the row with the service name
        const row = page.locator('tr').filter({ hasText: serviceName });

        // Click the dropdown menu button
        await row.getByRole('button', { name: 'Open menu' }).click();

        // Click Edit
        await page.getByRole('menuitem', { name: 'Edit' }).click();

        // Switch to Tools tab
        await page.getByRole('tab', { name: 'Tools' }).click();

        // Check empty state
        await expect(page.getByText('No tools defined')).toBeVisible();

        // Add Tool
        await page.getByRole('button', { name: 'Add Tool' }).click();

        // Fill Tool Form
        await page.fill('input#tool-name', 'get_ip');
        await page.fill('input#tool-desc', 'Get IP address');

        // Fill HTTP Call Details
        await page.fill('input#endpoint-path', '/ip');

        // Save Tool
        await page.getByRole('button', { name: 'Save Tool' }).click();

        // Verify in list
        await expect(page.locator('table').getByText('get_ip')).toBeVisible();
        await expect(page.locator('table').getByText('GET', { exact: true })).toBeVisible();
        await expect(page.locator('table').getByText('/ip')).toBeVisible();

        // Save Service
        await page.getByRole('button', { name: 'Save Changes' }).click();
        await expect(page.getByText('Service Updated')).toBeVisible();

        // Verify via API
        const res = await request.get(`/api/v1/services/${serviceName}`, { headers: { 'X-API-Key': 'test-token' } });
        const data = await res.json();

        // Handle response format (might be wrapped or not, assume backend standard)
        const service = data.service || data;
        const tools = service.http_service?.tools || [];
        const calls = service.http_service?.calls || {};

        expect(tools).toHaveLength(1);
        expect(tools[0].name).toBe('get_ip');

        const callId = tools[0].call_id;
        expect(calls[callId]).toBeDefined();
        expect(calls[callId].endpoint_path).toBe('/ip');
    });
});
