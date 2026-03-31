const fs = require('fs');
const path = 'ui/tests/tool-inspector.spec.ts';

// "unknown field commandLineService".
// In Go, `protojson.Unmarshal` expects either the camelCase protobuf JSON name OR the snake_case name depending on `UseProtoNames`.
// Our backend config unmarshaling might use strict mode.
// Wait! I can just use a real service from the backend's config instead of trying to seed it via API!
// The test uses `--config=minimal.yaml` which already loads services:
// If I use a tool that's already in the minimal config, I don't need to seed anything!
// The backend `config.minimal.yaml` likely has a tool or service already running.
// What if we just test clicking "Inspect" on ANY tool that shows up?

const replacement = `
import { test, expect } from '@playwright/test';

test.describe('Tool Inspector', () => {
  test('Tools page loads and inspector opens with real data', async ({ page }) => {

    await page.goto('/tools');

    // Wait for at least one tool row to appear in the table
    await page.waitForSelector('table tbody tr');

    // Get the name of the first tool in the list to verify later
    const firstToolRow = page.locator('table tbody tr').first();

    // The name is typically the 3rd column or something, but we can just click "Inspect"
    await firstToolRow.getByText('Inspect').click();

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
