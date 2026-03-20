import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from './test-data';

test.describe('User Management', () => {
    test.beforeEach(async ({ request }) => {
        await seedUser(request, 'user-test-bulk-1');
        await seedUser(request, 'user-test-bulk-2');
        await seedUser(request, 'user-test-bulk-3');
    });

    test.afterEach(async ({ request }) => {
        await cleanupUser(request, 'user-test-bulk-1').catch(() => {});
        await cleanupUser(request, 'user-test-bulk-2').catch(() => {});
        await cleanupUser(request, 'user-test-bulk-3').catch(() => {});
    });

    test('supports bulk actions to delete users', async ({ page }) => {
        await page.goto('/users');

        // Check if users exist and wait for table to load
        await expect(page.getByTestId('user-row-user-test-bulk-1')).toBeVisible();

        // Select the first two users
        await page.getByTestId('user-row-user-test-bulk-1').getByRole('checkbox').check();
        await page.getByTestId('user-row-user-test-bulk-2').getByRole('checkbox').check();

        // Verify the sticky header appears and shows correct count
        const bulkHeader = page.getByText('2 selected');
        await expect(bulkHeader).toBeVisible();

        // Handle the confirmation dialog
        page.on('dialog', dialog => dialog.accept());

        // Click delete
        await page.getByRole('button', { name: 'Delete' }).click();

        // Verify users are deleted from the view
        await expect(page.getByTestId('user-row-user-test-bulk-1')).not.toBeVisible();
        await expect(page.getByTestId('user-row-user-test-bulk-2')).not.toBeVisible();

        // The third user should still be visible
        await expect(page.getByTestId('user-row-user-test-bulk-3')).toBeVisible();
    });
});
