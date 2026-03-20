import { test, expect } from '@playwright/test';

test.describe('Traces Bulk Export', () => {
    test('should allow selecting multiple traces and exporting them', async ({ page }) => {

        // Because the E2E backend uses a minimal config (which lacks audit config) we will use page.route to mock
        // the network request. E2E environments that don't load full components must rely on routing to mock missing data.
        await page.route('**/api/v1/traces', async (route, request) => {
             if(request.method() === 'GET') {
                 const data = [
                     { id: "trace-1", rootSpan: { name: "test-tool" }, timestamp: new Date().toISOString(), totalDuration: 100, status: "success", trigger: "user" },
                     { id: "trace-2", rootSpan: { name: "test-tool" }, timestamp: new Date().toISOString(), totalDuration: 150, status: "success", trigger: "user" },
                     { id: "trace-3", rootSpan: { name: "test-tool" }, timestamp: new Date().toISOString(), totalDuration: 200, status: "success", trigger: "user" }
                 ];
                 await route.fulfill({ status: 200, json: data });
             } else {
                 await route.continue();
             }
        });

        // Visit the traces page to evaluate in context
        await page.goto('/traces');

        // Ensure the trace list is visible (wait for traces to appear)
        await expect(page.locator('.Virtuoso').or(page.locator('button[role="checkbox"]').first())).toBeVisible({ timeout: 10000 });

        // Ensure there are checkboxes available
        await expect(page.locator('button[role="checkbox"]')).not.toHaveCount(0);

        // Verify the "selected" text doesn't exist yet, or is 0
        await expect(page.getByText('0 selected')).toBeVisible();

        // Wait a tiny bit for the virtuoso list to render
        await page.waitForTimeout(1000);

        // Wait until at least 1 actual trace row is populated
        // The list must populate before "Select all traces" works properly with filtered items
        await expect(page.locator('button[aria-label^="Select trace "]')).not.toHaveCount(0, { timeout: 10000 });

        // Instead of clicking individual trace checkboxes, let's select all via the "Select all traces" checkbox
        // Make sure it's fully interactable
        const selectAllCheckbox = page.getByRole('checkbox', { name: 'Select all traces' });
        await selectAllCheckbox.waitFor({ state: 'visible' });

        await selectAllCheckbox.click({ force: true });

        // Wait for the text to update. Wait up to 5 seconds.
        await expect(page.getByText(/[1-9][0-9]* selected/)).toBeVisible({ timeout: 5000 });

        // Ensure the Export button appears
        const exportButton = page.getByRole('button', { name: 'Export', exact: true });
        await exportButton.waitFor({ state: 'visible' });

        // Setup download listener before clicking
        const downloadPromise = page.waitForEvent('download');

        // Click the Export button
        await exportButton.click();

        // Wait for the download to start
        const download = await downloadPromise;

        // Ensure the filename starts with mcp-traces-bulk-export
        expect(download.suggestedFilename()).toMatch(/^mcp-traces-bulk-export-.*\.json$/);

        // Ensure success toast appears
        await expect(page.getByText('Bulk Export Complete', { exact: true })).toBeVisible();
    });
});
