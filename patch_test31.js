const fs = require('fs');
const path = 'ui/tests/tool-inspector.spec.ts';

const replacement = `
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
    await inspectBtn.evaluate(b => b.click());

    // We just need to verify the dialog opened. In radix UI, the content has role="dialog".
    // Or we can just look for the class "dialog-content".
    // Or just look for "Playground" or "Test Tool" button.
    const runBtn = page.locator('button', { hasText: 'Run Tool' }).first();
    await expect(runBtn).toBeVisible({ timeout: 10000 });
  });
});
`;

fs.writeFileSync(path, replacement);
