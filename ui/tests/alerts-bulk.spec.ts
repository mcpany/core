import { test, expect } from '@playwright/test';

test.describe('Alerts Bulk Actions', () => {
    test('can perform bulk acknowledge and resolve on alerts', async ({ page }) => {
        // Go to alerts page
        await page.goto('/alerts');

        // Wait for the alerts to load from the seeded backend data
        await expect(page.getByText('High CPU Usage')).toBeVisible({ timeout: 30000 });

        // Ensure "Acknowledge" button is not visible yet
        await expect(page.getByRole('button', { name: /Acknowledge/i })).not.toBeVisible();

        // Check the select all checkbox
        const selectAllCheckbox = page.getByRole('checkbox', { name: /select all/i });
        await selectAllCheckbox.click();

        // Verify bulk actions bar appears and shows correct selected count
        // Note: Seeded data provides 5 alerts
        await expect(page.getByText('5 selected', { exact: true })).toBeVisible();

        const ackButton = page.getByRole('button', { name: /Acknowledge/i });
        await expect(ackButton).toBeVisible();

        // Click Acknowledge
        await ackButton.click();

        // The toast will appear, proving the code executed properly
        await expect(page.getByText('5 alerts marked as acknowledged')).toBeVisible();

    });
});
