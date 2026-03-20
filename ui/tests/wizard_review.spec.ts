import { test, expect } from '@playwright/test';

test.describe('Marketplace Wizard Review Step', () => {
  test('should format config as YAML in the Review step instead of raw JSON', async ({ page }) => {
    // Navigate to marketplace to trigger the wizard
    await page.goto('/marketplace');

    // Wait for page to load
    await expect(page.getByRole('heading', { name: 'Marketplace' })).toBeVisible();

    // Click "Create Config" button to open the wizard
    await page.getByRole('button', { name: 'Create Config' }).click();

    // The wizard should now be open
    await expect(page.getByRole('dialog')).toBeVisible();

    // Step 1: Service Type
    // Fill in required fields for Service Type (Assuming name is required, maybe others)
    // We just need to navigate to the Review step. We can look for the "Next" button.
    // Let's first fill in the service name.
    await page.getByPlaceholder('e.g. My Postgres DB').fill('Test Service Format');
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // Step 2: Parameters (or OpenAPI if we chose that, but default is usually HTTP or CMD)
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // Step 3: Webhooks
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // Step 4: Auth
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // Step 5: Review
    await expect(page.getByText('Configuration Ready')).toBeVisible();

    // Verify it's YAML (it shouldn't have raw JSON characters like {"name":"Test Service Format"} on one line)
    // We expect the pre block to contain formatted YAML, for example: `name: Test Service Format`
    const preBlock = page.locator('pre');
    await expect(preBlock).toContainText('name: Test Service Format');

    // Check that it does not contain raw JSON format for the name key
    const text = await preBlock.textContent();
    expect(text).not.toContain('"name": "Test Service Format"');

    // Skip backend verification in this test as we are mocking/not running backend.
    // The visual UI is what we test here.
  });
});
