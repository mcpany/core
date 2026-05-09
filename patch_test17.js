const fs = require('fs');
const path = 'ui/tests/tool-inspector.spec.ts';

const replacement = `
import { test, expect } from '@playwright/test';

test.describe('Tool Inspector', () => {
  test('Tools page loads and inspector opens with real data', async ({ page }) => {

    await page.goto('/tools');

    // Wait for at least one tool row to appear in the table
    // Let's make sure the table exists and has rows.
    await page.waitForSelector('table tbody tr', { state: 'visible', timeout: 15000 });

    // The name is typically the 3rd column. We just click the "Inspect" button in the first row.
    // Sometimes the backend takes time, so we wait for an Inspect button.
    const inspectBtn = page.locator('table tbody tr').first().locator('button:has-text("Inspect")');
    await expect(inspectBtn).toBeVisible();
    await inspectBtn.click();

    // The inspector dialog/sheet should open. Wait for any element with the text "Schema" which is a tab.
    const schemaTab = page.locator('button[role="tab"]', { hasText: 'Schema' }).first();
    await expect(schemaTab).toBeVisible({ timeout: 10000 });

    // Switch to Schema tab
    await schemaTab.click();

    // There should be a Visual tab
    const visualTab = page.locator('button[role="tab"]', { hasText: 'Visual' }).first();
    await expect(visualTab).toBeVisible();

    // Click JSON tab
    const jsonTab = page.locator('button[role="tab"]', { hasText: 'JSON' }).first();
    await expect(jsonTab).toBeVisible();
    await jsonTab.click();

    // Check for the rendered elements by syntax highlighter
    // The schema JSON view will have elements with .language-json or .react-syntax-highlighter
    // We just check that the JSON view rendered instead of raw <pre>
    await expect(page.locator('.language-json').first().or(page.locator('pre > code').first())).toBeVisible({ timeout: 5000 });
  });
});
`;

fs.writeFileSync(path, replacement);
