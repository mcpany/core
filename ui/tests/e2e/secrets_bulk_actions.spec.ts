import { test, expect, request } from '@playwright/test';
import { seedGlobalState } from './test-data';

const BASE_URL = process.env.BACKEND_URL || 'http://localhost:50050';
const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
const HEADERS = { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };

test.describe('Secrets Manager Bulk Actions', () => {
    test.beforeEach(async ({ page }) => {
        // Create context explicitly for API requests
        const req = await request.newContext({ baseURL: BASE_URL });

        // Seed base state to ensure user exists
        await seedGlobalState(req);

        // Seed 3 test secrets individually
        const secrets = [
            { id: 'secret-bulk-1', name: 'Bulk Secret 1', key: 'KEY_1', value: 'val1', provider: 'custom' },
            { id: 'secret-bulk-2', name: 'Bulk Secret 2', key: 'KEY_2', value: 'val2', provider: 'custom' },
            { id: 'secret-bulk-3', name: 'Bulk Secret 3', key: 'KEY_3', value: 'val3', provider: 'custom' }
        ];

        for (const secret of secrets) {
            const res = await req.post('/api/v1/secrets', { data: secret, headers: HEADERS });
            expect(res.ok()).toBeTruthy();
        }

        // Login
        await page.goto('/login');
        await page.fill('input[name="username"]', 'admin');
        await page.fill('input[name="password"]', 'admin');
        await page.click('button:has-text("Login")');
        await page.waitForURL('/dashboard');
    });

    test('should allow selecting multiple secrets and deleting them', async ({ page }) => {
        await page.goto('/secrets');
        await page.waitForSelector('h1:has-text("Secrets Manager")');

        // Ensure the 3 seeded secrets are visible
        await expect(page.locator('h4:has-text("Bulk Secret 1")')).toBeVisible();
        await expect(page.locator('h4:has-text("Bulk Secret 2")')).toBeVisible();
        await expect(page.locator('h4:has-text("Bulk Secret 3")')).toBeVisible();

        // Check if Select All works
        await page.click('span:has-text("Select All")');

        // Ensure "Delete Selected" button appears
        const deleteSelectedBtn = page.locator('button:has-text("Delete Selected")');
        await expect(deleteSelectedBtn).toBeVisible();

        // Uncheck all
        await page.click('span:has-text("Select All")');
        await expect(deleteSelectedBtn).not.toBeVisible();

        // Select the first two secrets by their specific checkboxes
        // The checkboxes are associated with the secret names
        await page.locator('div.group:has(h4:has-text("Bulk Secret 1")) button[role="checkbox"]').click();
        await page.locator('div.group:has(h4:has-text("Bulk Secret 2")) button[role="checkbox"]').click();

        // Verify the "Delete Selected" button is visible and click it
        await expect(deleteSelectedBtn).toBeVisible();

        // Intercept window.confirm
        page.on('dialog', dialog => dialog.accept());
        await deleteSelectedBtn.click();

        // Wait for removal
        await page.waitForResponse(resp => resp.url().includes('/api/v1/secrets') && resp.request().method() === 'DELETE');
        await page.waitForResponse(resp => resp.url().includes('/api/v1/secrets') && resp.request().method() === 'DELETE');

        // Verify first two secrets are gone
        await expect(page.locator('h4:has-text("Bulk Secret 1")')).not.toBeVisible();
        await expect(page.locator('h4:has-text("Bulk Secret 2")')).not.toBeVisible();

        // Verify the third secret is still there
        await expect(page.locator('h4:has-text("Bulk Secret 3")')).toBeVisible();
    });
});
