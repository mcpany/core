import { test, expect } from '@playwright/test';

test.describe('Trace Inspector', () => {
    test('supports bulk deletion of traces and verifies backend state change', async ({ page }) => {
        await page.goto('/inspector');

        // Initial check and seed traces
        await page.waitForSelector('text=Inspector');
        await page.getByRole('button', { name: 'Seed Trace' }).click();

        // Wait for traces to appear in the table. Seed creates 5 traces
        const firstTrace = page.getByText('trace-seed-').first();
        await expect(firstTrace).toBeVisible({ timeout: 30000 });

        // Count traces before deletion
        const traceCountBefore = await page.locator('table tbody tr').count();

        // Get the first trace ID row to select
        const firstTraceCheckbox = page.locator('table tbody tr').first().locator('button[role="checkbox"]').first();
        await firstTraceCheckbox.click();

        // The bulk action bar should appear
        const deleteButton = page.getByRole('button', { name: 'Delete' }).filter({ hasText: 'Delete' });
        await expect(deleteButton).toBeVisible();

        // Click delete
        await deleteButton.click();

        // Wait for the delete to finish and the trace to be removed from the DOM
        await page.waitForTimeout(500);

        // Verify the trace is actually removed from the list (Backend state changed + refetched)
        const traceCountAfter = await page.locator('table tbody tr').count();
        expect(traceCountAfter).toBeLessThan(traceCountBefore);

        // Verify the action bar disappeared
        await expect(deleteButton).not.toBeVisible();
    });
});
