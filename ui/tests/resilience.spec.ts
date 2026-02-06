import { test, expect } from '@playwright/test';

test.describe('Service Resilience & Bulk Management', () => {
  test.setTimeout(60000);
  const timestamp = Date.now();
  const serviceName = `resilience-test-${timestamp}`;

  test.beforeEach(async ({ page }) => {
    // Navigate to services page
    await page.goto('/upstream-services');
    await page.waitForLoadState('networkidle');
  });

  test('can configure resilience settings for a service', async ({ page }) => {
    // 1. Create new service
    await page.getByRole('button', { name: 'Add Service' }).click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 10000 });

    // Select Template
    await page.getByText('Custom Service').click();

    // Fill General
    await page.getByLabel('Service Name').fill(serviceName);
    await page.getByLabel('Version').fill('1.0.0');

    // Fill Connection (HTTP default is fine, just set URL)
    await page.getByRole('tab', { name: 'Connection' }).click();
    await page.getByLabel('Base URL').fill('https://httpbin.org');

    // Fill Resilience
    await page.getByRole('tab', { name: 'Advanced' }).click();

    // Check if Resilience Editor is present
    await expect(page.getByText('Timeouts')).toBeVisible();

    // Set Timeout
    await page.getByLabel('Request Timeout').fill('5s');

    // Enable Retry Policy (Assuming switch is unchecked by default)
    // The switch is inside the card header for "Retry Policy"
    // Finding the switch associated with "Retry Policy"
    const retryCard = page.locator('.rounded-xl', { hasText: 'Retry Policy' }).first();
    // Assuming structure: CardHeader > div(Title/Desc) + Switch
    // Shadcn Switch usually has role="switch"
    // We target the switch inside the card that contains "Retry Policy" text
    // Playwright locator chaining:
    // Locator for the card, then locator for the switch
    // Actually, let's try to be specific.
    await page.locator('.space-y-1:has-text("Retry Policy")').locator('xpath=..').locator('button[role="switch"]').click();

    await page.getByLabel('Max Retries').fill('3');
    await page.getByLabel('Base Backoff').fill('1s');

    // Save
    await page.getByRole('button', { name: 'Save Changes' }).click();
    await expect(page.getByText('New service registered successfully').first()).toBeVisible();

    // 2. Verify Persistence (Re-open)
    // Refresh page to ensure fresh data fetch
    await page.reload();

    // Find the row with serviceName
    const row = page.getByRole('row', { name: serviceName });
    await expect(row).toBeVisible();

    // Edit
    await row.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();

    // Check values
    await page.getByRole('tab', { name: 'Advanced' }).click();
    await expect(page.getByLabel('Request Timeout')).toHaveValue('5s');
    await expect(page.getByLabel('Max Retries')).toHaveValue('3');
  });

  test('can bulk edit resilience settings', async ({ page }) => {
    // Pre-requisite: Create 2 services
    const s1 = `bulk-1-${timestamp}`;
    const s2 = `bulk-2-${timestamp}`;

    // Helper to create simple service
    const createService = async (name: string) => {
        await page.getByRole('button', { name: 'Add Service' }).click();
        await page.getByText('Custom Service').click();
        await page.getByLabel('Service Name').fill(name);
        await page.getByRole('tab', { name: 'Connection' }).click();
        await page.getByLabel('Base URL').fill('https://example.com');
        await page.getByRole('button', { name: 'Save Changes' }).click();
        await expect(page.getByText('New service registered successfully').first()).toBeVisible();
        // Close sheet if it didn't close (it should)
        // Wait for list refresh
        await page.waitForTimeout(500);
    };

    await createService(s1);
    await createService(s2);

    // Reload to ensure list is stable
    await page.reload();

    // Select both
    await page.getByLabel(`Select ${s1}`).check();
    await page.getByLabel(`Select ${s2}`).check();

    // Click Bulk Edit
    await page.getByRole('button', { name: 'Bulk Edit' }).click();

    // Set Timeout
    await page.getByLabel('Timeout (optional)').fill('10s');
    await page.getByLabel('Max Retries (optional)').fill('5');

    // Apply
    await page.getByRole('button', { name: 'Apply Changes' }).click();
    await expect(page.getByText('2 services have been updated')).toBeVisible();

    // Verify s1
    await page.getByRole('row', { name: s1 }).getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();
    await page.getByRole('tab', { name: 'Advanced' }).click();
    await expect(page.getByLabel('Request Timeout')).toHaveValue('10s');

    // Check Retry Policy (should be enabled implicitly by setting Max Retries)
    await expect(page.getByLabel('Max Retries')).toHaveValue('5');
  });
});
