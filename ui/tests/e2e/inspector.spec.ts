import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Inspector Table', () => {
    test.beforeAll(async ({ request }) => {
        // "Real Data Law": Mandates backend database seeding
        await seedGlobalState(request);
    });

    test('should display table headers even when there are no traces', async ({ page }) => {
        await page.goto('/inspector');

        // Ensure the headers are visible
        await expect(page.getByRole('columnheader', { name: 'Timestamp' })).toBeVisible();
        await expect(page.getByRole('columnheader', { name: 'Type' })).toBeVisible();
        await expect(page.getByRole('columnheader', { name: 'Method / Name' })).toBeVisible();
        await expect(page.getByRole('columnheader', { name: 'Status' })).toBeVisible();
        await expect(page.getByRole('columnheader', { name: 'Duration' })).toBeVisible();

        // Wait for the empty state message inside the table
        await expect(page.getByText('No traces found.')).toBeVisible();
    });
});
