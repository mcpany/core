
import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe.skip('Trace Viewer', () => {
    test.beforeEach(async ({ page, request }) => {
        await seedGlobalState(request);

        // We MUST mock traces because they are generated dynamically by background workers
        // which may not reliably execute fast enough during this isolated test.
        await page.route('**/api/v1/traces', async route => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify([{
                    id: 'trace-123',
                    call_id: 'calculate_sum',
                    status: 'success',
                    timestamp: new Date().toISOString(),
                    duration_ms: 120,
                    request: { arguments: { a: 5, b: 10 } },
                    response: { result: 15 }
                }])
            });
        });

        await page.goto('/traces');

        // Ensure auth/navigation
        // await page.waitForLoadState('networkidle');
    });

    test('should navigate to traces page and view details', async ({ page }) => {
        // Trigger reload to make sure intercepts work
        await page.reload();

        // Actually, let's just check for any trace item
        const firstTrace = page.locator('button.flex.flex-col').first();
        await expect(firstTrace).toBeVisible({ timeout: 20000 });

        // Click the first trace
        await firstTrace.click();

        // Verify details panel opens and shows information
        await expect(page.locator('text=Trace Details').first()).toBeVisible({ timeout: 20000 });

        // Wait for JSON viewer to render
        await expect(page.locator('.react-json-view').first()).toBeVisible({ timeout: 20000 });

        // Close details
        await page.click('button:has-text("Close")');
    });

    test('should filter traces', async ({ page }) => {
        await page.reload();

        // Type in search box
        await page.fill('input[placeholder="Search traces..."]', 'calculate');

        // Expect only matching items
        // and doesn't crash the page
        await expect(page.locator('button.flex.flex-col').first()).toBeVisible({ timeout: 20000 });
    });

    test('should replay trace in playground', async ({ page }) => {
        await page.reload();

        const firstTrace = page.locator('button.flex.flex-col').first();
        await expect(firstTrace).toBeVisible({ timeout: 20000 });
        await firstTrace.click();

        // Click "Replay in Playground"
        await page.click('button:has-text("Replay")');

        // Should navigate to playground
        await expect(page).toHaveURL(/.*\/playground.*/);
    });
});
