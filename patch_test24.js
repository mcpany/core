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

    // Click "Inspect" button
    const inspectBtn = page.getByRole('button', { name: /Inspect/i }).first();
    await expect(inspectBtn).toBeVisible();
    await inspectBtn.click();

    // Now we look for the "Schema" tab.
    // Sometimes it's inside a dialog, and text might have trailing spaces or different casing.
    const schemaTab = page.locator('button[role="tab"]').filter({ hasText: /Schema/i }).first();
    await expect(schemaTab).toBeVisible({ timeout: 10000 });
    await schemaTab.click();

    // Click JSON tab
    const jsonTab = page.locator('button[role="tab"]').filter({ hasText: /JSON/i }).first();
    await expect(jsonTab).toBeVisible({ timeout: 10000 });
    await jsonTab.click();

    // Check for the rendered elements by syntax highlighter
    await expect(page.locator('.language-json').first().or(page.locator('pre > code').first())).toBeVisible({ timeout: 5000 });
  });
});
`;

fs.writeFileSync(path, replacement);
