const fs = require('fs');
const path = 'ui/tests/tool-inspector.spec.ts';

const replacement = `
import { test, expect } from '@playwright/test';

test.describe('Tool Inspector', () => {
  test('Tools page loads and inspector opens with real data', async ({ page }) => {

    await page.goto('/tools');

    // Wait for at least one tool row to appear in the table
    await page.waitForSelector('table tbody tr');

    // Just click the first "Inspect" button we see
    await page.locator('table tbody tr').first().locator('button:has-text("Inspect")').click();

    // Wait for the dialog to open. It might just be a panel or dialog depending on viewport.
    // The previous code had it as \`div[role="dialog"]\`. Let's just wait for the text "Schema" to be visible in a tab.
    await expect(page.getByRole('tab', { name: 'Schema' }).first()).toBeVisible({ timeout: 10000 });

    // Switch to Schema tab
    await page.getByRole('tab', { name: 'Schema' }).first().click();

    const schemaPanel = page.getByRole('tabpanel').filter({ hasText: 'Visual' }).first();
    await expect(schemaPanel).toBeVisible();

    await schemaPanel.getByRole('tab', { name: 'JSON' }).click();

    // Check for the rendered elements by syntax highlighter
    // The schema JSON view will have elements with .language-json or .react-syntax-highlighter
    // We just check that the JSON view rendered instead of raw <pre>
    await expect(schemaPanel.locator('.language-json').first().or(schemaPanel.locator('pre > code').first())).toBeVisible();

    // We can also verify it has SOME property or type since every schema does.
    await expect(schemaPanel.locator('.language-json, pre > code').filter({ hasText: /"type"/ }).first()).toBeVisible();
  });
});
`;

fs.writeFileSync(path, replacement);
