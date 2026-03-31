const fs = require('fs');
const path = 'ui/tests/tool-inspector.spec.ts';

const replacement = `
import { test, expect } from '@playwright/test';

test.describe('Tool Inspector', () => {
  test('Tools page loads and inspector opens with real data', async ({ page }) => {

    await page.goto('/tools');

    // It seems there are no tools at all loaded or visible.
    // The previous test failed waiting for table tbody tr td.
    // If the mock backend doesn't provide ANY tools in config.minimal.yaml, we MUST provide one!
    // But we don't know the exact protobuf names!
    // Wait, the API for tools is /api/v1/tools, let's just use page.route ONLY to mock a valid response to /api/v1/tools?
    // NO, the prompt says: "If you find yourself mocking a network request in the frontend, STOP. Go back and seed the database."
    // Let's seed the database using an "env" variable config, OR find a working config object!
    // Let's read server/config.minimal.yaml to see if it exposes any tools.
    // Usually there's a "system_status" or similar tool. Wait! Maybe we need to hit \`/api/v1/discovery/trigger\` to force discovery of tools!
    await page.request.post('/api/v1/discovery/trigger');
    await page.waitForTimeout(3000); // give it time to discover

    // reload
    await page.goto('/tools');

    // Wait for at least one tool row to appear in the table
    await page.waitForSelector('table tbody tr', { state: 'visible', timeout: 15000 });

    // Click "Inspect" button
    const inspectBtn = page.getByRole('button', { name: /Inspect/i }).first();
    await expect(inspectBtn).toBeVisible();
    await inspectBtn.click();

    const toolName = await page.locator('table tbody tr td').nth(2).innerText();

    // We expect the dialog to have the tool name.
    await expect(page.getByRole('dialog').getByText(toolName).first()).toBeVisible({ timeout: 10000 });

    // Now we look for the "JSON" tab.
    await page.getByText('Schema', { exact: true }).first().click();

    // Click JSON tab
    await page.getByText('JSON', { exact: true }).first().click();

    // Check for the rendered elements by syntax highlighter
    await expect(page.locator('.language-json').first().or(page.locator('pre > code').first())).toBeVisible({ timeout: 5000 });
  });
});
`;

fs.writeFileSync(path, replacement);
