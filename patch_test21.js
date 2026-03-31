const fs = require('fs');
const path = 'ui/tests/tool-inspector.spec.ts';

const replacement = `
import { test, expect } from '@playwright/test';

test.describe('Tool Inspector', () => {
  test('Tools page loads and inspector opens with real data', async ({ page }) => {

    await page.goto('/tools');

    // Wait for at least one tool row to appear in the table
    await page.waitForSelector('table tbody tr', { state: 'visible', timeout: 15000 });

    // Click "Inspect" button
    const inspectBtn = page.getByRole('button', { name: /Inspect/i }).first();
    await expect(inspectBtn).toBeVisible();
    await inspectBtn.click();

    // The inspector dialog/sheet should open
    // Since "Schema" is just a tab, and "JSON" is a tab...
    // The previous error showed getByText('JSON') failed. This might mean the dialog didn't actually open, or the tab list is not rendering.
    // Let's first wait for the inspector dialog title or some known text.
    // E.g. "Available Tools" is on the page, the sheet will show the tool name.
    const toolName = await page.locator('table tbody tr td').nth(2).innerText();

    // We expect the dialog to have the tool name.
    await expect(page.getByRole('dialog').getByText(toolName).first()).toBeVisible({ timeout: 10000 });

    // Now we look for the "JSON" tab. Note that Radix Tabs might be constructed differently.
    // Let's find the element with exactly "JSON" or "Schema".
    await page.getByText('Schema', { exact: true }).first().click();

    // Click JSON tab
    await page.getByText('JSON', { exact: true }).first().click();

    // Check for the rendered elements by syntax highlighter
    await expect(page.locator('.language-json').first().or(page.locator('pre > code').first())).toBeVisible({ timeout: 5000 });
  });
});
`;

fs.writeFileSync(path, replacement);
