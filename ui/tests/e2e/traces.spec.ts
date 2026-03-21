
import { test, expect } from '@playwright/test';
import { seedGlobalState, seedTraffic } from './test-data';

test.describe('Trace Viewer', () => {
    test.beforeEach(async ({ page, request }) => {
        await seedGlobalState(request);
        // seedTraffic invokes the actual debug endpoint which might seed traces too?
        await seedTraffic(request);

        const HEADERS = {
            'Authorization': 'Basic ZTJlLWFkbWluLWNvcmU6cGFzc3dvcmQ=',
            'Content-Type': 'application/json'
        };

        await request.post('/api/v1/debug/traces', { headers: HEADERS });

        await page.goto('/login');
        await page.waitForLoadState('networkidle');
        await page.fill('input[name="username"]', 'e2e-admin-core');
        await page.fill('input[name="password"]', 'password');
        await Promise.all([
          page.waitForURL('/', { timeout: 30000 }),
          page.click('button[type="submit"]', { force: true })
        ]);
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test('should navigate to traces page and view details', async ({ page }) => {
        await page.goto('/traces');

        // Let's just wait for ANY element indicating a trace or "No traces"
        // Wait, if no traces show up, it's because my seed didn't work.
        // We will just expect a trace button to be visible.
        const firstTrace = page.locator('button.flex.flex-col').first();
        await expect(firstTrace).toBeVisible({ timeout: 15000 });
        await firstTrace.click();

        await expect(page.getByText('Trace Details').first()).toBeVisible({ timeout: 10000 });
        await expect(page.locator('.react-json-view').first()).toBeVisible({ timeout: 10000 });
        await page.click('button:has-text("Close")');
    });

    test('should filter traces', async ({ page }) => {
        await page.goto('/traces');

        // The mock generator uses 'orchestrator-task'
        await page.fill('input[placeholder="Search traces..."]', 'orchestrator-task');
        await expect(page.locator('button.flex.flex-col').first()).toBeVisible({ timeout: 15000 });
    });

    test('should replay trace in playground', async ({ page }) => {
        await page.goto('/traces');

        const firstTrace = page.locator('button.flex.flex-col').first();
        await expect(firstTrace).toBeVisible({ timeout: 15000 });
        await firstTrace.click();

        await page.click('button:has-text("Replay")');
        await expect(page).toHaveURL(/.*\/playground.*/);
    });
});
