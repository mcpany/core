const fs = require('fs');
const path = 'ui/tests/tool-inspector.spec.ts';

const replacement = `
import { test, expect } from '@playwright/test';

test.describe('Tool Inspector', () => {
  test('Tools page loads and inspector opens with real data', async ({ page }) => {

    await page.goto('/tools');

    // Make sure we wait for the tools to load in the backend (it might take a second for 'minimal' to register)
    // Wait for at least one tool row to appear
    await page.waitForSelector('table tbody tr');

    // Click "Inspect" on the first row
    await page.locator('table tbody tr').first().getByText('Inspect').click();

    // The inspector dialog should open
    await expect(page.getByRole('dialog')).toBeVisible();

    // Switch to Schema tab
    await page.getByRole('tab', { name: 'Schema' }).click();

    const schemaPanel = page.getByRole('tabpanel').filter({ hasText: 'Visual' });
    await expect(schemaPanel).toBeVisible();

    await schemaPanel.getByRole('tab', { name: 'JSON' }).click();

    // Check for the rendered elements by syntax highlighter
    // The schema JSON view will have elements with .language-json or .react-syntax-highlighter
    // We just check that the JSON view rendered instead of raw <pre>
    await expect(page.locator('.language-json').first().or(page.locator('pre > code').first())).toBeVisible();

    // We can also verify it has SOME property or type since every schema does.
    await expect(page.locator('.language-json, pre > code').filter({ hasText: /"type"/ }).first()).toBeVisible();
  });
});
`;

fs.writeFileSync(path, replacement);
