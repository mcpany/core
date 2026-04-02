import { test, expect } from '@playwright/test';

test.describe('Property Inspector', () => {
    test('verifies interactive JsonTree renders correctly for complex data', async ({ page }) => {
        // 1. Navigate to Inspector
        await page.goto('/inspector');

        // 2. Click Seed Trace
        const seedButton = page.getByRole('button', { name: /Seed Trace/i });
        await seedButton.waitFor({ state: 'visible', timeout: 60000 });
        await seedButton.click();

        // Wait for the trace to appear in the table. The mocked trace has "code-refactor".
        const traceRow = page.getByRole('row').filter({ hasText: 'code-refactor' }).first();
        await expect(traceRow).toBeVisible({ timeout: 60000 });

        // 3. Open the trace details
        await traceRow.click();

        // 4. Wait for the detail sheet to open
        const sheet = page.getByRole('dialog');
        await expect(sheet).toBeVisible();

        // 5. Navigate to Payload tab
        await sheet.getByRole('tab', { name: /Payload/i }).click();

        // 6. Verify the JsonTree (Property Inspector) is rendered instead of a raw SyntaxHighlighter
        // We look for the "Apple Design Standard" specific classes or texts we added
        const jsonTreeContainer = sheet.locator('.bg-muted\\/10.backdrop-blur-sm').first();
        await expect(jsonTreeContainer).toBeVisible();

        // 7. Expand the root object if not expanded
        // Click the first chevron in the response payload JsonTree to expand
        // It should contain the "metadata" key that we added in our backend seed
        const responsePayloadTree = sheet.locator('.bg-muted\\/10.backdrop-blur-sm').nth(1);
        await expect(responsePayloadTree).toBeVisible();

        // Verify the "metadata" key exists (we need to ensure it's expanded or we can just see it)
        await expect(responsePayloadTree.getByText('metadata', { exact: true })).toBeVisible();

        // Verify Type Badges exist (e.g. string, number, boolean)
        await expect(responsePayloadTree.getByText('string', { exact: true }).first()).toBeVisible();
        await expect(responsePayloadTree.getByText('number', { exact: true }).first()).toBeVisible();
    });
});