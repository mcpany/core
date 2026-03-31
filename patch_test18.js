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

    // Click "Inspect" button. The button has text "Inspect".
    const inspectBtn = page.getByRole('button', { name: /Inspect/i }).first();
    await expect(inspectBtn).toBeVisible();
    await inspectBtn.click();

    // The inspector dialog/sheet should open. Wait for any element with the text "Schema" which is a tab.
    // The tab is typically a role="tab" with text "Schema" or similar.
    const schemaTab = page.getByText('Schema').first();
    await expect(schemaTab).toBeVisible({ timeout: 10000 });

    // Switch to Schema tab
    await schemaTab.click();

    // There should be a Visual tab
    const visualTab = page.getByText('Visual').first();
    await expect(visualTab).toBeVisible();

    // Click JSON tab
    const jsonTab = page.getByText('JSON', { exact: true }).first();
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
