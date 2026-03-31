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

    // The inspector dialog/sheet has role "dialog" in Radix UI
    const dialog = page.locator('[role="dialog"]').first();
    await expect(dialog).toBeVisible();

    // Switch to Schema tab
    await dialog.getByRole('tab', { name: 'Schema' }).click();

    // The visual schema panel
    const schemaPanel = dialog.getByRole('tabpanel').filter({ hasText: 'Visual' });
    await expect(schemaPanel).toBeVisible();

    // Click JSON tab
    await schemaPanel.getByRole('tab', { name: 'JSON' }).click();

    // Verify JSON View instead of <pre>
    await expect(schemaPanel.locator('.language-json').first().or(schemaPanel.locator('pre > code').first())).toBeVisible();
  });
});
`;

fs.writeFileSync(path, replacement);
