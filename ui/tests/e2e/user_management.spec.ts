import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser, seedProfiles, cleanupProfiles } from '../e2e/test-data';

test.describe('User Management', () => {
    test.beforeEach(async ({ request, page }) => {
        await cleanupUser(request, "test-api-user").catch(() => { });
        await seedProfiles(request);
        await seedUser(request, "e2e-admin-users");
        page.on('console', msg => console.log(`BROWSER_LOG: ${msg.text()}`));
        await page.goto('/login');
        await page.fill('input[name="username"]', 'e2e-admin-users');
        await page.fill('input[name="password"]', 'password');
        await Promise.all([
            page.waitForURL('/', { timeout: 30000 }),
            page.click('button[type="submit"]', { force: true })
        ]);
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        await cleanupUser(request, "e2e-admin-users").catch(() => { });
        await cleanupUser(request, "test-api-user").catch(() => { });
        await cleanupUser(request, "bulk-user-1").catch(() => { });
        await cleanupUser(request, "bulk-user-2").catch(() => { });
        await cleanupProfiles(request);
    });

    test('should support bulk deleting users', async ({ request, page }) => {
        // Seed database with users to delete
        await seedUser(request, "bulk-user-1");
        await seedUser(request, "bulk-user-2");

        await page.goto('/users');
        await expect(page.locator('h2:has-text("Users")')).toBeVisible();

        // Ensure users are present
        await expect(page.locator('tr').filter({ hasText: 'bulk-user-1' })).toBeVisible();
        await expect(page.locator('tr').filter({ hasText: 'bulk-user-2' })).toBeVisible();

        // Select the two users via their checkboxes
        const row1 = page.locator('tr').filter({ hasText: 'bulk-user-1' });
        await row1.getByRole('checkbox').click();

        const row2 = page.locator('tr').filter({ hasText: 'bulk-user-2' });
        await row2.getByRole('checkbox').click();

        // Verify bulk actions toolbar appears
        const bulkToolbar = page.locator('.animate-in:has-text("2 selected")');
        await expect(bulkToolbar).toBeVisible();

        // Take a screenshot of the selected state
        await page.screenshot({ path: '/home/jules/verification/user_management_bulk_select.png' });

        // Click delete and handle confirmation
        page.on('dialog', async (dialog) => {
            expect(dialog.message()).toContain('delete 2 users');
            await dialog.accept();
        });
        await bulkToolbar.getByRole('button', { name: 'Delete' }).click();

        // Verify toast and deletion
        await expect(page.getByText('2 users have been removed.')).toBeVisible();
        await expect(page.locator('tr').filter({ hasText: 'bulk-user-1' })).toBeHidden();
        await expect(page.locator('tr').filter({ hasText: 'bulk-user-2' })).toBeHidden();
    });

    test('should allow creating a user with API Key', async ({ page }) => {
        await page.goto('/users');

        // Wait for list to load
        await expect(page.locator('h2:has-text("Users")')).toBeVisible();

        // Click Add User
        await page.click('button:has-text("Add User")');

        // Expect Sheet to open
        await expect(page.locator('div[role="dialog"]')).toBeVisible();
        await expect(page.locator('h2:has-text("Add New User")')).toBeVisible();

        // Fill username
        await page.fill('input[name="id"]', 'test-api-user');

        // Select API Key Tab
        await page.click('button[role="tab"]:has-text("API Key")');

        // Wait for and click Generate
        const generateButton = page.getByRole('button', { name: 'Generate' });
        await expect(generateButton).toBeVisible();
        await generateButton.click();

        // Wait for key to be generated
        await expect(page.getByText('Warning: This key will only be shown once')).toBeVisible({ timeout: 10000 });
        const codeBlock = page.locator('pre, div').filter({ hasText: 'mcp_sk_' }).first();
        await expect(codeBlock).toBeVisible({ timeout: 10000 });

        // Save
        const saveButton = page.getByRole('button', { name: 'Save Changes' });
        await expect(saveButton).toBeEnabled();
        await saveButton.click();

        // Verify Sheet closed
        await expect(page.locator('div[role="dialog"]')).toBeHidden({ timeout: 10000 });

        // Verify user created in list
        const row = page.locator('tr').filter({ hasText: 'test-api-user' });
        await expect(row).toBeVisible({ timeout: 15000 });

        // Row should indicate API Key auth
        await expect(row.getByText('API Key')).toBeVisible();

        // Row should have Viewer role (default)
        await expect(row.getByText('viewer', { exact: false })).toBeVisible();
    });
});
