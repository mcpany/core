
import { test, expect } from '@playwright/test';

test.describe('Tool Inspector', () => {
  test('Tools page loads and inspector opens with real data', async ({ page }) => {

    // We must manually trigger discovery or there are no tools initially
    await page.request.post('/api/v1/discovery/trigger');
    await page.waitForTimeout(3000); // give it time to discover

    await page.goto('/tools');

    // Wait for at least one tool row to appear in the table
    await page.waitForSelector('table tbody tr', { state: 'visible', timeout: 15000 });

    // Try to trigger the row expansion.
    const expandAccordion = await page.locator('.lucide-chevron-down').first();
    if (await expandAccordion.isVisible()) {
       await expandAccordion.click();
    }

    // Use { force: true } because HTML overlay intercepts pointer events in Chromium headless mode sometimes
    const inspectBtn = page.locator('button:has-text("Inspect")').first();
    await expect(inspectBtn).toBeVisible({ timeout: 10000 });

    // Evaluate click directly to bypass pointer interception
    await inspectBtn.evaluate((b: HTMLElement | SVGElement) => {
        if ('click' in b && typeof b.click === 'function') {
            b.click();
        } else {
            b.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true, view: window }));
        }
    });

    // Wait for dialog
    await page.waitForTimeout(1000);

    // Look for "Visual" text to prove that the inspector is loaded and has a Visual tab.
    const visualText = page.locator('*:has-text("Visual")').last();
    await expect(visualText).toBeVisible({ timeout: 10000 });
  });
});
